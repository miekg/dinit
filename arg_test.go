package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestArgString(t *testing.T) {
	cmdline := "-r /bin/sleep 10 -r /bin/echo -- \\-r"
	commands, flags := Args(strings.Fields(cmdline))
	if x := strings.Join(String(commands), " "); x != cmdline {
		t.Fatalf("expected '%s', got '%s'", cmdline, x)
	}
	if len(flags) != 0 {
		t.Fatalf("expected 0 flags, got %q", flags)
	}
}

func TestArgStringSpaces(t *testing.T) {
	cmdline := []string{"-r", "/bin/bash", "-c", "/bin/sleep 10"}
	commands, flags := Args(cmdline)
	if x := String(commands); !reflect.DeepEqual(x, cmdline) {
		t.Fatalf("expected %q, got %q", cmdline, x)
	}
	if len(flags) != 0 {
		t.Fatalf("expected 0 flags, got %q", flags)
	}
}
