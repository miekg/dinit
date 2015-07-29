package main

import (
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"
)

func TestEnv(t *testing.T) {
	varname := "DINIT_BOOVAR"
	os.Setenv(varname, "")
	c := command("echo " + "$" + varname)
	if c.Args[1] != "" {
		t.Fatalf("%s should be a env. var", varname)
	}
	os.Setenv(varname, "blah")
	c = command("echo " + "$" + varname)
	if c.Args[1] != "blah" {
		t.Fatalf("%s should be a env. var", varname)
	}
	os.Setenv(varname, "blah")
	c = command("echo " + "$" + varname + ".morestuff")
	if c.Args[1] != "blah.morestuff" {
		t.Fatalf("%s should be a env. var", varname)
	}
}

func ExampleRun() {
	test = true

	run([]*exec.Cmd{command("cat /dev/null")})
	wait()
	// Output: dinit: pid 123 started: [cat /dev/null]
	// dinit: pid 123, finished: [cat /dev/null] with error: <nil>
	// dinit: all processes exited, goodbye!
}

func ExampleRunINT() {
	test = true

	run([]*exec.Cmd{command("sleep 10")})
	go func() {
		time.Sleep(10 * time.Millisecond)
		procs.Signal(syscall.SIGINT)
	}()
	wait()
	// Output: dinit: pid 123 started: [sleep 10]
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123, finished: [sleep 10] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}

func ExampleFailToStart() {
	test = true
	run([]*exec.Cmd{command("sleep 10"), command("verbose")})
	wait()
	// Output: dinit: pid 123 started: [sleep 10]
	// dinit: exec: "verbose": executable file not found in $PATH
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123, finished: [sleep 10] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}

func ExampleTestAllPrimary() {
	test = true
	primary = true
	run([]*exec.Cmd{command("sleep 2"), command("sleep 20")})
	wait()
	// Output: dinit: pid 123 started: [sleep 2]
	// dinit: pid 123 started: [sleep 20]
	// dinit: pid 123, finished: [sleep 2] with error: <nil>
	// dinit: all processes considered primary, signalling other processes
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123, finished: [sleep 20] with error: signal: interrupt
	// dinit: all processes considered primary, signalling other processes
	// dinit: all processes exited, goodbye!
}

// Test is flaky because of random output ordering.
func exampleTestPrimary() {
	test = true
	run([]*exec.Cmd{command("less -"), command("killall -SEGV cat"), command("cat")})
	wait()
	// Output: dinit: pid 123 started: [less -]
	// dinit: pid 123 started: [killall -SEGV cat]
	// dinit: pid 123, finished: [less -] with error: <nil>
	// dinit: pid 123 started: [cat]
	// dinit: pid 123, finished: [cat] with error: <nil>
	// dinit: pid 123 was primary, signalling other processes
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123, finished: [killall -SEGV cat] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}
