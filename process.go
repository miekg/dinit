// Dinit is a mini init replacement useful for use inside Docker containers.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	verbose              bool
	port, sleep          int
	namespace, subsystem string
)

func main() {
	flag.IntVar(&port, "port", envInt("DINIT_PORT", 0), "port to export metricss for prometheus (DINIT_PORT)")
	flag.IntVar(&sleep, "sleep", envInt("DINIT_SLEEP", 5), "how many seconds to sleep before force killing programs (DINIT_SLEEP)")
	flag.StringVar(&namespace, "namespace", envString("DINIT_NAMESPACE", ""), "namespace to use for prometheus (DINIT_NAMESPACE)")
	flag.StringVar(&subsystem, "subsystem", envString("DINIT_SUBSYSTEM", ""), "subsystem to use for prometheus (DINIT_SUBSYSTEM)")
	flag.BoolVar(&verbose, "verbose", envBool("DINIT_VERBOSE", false), "be more verbose and show stdout/stderr of programs (DINIT_VERBOSE)")

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: dinit [OPTION]... PROGRAM [PROGRAM]...")
		fmt.Fprintln(os.Stderr, "Start PROGRAMs by passing the enviroment and reap any zombies.\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(flag.Args()) == 0 {
		log.Fatal("dinit: need at least one program")
	}

	if port > 0 {
		metrics()
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

			time.Sleep(time.Duration(sleep) * time.Second)

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
		zombies.Inc()
	}
}

func logf(format string, v ...interface{}) {
	if !verbose {
		return
	}
	log.Printf("dinit: "+format, v...)
}

func envBool(k string, d bool) bool {
	x := os.Getenv(k)
	switch strings.ToLower(x) {
	case "true":
		return true
	case "false":
		return false
	}
	return d

}

func envInt(k string, d int) int {
	x := os.Getenv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e != nil {
			return x1
		}
	}
	return d
}

func envString(k, d string) string {
	x := os.Getenv(k)
	if x != "" {
		return x
	}
	return d
}
