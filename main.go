// Dinit is a mini init replacement useful for use inside Docker containers.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var (
	timeout     time.Duration
	maxproc     float64
	start, stop string

	test bool // only used then testing

	cmds = NewCommands()
	prim = NewPrimary()
)

const testPid = 123

func main() {
	// -r CMD [OPTION...] is done first and we cleanup what we parse
	var cmd *exec.Cmd = nil
	cmds := []*exec.Cmd{}
	minr := false

	for i := 0; i < len(os.Args); i++ {
		if os.Args[i] == "-r" {
			println("new")
			if cmd != nil {
				cmds = append(cmds, cmd)
			}

			minr = true

			cmd = new(exec.Cmd)
			cmd.Args = append(cmd.Args, os.Args[i+1])
			cmd.Path = os.Args[i+1]
			// TODO(miek): fix overflow here

			os.Args[i] = ""
			os.Args[i+1] = ""

			i++
			continue
		}
		if minr {
			cmd.Args = append(cmd.Args, os.Args[i])
			os.Args[i] = ""
		}
	}
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	fmt.Printf("%v\n", os.Args)
	for _, c := range cmds {
		fmt.Printf("%+v\n", c)
	}

	flag.DurationVar(&timeout, "timeout", envDuration("DINIT_TIMEOUT", 10*time.Second), "time in seconds between SIGTERM and SIGKILL (DINIT_TIMEOUT)")
	flag.Float64Var(&maxproc, "maxproc", 0.0, "set GOMAXPROCS to runtime.NumCPU() * maxproc, when GOMAXPROCS already set use that")
	flag.Float64Var(&maxproc, "core-fraction", 0.0, "set GOMAXPROCS to runtime.NumCPU() * core-fraction, when GOMAXPROCS already set use that")
	flag.StringVar(&start, "start", "", "command to run during startup, non-zero exit status abort dinit")
	flag.StringVar(&stop, "stop", "", "command to run during teardown")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: dinit [OPTION]... CMD [CMD]...") // TODO(miek): fix
		fmt.Fprintln(os.Stderr, "Start CMDs by passing the environment.")
		fmt.Fprintln(os.Stderr, "Distribute SIGHUP, SIGTERM and SIGINT to the processes.\n")
		flag.PrintDefaults()
	}

	if len(flag.Args()) == 0 {
		logFatalf("need at least one command")
	}

	if maxproc > 0.0 {
		if v := os.Getenv("GOMAXPROCS"); v != "" {
			logPrintf("GOMAXPROCS already set, using that value: %s", v)
		} else {
			numcpu := strconv.Itoa(int(math.Ceil(float64(runtime.NumCPU()) * maxproc)))
			logPrintf("using %s as GOMAXPROCS", numcpu)
			os.Setenv("GOMAXPROCS", numcpu)
		}
	}

	if start != "" {
		startcmd := command(start)
		if err := startcmd.Run(); err != nil {
			logFatalf("start command failed: %s", err)
		}
	}
	if stop != "" {
		stopcmd := command(stop)
		defer stopcmd.Run()
	}

	// everyhing is
	run(flag.Args())
	wait()
}

// run runs the commands as given on the command line.
func run(args []string) {
	for i, arg := range args {
		cmd := command(arg)
		if err := cmd.Start(); err != nil {
			logPrintf("%s", err)
			cmds.Cleanup(syscall.SIGINT)
			return
		}
		// TODO(miek): primary is going to be last
		if i == 0 {
			prim.Set(cmd.Process.Pid)
		}

		pid := cmd.Process.Pid
		if test {
			pid = testPid
		}
		logPrintf("pid %d started: %v", pid, cmd.Args)

		cmds.Insert(cmd)

		go func() {
			err := cmd.Wait()
			pid := cmd.Process.Pid
			if test {
				pid = testPid
			}
			logPrintf("pid %d, finished: %v with error: %v", pid, cmd.Args, err)
			cmds.Remove(cmd)
			if prim.Primary(cmd.Process.Pid) && cmds.Len() > 0 {
				logPrintf("pid %d was primary, signalling other processes", pid)
				cmds.Cleanup(syscall.SIGINT)
			}
		}()
	}
	return
}

// wait waits for commands to finish.
func wait() {
	defer func() { logPrintf("all processes exited, goodbye!") }()

	ints := make(chan os.Signal)
	signal.Notify(ints, syscall.SIGINT, syscall.SIGTERM)

	other := make(chan os.Signal)
	signal.Notify(other, syscall.SIGHUP)

	tick := time.Tick(100 * time.Millisecond) // 0.1 sec

	for {
		select {
		case <-tick:
			if cmds.Len() == 0 {
				return
			}
		case sig := <-other:
			cmds.Signal(sig)
		case sig := <-ints:
			cmds.Cleanup(sig)
		}
	}
}

func logPrintf(format string, v ...interface{}) {
	if test {
		fmt.Printf("dinit: "+format+"\n", v...)
		return
	}
	log.Printf("dinit: "+format, v...)
}

func logFatalf(format string, v ...interface{}) {
	log.Fatalf("dinit: "+format, v...)
}
