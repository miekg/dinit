package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func envBool(k string, d bool) bool {
	x := os.Getenv(k)
	switch strings.ToLower(x) {
	case "true":
		return true
	case "false":
		return false
	}
	return d

}

func envInt(k string, d int) int {
	x := os.Getenv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e != nil {
			return x1
		}
	}
	return d
}

func envDuration(k string, d time.Duration) time.Duration {
	x := os.Getenv(k)
	if x != "" {
		if x1, e := strconv.Atoi(x); e != nil {
			return time.Duration(x1) * time.Second
		}
	}
	return d
}

func envString(k, d string) string {
	x := os.Getenv(k)
	if x != "" {
		return x
	}
	return d
}
