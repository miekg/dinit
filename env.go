package main

import (
	"os"
	"strconv"
	"strings"
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

func envString(k, d string) string {
	x := os.Getenv(k)
	if x != "" {
		return x
	}
	return d
}
