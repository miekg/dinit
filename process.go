package main

import (
	"os"
	"os/exec"
	"strings"
	"sync"
)

// Commands holds the processes that we run.
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
		if isEnv(a) {
			args[i] = os.ExpandEnv(a)
		}
	}

	cmd := exec.Command(args[0], args[1:]...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd
}
