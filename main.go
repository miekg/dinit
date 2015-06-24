// Dinit is a mini init replacement useful for use inside Docker containers.
package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var (
	verbose     bool
	test        bool
	timeout     time.Duration
	maxproc     float64
	start, stop string

	cmds = NewCommands()
)

const testPid = 123

func main() {
	flag.BoolVar(&verbose, "verbose", envBool("DINIT_VERBOSE", false), "be more verbose and show stdout/stderr of commands (DINIT_VERBOSE)")
	flag.DurationVar(&timeout, "timeout", envDuration("DINIT_TIMEOUT", 10*time.Second), "time in seconds between SIGTERM and SIGKILL (DINIT_TIMEOUT)")
	flag.Float64Var(&maxproc, "maxproc", 0.0, "set GOMAXPROCS to runtime.NumCPU() * maxproc, when GOMAXPROCS already set use that")
	flag.StringVar(&start, "start", "", "command to run during startup, non-zero exit status abort dinit")
	flag.StringVar(&stop, "stop", "", "command to run during teardown")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: dinit [OPTION]... CMD [CMD]...")
		fmt.Fprintln(os.Stderr, "Start CMDs by passing the environment and reap any zombies.")
		fmt.Fprintln(os.Stderr, "Distribute SIGHUP, SIGTERM and SIGINT to CMDs.\n")
		flag.PrintDefaults()
	}
	flag.Parse()

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

	run(flag.Args())
	wait()
}

// run runs the commands as given on the commandline.
func run(args []string) {
	for _, arg := range args {
		cmd := command(arg)
		if err := cmd.Start(); err != nil {
			logFatalf("%s", err)
		}

		if test {
			logPrintf("pid %d started: %v", testPid, cmd.Args)

		} else {
			logPrintf("pid %d started: %v", cmd.Process.Pid, cmd.Args)
		}

		cmds.Insert(cmd)

		go func() {
			err := cmd.Wait()
			if test {
				logPrintf("pid %d, finished: %v with error: %v", testPid, cmd.Args, err)
			} else {
				logPrintf("pid %d, finished: %v with error: %v", cmd.Process.Pid, cmd.Args, err)
			}
			cmds.Remove(cmd)
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
			cmds.Signal(sig)

			time.Sleep(2 * time.Second)

			if cmds.Len() > 0 {
				logPrintf("%d processes still alive after SIGINT/SIGTERM", cmds.Len())
				time.Sleep(timeout)
			}
			cmds.Signal(syscall.SIGKILL)
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
