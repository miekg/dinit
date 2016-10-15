// Dinit is a mini init replacement useful for use inside Docker containers.
package main

import (
	"flag"
	"fmt"
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
	Version = "0.6.2" // Remove and just use git hash when building it?

	timeout       time.Duration
	maxproc       float64
	start, stop   string
	primary, sock bool
	version       bool

	procs = NewProcs()
	prim  = NewPrimary()

	test = &Test{} // only for testing
	lg   = &Log{}
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
	commands, flags := Args(os.Args)
	os.Args = flags

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

	prim.SetAll(primary)

	if !sock {
		// If sending something over the socket, don't open it.
		go socket(socketName)
		defer os.Remove(socketName)
	}

	if maxproc > 0.0 {
		if v := os.Getenv("GOMAXPROCS"); v != "" {
			lg.Printf("GOMAXPROCS already set, using that value: %s", v)
		} else {
			numcpu := strconv.Itoa(int(math.Ceil(float64(runtime.NumCPU()) * maxproc)))
			lg.Printf("using %s as GOMAXPROCS", numcpu)
			os.Setenv("GOMAXPROCS", numcpu)
		}
	}

	if start != "" {
		startcmd := command(start)
		if err := startcmd.Run(); err != nil {
			lg.Fatalf("start command failed: %s", err)
		}
	}
	if stop != "" {
		stopcmd := command(stop)
		defer stopcmd.Run()
	}
	if sock {
		err := write(socketName, commands)
		if err != nil {
			lg.Fatalf("failed to write to unix socket: %s", err)
		}
		return
	}
	if os.Getpid() == 1 {
		go reap()
	}
	run(commands, false)
	wait(sock)
}

// run runs the commands as given on the command line. If noprimary is
// true none of the processes will be considered primary.
func run(commands []*exec.Cmd, fromsocket bool) {
	for i, _ := range commands {
		// Need to copy here, because otherwise the closure below will access
		// to wrong command, when we run in a loop.
		c := commands[i]
		if err := c.Start(); err != nil {
			lg.Printf("process failed to start: %v", err)
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
		if test.Test() {
			pid = testPid
		}
		lg.Printf("pid %d started: %v", pid, c.Args)

		procs.Insert(c)

		go func() {
			err := c.Wait()
			pid := c.Process.Pid
			if test.Test() {
				pid = testPid
			}

			switch err {
			default:
				_, ok := err.(*os.SyscallError)
				if !ok {
					lg.Printf("pid %d finished: %v with error: %v", pid, c.Args, err)
					break
				}
				fallthrough
			case nil:
				lg.Printf("pid %d finished: %v", pid, c.Args)

			}

			procs.Remove(c)
			if prim.All() || prim.Primary(c.Process.Pid) {
				if prim.All() {
					lg.Printf("all processes considered primary, signalling other processes")
				} else {
					lg.Printf("pid %d was primary, signalling other processes", pid)
				}
				procs.Cleanup(syscall.SIGINT)
			}
		}()
	}
	return
}

// wait waits for commands to finish.
func wait(sock bool) {
	defer func() { lg.Printf("all processes exited, goodbye!") }()
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
	args, err := ReadArgs(strings.NewReader(arg))
	if err != nil {
		lg.Fatalf("invalid command %q", arg)
	}
	if len(args) == 0 {
		lg.Fatalf("invalid empty command")
	}
	for i, a := range args {
		args[i] = os.ExpandEnv(a)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
