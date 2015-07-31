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
	buf := make([]byte, socketMaxLen)
	n, err := c.Read(buf)

	defer c.Close()
	if err != nil {
		logPrintf("socket: error reading data: %s", err)
		return
	}

	data := string(buf[0:n])
	run([]*exec.Cmd{command(data)}, true)
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
