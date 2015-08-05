package main

import (
	"net"
	"os/exec"
	"strings"
)

const (
	socketName   = "/tmp/dinit.sock"
	socketMaxLen = 512
)

func startCommand(c net.Conn) {
	buf := make([]byte, socketMaxLen)
	n, err := c.Read(buf)

	defer c.Close()
	if err != nil {
		logPrintf("socket: error reading data: %s", err)
		return
	}

	cmdargs := strings.Fields(string(buf[0:n]))
	commands := Args(cmdargs)
	run(commands, true)
}

func socket() {
	l, err := net.Listen("unix", socketName)
	if err != nil {
		logFatalf("socket: listen error: %s", err)
	}

	for {
		fd, err := l.Accept()
		if err != nil {
			logPrintf("socket: accept error: %s", err)
			continue
		}

		go startCommand(fd)
	}
}

func write(cmds []*exec.Cmd) error {
	c, err := net.Dial("unix", socketName)
	if err != nil {
		return err
	}
	str := String(cmds)
	_, err = c.Write([]byte(str))
	if err != nil {
		return err
	}
	return nil
}
