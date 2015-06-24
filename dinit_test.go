package main

import (
	"syscall"
	"time"
)

//import "testing"

func ExampleRun() {
	test = true
	cmd := "echo Hi"

	run([]string{cmd})
	wait()
	// Output: dinit: pid 123 started: [echo Hi]
	// dinit: pid 123, finished: [echo Hi] with error: <nil>
	// dinit: all processes exited, goodbye!
}

func ExampleRunINT() {
	test = true
	cmd := "sleep 10"

	run([]string{cmd})
	go func() {
		time.Sleep(10 * time.Millisecond)
		cmds.Signal(syscall.SIGINT)
	}()
	wait()
	// Output: dinit: pid 123 started: [sleep 10]
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123, finished: [sleep 10] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}
