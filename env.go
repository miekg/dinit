package main

import (
	"os"
	"strconv"
	"time"
)

func envInt(k string, d int) int {
	x := os.ExpandEnv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e == nil {
			return x1
		}
	}
	return d
}

func envDuration(k string, d time.Duration) time.Duration {
	x := os.ExpandEnv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e == nil {
			return time.Duration(x1) * time.Second
		}
	}
	return d
}

func envString(k, d string) string {
	x := os.ExpandEnv(k)
	if x != "" {
		return x
	}
	return d
}
