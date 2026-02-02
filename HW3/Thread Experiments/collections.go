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

func runOnce() (int, time.Duration) {
	start := time.Now()

	m := make(map[int]int)
	var wg sync.WaitGroup

	wg.Add(numG)
	for g := 0; g < numG; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < numI; i++ {
				m[g*numI+i] = i // unsafe concurrent write
			}
		}(g)
	}

	wg.Wait()
	return len(m), time.Since(start)
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
