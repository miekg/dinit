package main

import (
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Primary holds which pid is considered the primary process. If that
// dies, the whole container should be killed.
type Primary struct {
	sync.RWMutex
	first map[int]bool
}

func NewPrimary() *Primary {
	return &Primary{first: make(map[int]bool)}
}

func (p *Primary) Set(pid int) {
	p.Lock()
	defer p.Unlock()
	p.first[pid] = true
}

func (p *Primary) Primary(pid int) bool {
	p.RLock()
	defer p.RUnlock()
	_, ok := p.first[pid]
	return ok
}

// Commands holds the processes that we run.
type Commands struct {
	sync.RWMutex
	pids map[int]*exec.Cmd
}

func NewCommands() *Commands {
	return &Commands{pids: make(map[int]*exec.Cmd)}
}

func (c *Commands) Insert(cmd *exec.Cmd) {
	c.Lock()
	defer c.Unlock()
	c.pids[cmd.Process.Pid] = cmd
}

func (c *Commands) Remove(cmd *exec.Cmd) {
	c.Lock()
	defer c.Unlock()
	delete(c.pids, cmd.Process.Pid)
}

// Signal sends sig to all processes in Commands.
func (c *Commands) Signal(sig os.Signal) {
	c.RLock()
	defer c.RUnlock()
	for pid, cmd := range c.pids {
		if test {
			logPrintf("signal %d sent to pid %d", sig, testPid)
		} else {
			logPrintf("signal %d sent to pid %d", sig, pid)
		}
		cmd.Process.Signal(sig)
	}
}

// Cleanup will send signal sig to the commands and after a short time send
// a SIGKKILL.
func (c *Commands) Cleanup(sig os.Signal) {
	cmds.Signal(sig)

	time.Sleep(2 * time.Second)

	if cmds.Len() > 0 {
		logPrintf("%d processes still alive after SIGINT/SIGTERM", cmds.Len())
		time.Sleep(timeout)
	}
	cmds.Signal(syscall.SIGKILL)
}

// Len returns the number of processs in Commands.
func (c *Commands) Len() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.pids)
}

// command parses arg and returns an *exec.Cmd that is ready to be run.
func command(arg string) *exec.Cmd {
	args := strings.Fields(arg) // Split on spaces and execute.
	// Loop to check for env vars
	for i, a := range args {
		args[i] = os.ExpandEnv(a)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}
