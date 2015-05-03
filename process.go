package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
)

// Cmd holds our commands.
type Cmd struct {
	sync.RWMutex
	Cmd map[int]*exec.Cmd
}

func New() *Cmd { return &Cmd{sync.RWMutex{}, make(map[int]*exec.Cmd)} }

func (c *Cmd) Start(cmd *exec.Cmd) error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	err := cmd.Start()
	if err != nil {
		c.Cmd[cmd.Process.Pid] = cmd
	}
	return err
}

func (c *Cmd) Stop(pid int) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	delete(c.Cmd, pid)
}

// Child return true is pid is a direct child of ours.
func (c *Cmd) Child(pid int) bool {
	if _, ok := c.Cmd[pid]; ok {
		return true
	}
	return false
}

func main() {
	restart := flag.Bool("r", false, "restart programs when they die")
	flag.Parse()

	Cmd := New()
	done := make(chan int)

	for _, arg := range flag.Args() {
		// Split on spaces and execute. Note that with docker we can only have
		// one entry point, but we support multiple.
		args := strings.Fields(arg)
		cmd := exec.Command(args[0], args[1:]...)

		go func() {
			err := Cmd.Start(cmd)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("dinit: command [pid %d] started: %v, pid %d", cmd.Process.Pid, cmd.Args)

			err = cmd.Wait()
			log.Printf("dinit: command [pid %d], finished with error: %v", cmd.Process.Pid, err)
			done <- cmd.Process.Pid
		}()
	}

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGCHLD)

	i := 0
Wait:
	for {
		select {
		case args := <-done:
			if *restart {

			}
			log.Printf("dinit: command [pid %d] finished: %v", 10, args)
			i++
			if len(cmds) == i {
				break Wait
			}
		case sig := <-sigs:
			if sig == syscall.SIGCHLD {
				// If for my own children don't wait here, as we were waiting above.
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
