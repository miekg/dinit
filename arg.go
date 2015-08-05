package main

import (
	"os"
	"os/exec"
)

func Args(args []string) []*exec.Cmd {
	var (
		cmd  *exec.Cmd
		seen bool
	)
	commands := []*exec.Cmd{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-r":
			if cmd != nil {
				commands = append(commands, cmd)
			}

			seen = true

			cmd = &exec.Cmd{Stdout: os.Stdout, Stderr: os.Stderr}

			if i+1 == len(args) {
				logFatalf("need a command after -r")
			}
			cmd.Args = append(cmd.Args, args[i+1])
			cmd.Path = args[i+1]

			// Clear the args so flag parsing keeps working.
			args[i] = ""
			args[i+1] = ""

			i++
			continue
		case "\\-r":
			args[i] = "-r"
		}

		if seen {
			cmd.Args = append(cmd.Args, os.ExpandEnv(args[i]))
			args[i] = ""
		}
	}
	if cmd != nil {
		commands = append(commands, cmd)
	}
	return commands
}

// String is the opposite of Args and returns the full command line
// string as first seen.
func String(cmds []*exec.Cmd) string {
	// Bit lame that we encode and decode twice when writing to the socket...
	s := ""
	for i, c := range cmds {
		s += "-r "
		for j, a := range c.Args {
			if a == "-r" {
				a = "\\-r"
			}
			if j == len(c.Args)-1 {
				s += a
				continue
			}
			s += a + " "
		}
		if i < len(cmds)-1 {
			s += " "
		}
	}
	return s
}
