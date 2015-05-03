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
)

func main() {
	flag.Parse()

	// make map and protect with mutex to keep track of children.
	cmds := []*exec.Cmd{}
	done := make(chan []string)
	for _, arg := range flag.Args() {
		// Split on spaces and execute. Note that with docker we can only have
		// one entry point, but we support multiple.
		args := strings.Fields(arg)
		cmd := exec.Command(args[0], args[1:]...)
		cmds = append(cmds, cmd)

		go func() {
			log.Printf("dinit: command [pid %d] started: %v, pid %d", cmd.Process.Pid, cmd.Args)
			err := cmd.Start()
			if err != nil {
				log.Fatal(err)
			}

			err = cmd.Wait()
			log.Printf("dinit: command [pid %d], finished with error: %v", cmd.Process.Pid, err)
			done <- cmd.Args
		}()
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGCHLD)

	i := 0
Wait:
	for {
		select {
		case args := <-done:
			log.Printf("dinit: coommand finished: %v", args)
			i++
			if len(cmds) == i {
				break Wait
			}
		case sig := <-sigs:
			if sig == syscall.SIGCHLD {
				// If for my own children don't wait here, as we were waiting
				// above.
				var wstatus syscall.WaitStatus

				for {
					pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
					if err != nil {
						fmt.Println("syscall err ", err)
						break
					}

					fmt.Println("Child PID", pid)

				}
				break
			}

			// There is a race here, because the process could have died. We don't care.
			for _, cmd := range cmds {
				log.Printf("dinit: signal %d sent to %d", sig, cmd.Process.Pid)
				cmd.Process.Signal(sig)
			}
		}
	}
}
