package main

import (
	"net"
	"os/exec"
)

const (
	socketName   = "/tmp/dinit.sock"
	socketMaxLen = 512
)

func startCommand(c net.Conn) {
	defer c.Close()

	cmdargs, err := ReadArgs(c)
	if err != nil {
		lg.Printf("socket: error reading data: %s", err)
		return
	}

	commands, _ := Args(cmdargs)
	run(commands, true)
}

func socket(sock string) {
	l, err := net.Listen("unix", sock)
	if err != nil {
		lg.Fatalf("socket: listen error: %s", err)
	}

	lg.Printf("socket: successfully created")

	for {
		fd, err := l.Accept()
		if err != nil {
			lg.Printf("socket: accept error: %s", err)
			continue
		}

		go startCommand(fd)
	}
}

func write(sock string, cmds []*exec.Cmd) error {
	c, err := net.Dial("unix", sock)
	if err != nil {
		return err
	}
	defer c.Close()
	return WriteArgs(c, String(cmds))
}
