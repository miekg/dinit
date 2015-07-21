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

	run([]string{"cat /dev/null"})
	wait()
	// Output: dinit: pid 123 started: [cat /dev/null]
	// dinit: pid 123, finished: [cat /dev/null] with error: <nil>
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

// Test is flaky because the ordering of the output is not fixed.
func examplePrimary() {
	test = true
	run([]string{"cat /dev/zero", "less -f /dev/zero", "killall -SEGV cat"})
	wait()
	// Output: dinit: pid 123 started: [cat /dev/zero]
	// dinit: pid 123 started: [less -f /dev/zero]
	// dinit: pid 123, finished: [cat /dev/zero] with error: signal: segmentation fault
	// dinit: pid 123 was primary, signalling other processes
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123 started: [killall -SEGV cat]
	// dinit: pid 123, finished: [killall -SEGV cat] with error: <nil>
	// dinit: pid 123, finished: [less -f /dev/zero] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}
