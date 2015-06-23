package main

import (
	"os"
	"os/exec"
	"strings"
	"sync"
)

type Commands struct {
	sync.RWMutex
	pids map[int]*exec.Cmd
}

func NewCommands() *Commands {
	c := new(Commands)
	c.pids = make(map[int]*exec.Cmd)
	return c
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

func (c *Commands) Signal(sig os.Signal) {
	c.RLock()
	defer c.RUnlock()
	for pid, cmd := range c.pids {
		logF("signal %d sent to pid %d", sig, pid)
		cmd.Process.Signal(sig)
	}
}

func (c *Commands) Len() int {
	return len(c.pids)
}

// command parses arg and return an *exec.Cmd that is ready to be run.
func command(arg string) *exec.Cmd {
	args := strings.Fields(arg) // Split on spaces and execute.
	cmd := exec.Command(args[0], args[1:]...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
