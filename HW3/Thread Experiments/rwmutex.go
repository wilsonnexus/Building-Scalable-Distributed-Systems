package main

import (
	"fmt"
	"sync"
	"time"
)

const (
	numRuns = 3
	numG    = 50
	numI    = 1000
)

type SafeMap struct {
	mu sync.RWMutex
	m  map[int]int
}

func NewSafeMap() *SafeMap {
	return &SafeMap{m: make(map[int]int)}
}

func (s *SafeMap) Set(k, v int) {
	s.mu.Lock()
	s.m[k] = v
	s.mu.Unlock()
}

func (s *SafeMap) Len() int {
	s.mu.RLock()
	l := len(s.m)
	s.mu.RUnlock()
	return l
}

func runOnce() (int, time.Duration) {
	start := time.Now()

	sm := NewSafeMap()
	var wg sync.WaitGroup

	wg.Add(numG)
	for g := 0; g < numG; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < numI; i++ {
				sm.Set(g*numI+i, i)
			}
		}(g)
	}

	wg.Wait()
	return sm.Len(), time.Since(start)
}

func main() {
	var total time.Duration
	lastLen := 0

	for r := 1; r <= numRuns; r++ {
		l, d := runOnce()
		lastLen = l
		total += d
		fmt.Printf("run %d: len=%d time=%v\n", r, l, d)
	}

	fmt.Printf("mean time over %d runs: %v (last len=%d)\n",
		numRuns, total/numRuns, lastLen)
}
