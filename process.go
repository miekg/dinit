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

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGCHLD)

	i := 0
Wait:
	for {
		select {
		case <-done:
			i++
			if len(cmds) == i {
				break Wait
			}
		case sig := <-sigs:
			if sig == syscall.SIGCHLD {
				// If for my own children don't wait here, as we were waiting above.
				go func() {
					for {
						var wstatus syscall.WaitStatus
						pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
						if err != nil {
							return
						}
						log.Printf("dinit: pid %d reaped", pid)
						zombies.Inc()
					}
				}()
				break
			}

			// There is a race here, because the process could have died, we don't care.
			for _, cmd := range cmds {
				log.Printf("dinit: signal %d sent to pid %d", sig, cmd.Process.Pid)
				cmd.Process.Signal(sig)
			}
			// TODO(miek): should be conditional, i.e. only kill whats is really left.
				log.Printf("dinit: SIGKILL remaining processes in %d seconds", sleep)
			time.Sleep(time.Duration(sleep) * time.Second)
			for _, cmd := range cmds {
				cmd.Process.Signal(syscall.SIGKILL)
			}
			break Wait
		}
	}
}
