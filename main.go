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
	"strings"
	"syscall"
	"time"
)

var (
	Version = "0.6.1" // Remove and just use git hash when building it

	timeout       time.Duration
	maxproc       float64
	start, stop   string
	primary, sock bool
	version       bool

	test bool // only used then testing

	procs = NewProcs()
	prim  = NewPrimary()
)

const testPid = 123

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: dinit (version %s) [OPTION...] -r CMD [OPTION..] [-r CMD [OPTION...]]...\n", Version)

		fmt.Fprintln(os.Stderr, "Start CMDs by passing the environment.")
		fmt.Fprintln(os.Stderr, "Distribute SIGHUP, SIGTERM and SIGINT to the processes.\n")
		flag.PrintDefaults()
	}

	// -r CMD [OPTION...]
	commands := Args(os.Args)

	flag.DurationVar(&timeout, "timeout", envDuration("$DINIT_TIMEOUT", 10*time.Second), "time in seconds between SIGTERM and SIGKILL (DINIT_TIMEOUT)")
	flag.Float64Var(&maxproc, "maxproc", 0.0, "set GOMAXPROCS to runtime.NumCPU() * maxproc, when GOMAXPROCS already set use that")
	flag.Float64Var(&maxproc, "core-fraction", 0.0, "set GOMAXPROCS to runtime.NumCPU() * core-fraction, when GOMAXPROCS already set use that")
	flag.StringVar(&start, "start", envString("$DINIT_START", ""), "command to run during startup, non-zero exit status aborts dinit (DINIT_START)")
	flag.StringVar(&stop, "stop", envString("$DINIT_STOP", ""), "command to run during teardown (DINIT_STOP)")
	flag.BoolVar(&sock, "submit", false, "write -r CMD... to the unix socket "+socketName)
	flag.BoolVar(&primary, "primary", false, "all processes are primary")

	if len(commands) == 0 {
		flag.Usage()
		return
	}
	flag.Parse()

	if !sock {
		// If sending something over the socket, don't open it.
		go socket()
		defer os.Remove(socketName)
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
	if sock {
		err := write(commands)
		if err != nil {
			logFatalf("failed to write to unix socket: %s", err)
		}
		return
	}
	if os.Getpid() == 1 {
		go reap()
	}
	run(commands, false)
	wait()
}

// run runs the commands as given on the command line. If noprimary is
// true none of the processes will be considered primary.
func run(commands []*exec.Cmd, fromsocket bool) {
	for i, _ := range commands {
		// Need to copy here, because otherwise the closure below will access
		// to wrong command, when we run in a loop.
		c := commands[i]
		if err := c.Start(); err != nil {
			logPrintf("process failed to start: %v", err)
			if !fromsocket {
				procs.Cleanup(syscall.SIGINT)
				return
			}
			continue
		}

		if i == len(commands)-1 && !fromsocket {
			prim.Set(c.Process.Pid)
		}

		pid := c.Process.Pid
		if test {
			pid = testPid
		}
		logPrintf("pid %d started: %v", pid, c.Args)

		procs.Insert(c)

		go func() {
			err := c.Wait()
			pid := c.Process.Pid
			if test {
				pid = testPid
			}

			switch err {
			default:
				_, ok := err.(*os.SyscallError)
				if !ok {
					logPrintf("pid %d finished: %v with error: %v", pid, c.Args, err)
					break
				}
				fallthrough
			case nil:
				logPrintf("pid %d finished: %v", pid, c.Args)

			}

			procs.Remove(c)
			if primary || prim.Primary(c.Process.Pid) {
				if primary {
					logPrintf("all processes considered primary, signalling other processes")
				} else {
					logPrintf("pid %d was primary, signalling other processes", pid)
				}
				procs.Cleanup(syscall.SIGINT)
			}
		}()
	}
	return
}

// wait waits for commands to finish.
func wait() {
	defer func() { logPrintf("all processes exited, goodbye!") }()
	if !sock {
		defer os.Remove(socketName)
	}

	ints := make(chan os.Signal)
	signal.Notify(ints, syscall.SIGINT, syscall.SIGTERM)

	other := make(chan os.Signal)
	signal.Notify(other, syscall.SIGHUP)

	tick := time.Tick(100 * time.Millisecond) // 0.1 sec

	for {
		select {
		case <-tick:
			if procs.Len() == 0 {
				return
			}
		case sig := <-other:
			procs.Signal(sig)
		case sig := <-ints:
			procs.Cleanup(sig)
		}
	}
}

// command parses arg and returns an *exec.Cmd that is ready to be run.
// This is currently only used for the start and stop commands.
func command(arg string) *exec.Cmd {
	args := strings.Fields(arg) // Split on spaces and execute.
	for i, a := range args {
		args[i] = os.ExpandEnv(a)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
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
