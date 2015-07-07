package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestIsEnv(t *testing.T) {
	varname := "DINIT_BOOVAR"
	if isEnv(varname) {
		t.Fatalf("%s should not be a env. var", varname)
	}
	os.Setenv(varname, "blah")
	if isEnv(varname) {
		t.Fatalf("%s should be a env. var", varname)
	}
	t.Logf("var %s, value %s", varname, os.Getenv(varname))
}

func TestEnv(t *testing.T) {
	varname := "DINIT_BOOVAR"
	os.Setenv(varname, "")
	c := command("echo " + "$" + varname)
	if c.Args[1] != "$"+varname {
		t.Fatalf("%s should not be a env. var", varname)
	}
	os.Setenv(varname, "blah")
	c = command("echo " + "$" + varname)
	if c.Args[1] != "blah" {
		t.Fatalf("%s should be a env. var", varname)
	}
}

func ExampleRun() {
	test = true

	run([]string{"echo Hi"})
	wait()
	// Output: dinit: pid 123 started: [echo Hi]
	// Hi
	// dinit: pid 123, finished: [echo Hi] with error: <nil>
	// dinit: all processes exited, goodbye!
}

func ExampleRunINT() {
	test = true

	run([]string{"sleep 10"})
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

func ExampleFailToStart() {
	test = true
	run([]string{"sleep 10", "verbose"})
	wait()
	// Output: dinit: pid 123 started: [sleep 10]
	// dinit: exec: "verbose": executable file not found in $PATH
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123, finished: [sleep 10] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}
