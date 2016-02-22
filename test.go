package main

import "sync"

type Test struct {
	b bool
	sync.RWMutex
}

func (t *Test) Test() bool {
	t.RLock()
	defer t.RUnlock()
	return t.b
}

func (t *Test) SetTest(b bool) {
	t.Lock()
	defer t.Unlock()
	t.b = b
}
