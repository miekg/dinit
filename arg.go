package main

import (
	"encoding/csv"
	"io"
	"os"
	"os/exec"
)

func Args(args []string) ([]*exec.Cmd, []string) {
	var flags []string
	var cmd *exec.Cmd
	commands := []*exec.Cmd{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-r":
			if i+1 == len(args) {
				lg.Fatalf("need a command after -r")
			}

			if cmd != nil {
				commands = append(commands, cmd)
			}
			cmd = &exec.Cmd{Stdout: os.Stdout, Stderr: os.Stderr}

			path, err := exec.LookPath(os.ExpandEnv(args[i+1]))
			if err != nil {
				lg.Fatalf("invalid command %q: %v", args[i+1], err)
			}
			cmd.Args = []string{path}
			cmd.Path = path

			i++
		default:
			if cmd == nil {
				flags = append(flags, args[i])
				break
			}
			if args[i] == "\\-r" {
				args[i] = "-r"
			}
			cmd.Args = append(cmd.Args, os.ExpandEnv(args[i]))
		}
	}
	if cmd != nil {
		commands = append(commands, cmd)
	}
	return commands, flags
}

// String is the opposite of Args and returns the full command line
// string as first seen.
func String(cmds []*exec.Cmd) []string {
	var s []string
	for _, c := range cmds {
		s = append(s, "-r")
		for _, a := range c.Args {
			if a == "-r" {
				a = "\\-r"
			}
			s = append(s, a)
		}
	}
	return s
}

func ReadArgs(r io.Reader) ([]string, error) {
	reader := csv.NewReader(r)
	reader.Comma = ' '
	reader.FieldsPerRecord = -1
	return reader.Read()
}

func WriteArgs(w io.Writer, args []string) error {
	writer := csv.NewWriter(w)
	writer.Comma = ' '
	err := writer.Write(args)
	if err != nil {
		return err
	}
	writer.Flush()
	return writer.Error()
}
