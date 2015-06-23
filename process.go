package main

import (
	"os"
	"os/exec"
	"strings"
)

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
