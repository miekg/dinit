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

func TestStart(t *testing.T) {
	cmd := "/bin/sleep 1"
	os.Setenv("DINIT_START", cmd)
	start := envString("$DINIT_START", "")
	if start != cmd {
		t.Fatalf("got %s, expected %s", cmd, start)
	}
}

func ExampleRun() {
	test.SetTest(true)

	run([]*exec.Cmd{command("cat /dev/null")}, false)
	wait(false)
	// Output: dinit: pid 123 started: [cat /dev/null]
	// dinit: pid 123 finished: [cat /dev/null]
	// dinit: pid 123 was primary, signalling other processes
	// dinit: all processes exited, goodbye!
}

func ExampleRunINT() {
	test.SetTest(true)

	run([]*exec.Cmd{command("sleep 10")}, false)
	go func() {
		time.Sleep(10 * time.Millisecond)
		procs.Signal(syscall.SIGINT)
	}()
	wait(false)
	// Output: dinit: pid 123 started: [sleep 10]
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123 finished: [sleep 10] with error: signal: interrupt
	// dinit: pid 123 was primary, signalling other processes
	// dinit: all processes exited, goodbye!
}

func ExampleFailToStart() {
	test.SetTest(true)
	run([]*exec.Cmd{command("sleep 10"), command("verbose")}, false)
	wait(false)
	// Output: dinit: pid 123 started: [sleep 10]
	// dinit: process failed to start: exec: "verbose": executable file not found in $PATH
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123 finished: [sleep 10] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}

func ExampleTestAllPrimary() {
	test.SetTest(true)
	prim.SetAll(true)
	defer prim.SetAll(false)
	run([]*exec.Cmd{command("sleep 2"), command("sleep 20")}, false)
	wait(false)
	time.Sleep(3 * time.Second) // wait for Cleanup to terminate
	// Output: dinit: pid 123 started: [sleep 2]
	// dinit: pid 123 started: [sleep 20]
	// dinit: pid 123 finished: [sleep 2]
	// dinit: all processes considered primary, signalling other processes
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123 finished: [sleep 20] with error: signal: interrupt
	// dinit: all processes considered primary, signalling other processes
	// dinit: all processes exited, goodbye!
}

// Test is flaky because of random output ordering.
func ExampleTestSubmit() {
	const name = "dinit.sock"

	test.SetTest(true)
	go socket(name)
	defer os.Remove(name)
	time.Sleep(1 * time.Second)

	run([]*exec.Cmd{command("sleep 3")}, false)
	write(name, []*exec.Cmd{command("/bin/bash -c \"trap '' INT; /bin/sleep 4\"")})

	time.Sleep(1 * time.Second)

	procs.Signal(syscall.SIGINT)
	wait(true)
	// Output:
	// dinit: socket: successfully created
	// dinit: pid 123 started: [sleep 3]
	// dinit: pid 123 started: [/bin/bash -c trap '' INT; /bin/sleep 4]
	// dinit: signal 2 sent to pid 123
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123 finished: [sleep 3] with error: signal: interrupt
	// dinit: pid 123 was primary, signalling other processes
	// dinit: signal 2 sent to pid 123
	// dinit: 1 processes still alive after SIGINT/SIGTERM
	// dinit: signal 9 sent to pid 123
	// dinit: pid 123 finished: [/bin/bash -c trap '' INT; /bin/sleep 4] with error: signal: killed
	// dinit: all processes exited, goodbye!
}

// Test is flaky because of random output ordering.
func exampleTestPrimary() {
	test.SetTest(true)
	run([]*exec.Cmd{command("less -"), command("killall -SEGV cat"), command("cat")}, false)
	wait(false)
	// Output: dinit: pid 123 started: [less -]
	// dinit: pid 123 started: [killall -SEGV cat]
	// dinit: pid 123 finished: [less -]
	// dinit: pid 123 started: [cat]
	// dinit: pid 123 finished: [cat]
	// dinit: pid 123 was primary, signalling other processes
	// dinit: signal 2 sent to pid 123
	// dinit: pid 123 finished: [killall -SEGV cat] with error: signal: interrupt
	// dinit: all processes exited, goodbye!
}

// Can test outside of Docker - i.e. with proper init running.
func exampleTestSubProcessReaping() {
	run([]*exec.Cmd{command("./zombie.sh"), command("less - ")}, false)
	wait(false)
	time.Sleep(5 * time.Second)
}
