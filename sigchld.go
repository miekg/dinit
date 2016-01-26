package main

import (
	"os"
	"os/signal"
	"syscall"
)

func sigChld() {
	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGCHLD)

	for {
		select {
		case <-sigs:
			reap()
		default:
		}
	}

}

func reap() {
	for {
		println("called")
		var wstatus syscall.WaitStatus

		pid, err := syscall.Wait4(-1, &wstatus, 0, nil)
		switch err {
		case syscall.EINTR:
			pid, err = syscall.Wait4(-1, &wstatus, 0, nil)
		case syscall.ECHILD:
			return
		}

		logPrintf("pid %d, finished, wstatus: %+v", pid, wstatus)

	}
}
