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
	timeout     time.Duration
	maxproc     float64
	start, stop string

	cmds = NewCommands()
)

func main() {
	flag.BoolVar(&verbose, "verbose", envBool("DINIT_VERBOSE", false), "be more verbose and show stdout/stderr of commands (DINIT_VERBOSE)")
	flag.DurationVar(&timeout, "timeout", envDuration("DINIT_TIMEOUT", 10*time.Second), "time in seconds between SIGTERM and SIGKILL (DINIT_TIMEOUT)")
	flag.Float64Var(&maxproc, "maxproc", 0.0, "set GOMAXPROC to runtime.NumCPU() * maxproc, when 0.0 use GOMAXPROCS")
	flag.StringVar(&start, "start", "", "command to run during startup, non-zero exit status abort dinit")
	flag.StringVar(&stop, "stop", "", "command to run during teardown")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: dinit [OPTION]... CMD [CMD]...")
		fmt.Fprintln(os.Stderr, "Start CMDs by passing the environment and reap any zombies.\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		logFatalf("need at least one command")
	}

	if maxproc > 0.0 {
		numcpu := strconv.Itoa(int(math.Ceil(float64(runtime.NumCPU()) * maxproc)))
		logF("using %s as GOMAXPROCS", numcpu)
		os.Setenv("GOMAXPROCS", numcpu)
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

		logF("pid %d started: %v", cmd.Process.Pid, cmd.Args)

		cmds.Insert(cmd)

		go func() {
			err := cmd.Wait()
			if err != nil {
				logF("pid %d, finished: %v with error: %s", cmd.Process.Pid, cmd.Args, err)
			} else {
				logF("pid %d, finished: %v", cmd.Process.Pid, cmd.Args)
			}
			cmds.Remove(cmd)
		}()
	}
	return
}

// wait waits for commands to finish.
func wait() {

	ints := make(chan os.Signal)
	signal.Notify(ints, syscall.SIGINT, syscall.SIGTERM)
	tick := time.Tick(100 * time.Millisecond) // 0.1 sec

	for {
		select {
		case <-tick:
			if cmds.Len() == 0 {
				return
			}
		case sig := <-ints:
			cmds.Signal(sig)

			time.Sleep(2 * time.Second)

			if cmds.Len() > 0 {
				logF("%d processes still alive after SIGINT", cmds.Len())
				time.Sleep(timeout)
			}
			cmds.Signal(syscall.SIGKILL)
		}
	}
}

func logF(format string, v ...interface{}) {
	if !verbose {
		return
	}
	log.Printf("dinit: "+format, v...)
}

func logFatalf(format string, v ...interface{}) {
	log.Fatalf("dinit: "+format, v...)
}
