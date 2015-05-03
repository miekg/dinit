// Dinit is a mini init replacement useful for use inside Docker containers.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	port, sleep          int
	namespace, subsystem string
)

func main() {
	flag.IntVar(&port, "port", 0, "port to export metricss for Prometheus")
	flag.IntVar(&sleep, "sleep", 5, "how many seconds to sleep before force killing programs")
	flag.StringVar(&namespace, "namespace", "", "namespace to use for Prometheus")
	flag.StringVar(&subsystem, "subsystem", "", "subsystem to use for Prometheus")

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
			err := cmd.Start()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("dinit: pid %d started: %v", cmd.Process.Pid, cmd.Args)

			err = cmd.Wait()
			if err != nil {
				log.Printf("dinit: pid %d, finished with error: %s", cmd.Process.Pid, err)
			} else {
				log.Printf("dinit: pid %d, finished: %v", cmd.Process.Pid, cmd.Args)
			}
			done <- true
		}()
	}

	ints := make(chan os.Signal)
	chld := make(chan os.Signal)
	signal.Notify(ints, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(chld, syscall.SIGCHLD)

	i := 0
Wait:
	for {
		select {
		case <-chld:
			go reaper()
		case <-done:
			i++
			if len(cmds) == i {
				reaper()
				break Wait
			}
		case sig := <-ints:
			// There is a race here, because the process could have died, we don't care.
			for _, cmd := range cmds {
				log.Printf("dinit: signal %d sent to pid %d", sig, cmd.Process.Pid)
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
				log.Printf("dinit: SIGKILL sent to pid %d", p.Pid)
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
		log.Printf("dinit: pid %d reaped", pid)
		zombies.Inc()
	}
}
