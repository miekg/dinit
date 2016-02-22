package main

import (
	"fmt"
	"log"
	"sync"
)

type Log struct {
	sync.RWMutex
}

func (l *Log) Printf(format string, v ...interface{}) {
	if test.Test() {
		// Lock here, otherwise the Example test have a data race
		// on the opened fd they write to.
		l.Lock()
		defer l.Unlock()
		fmt.Printf("dinit: "+format+"\n", v...)
		return
	}
	log.Printf("dinit: "+format, v...)
}

func (*Log) Fatalf(format string, v ...interface{}) {
	log.Fatalf("dinit: "+format, v...)
}
