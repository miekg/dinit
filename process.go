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
	"strings"
	"syscall"
	"time"
)

var (
	verbose bool
	timeout time.Duration
	maxproc float64
)

func main() {
	flag.BoolVar(&verbose, "verbose", envBool("DINIT_VERBOSE", false), "be more verbose and show stdout/stderr of programs (DINIT_VERBOSE)")
	flag.DurationVar(&timeout, "timeout", envDuration("DINIT_TIMEOUT", 10*time.Second), "time in seconds between SIGTERM and SIGKILL (DINIT_TIMEOUT)")
	flag.Float64Var(&maxproc, "maxproc", 0.0, "set GOMAXPROC to os.NumCPU * maxproc, when 0.0 use GOMAXPROCS")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: dinit [OPTION]... PROGRAM [PROGRAM]...")
		fmt.Fprintln(os.Stderr, "Start PROGRAMs by passing the environment and reap any zombies.\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatal("dinit: need at least one program")
	}

	if maxproc > 0.0 {
		numcpu := strconv.Itoa(math.Ceil(os.NumCPU() * maxproc))
		log.Printf("dinit: using %d as GOMAXPROCS", numcpu)
		os.Setenv("GOMAXPROCS", numcpu)
	}

	cmds := []*exec.Cmd{}
	done := make(chan bool)

	for _, arg := range flag.Args() {
		args := strings.Fields(arg) // Split on spaces and execute.
		cmd := exec.Command(args[0], args[1:]...)
		cmds = append(cmds, cmd)

		go func() {
			if verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}

			err := cmd.Start()
			if err != nil {
				log.Fatal(err)
			}

			logf("dinit: pid %d started: %v", cmd.Process.Pid, cmd.Args)

			err = cmd.Wait()
			if err != nil {
				logf("dinit: pid %d, finished with error: %s", cmd.Process.Pid, err)
			} else {
				logf("dinit: pid %d, finished: %v", cmd.Process.Pid, cmd.Args)
			}
			done <- true
		}()
	}

	ints := make(chan os.Signal)
	chld := make(chan os.Signal)
	signal.Notify(ints, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(chld, syscall.SIGCHLD)

	i := 0
	defer reaper()
Wait:
	for {
		select {
		case <-chld:
			go reaper()
		case <-done:
			i++
			if len(cmds) == i {
				break Wait
			}
		case sig := <-ints:
			// There is a race here, because the process could have died, we don't care.
			for _, cmd := range cmds {
				logf("dinit: signal %d sent to pid %d", sig, cmd.Process.Pid)
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
				logf("dinit: SIGKILL sent to pid %d", p.Pid)
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
		logf("dinit: pid %d reaped", pid)
	}
}

func logf(format string, v ...interface{}) {
	if !verbose {
		return
	}
	log.Printf("dinit: "+format, v...)
}
