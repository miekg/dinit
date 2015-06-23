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
	verbose     bool
	timeout     time.Duration
	maxproc     float64
	start, stop string
)

func main() {
	flag.BoolVar(&verbose, "verbose", envBool("DINIT_VERBOSE", false), "be more verbose and show stdout/stderr of commands (DINIT_VERBOSE)")
	flag.DurationVar(&timeout, "timeout", envDuration("DINIT_TIMEOUT", 10*time.Second), "time in seconds between SIGTERM and SIGKILL (DINIT_TIMEOUT)")
	flag.Float64Var(&maxproc, "maxproc", 0.0, "set GOMAXPROC to os.NumCPU * maxproc, when 0.0 use GOMAXPROCS")
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
		logF("using %d as GOMAXPROCS", numcpu)
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

	done := make(chan bool)
	cmds := run(flag.Args(), done)

	defer reaper()

	wait(done, cmds)
}

func run(args []string, done chan bool) []*exec.Cmd {
	cmds := []*exec.Cmd{}
	for _, arg := range args {
		cmd := command(arg)
		cmds = append(cmds, cmd)

		go func() {
			if err := cmd.Start(); err != nil {
				logFatalf("%s", err)
			}

			logF("pid %d started: %v", cmd.Process.Pid, cmd.Args)

			err := cmd.Wait()
			if err != nil {
				logF("pid %d, finished with error: %s", cmd.Process.Pid, err)
			} else {
				logF("pid %d, finished: %v", cmd.Process.Pid, cmd.Args)
			}
			done <- true
		}()
	}
	return cmds
}

// wait waits for commands to finish.
func wait(done chan bool, cmds []*exec.Cmd) {
	i := 0

	ints := make(chan os.Signal)
	chld := make(chan os.Signal)
	signal.Notify(ints, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(chld, syscall.SIGCHLD)

	for {
		select {
		case <-chld:
			go reaper()
		case <-done:
			i++
			if len(cmds) == i {
				return
			}
		case sig := <-ints:
			// There is a race here, because the process could have died, we don't care.
			for _, cmd := range cmds {
				logF("signal %d sent to pid %d", sig, cmd.Process.Pid)
				cmd.Process.Signal(sig)
			}

			time.Sleep(timeout)

			kill := []*os.Process{}
			for _, cmd := range cmds {
				if p, err := os.FindProcess(cmd.Process.Pid); err != nil {
					kill = append(kill, p)
				}
			}
			for _, p := range kill {
				logF("SIGKILL sent to pid %d", p.Pid)
				p.Signal(syscall.SIGKILL)
			}
		}
	}
}

func reaper() {
	for {
		var wstatus syscall.WaitStatus
		pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
		if err != nil {
			return
		}
		logF("pid %d reaped", pid)
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
