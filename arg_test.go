package main

import (
	"strings"
	"testing"
)

func TestArgString(t *testing.T) {
	cmdline := "-r /bin/sleep 10 -r /bin/echo -- \\-r"
	if x := String(Args(strings.Fields(cmdline))); x != cmdline {
		t.Fatalf("expected '%s', got '%s'", cmdline, x)
	}
}
