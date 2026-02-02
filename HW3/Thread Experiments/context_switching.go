package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const roundTrips = 1_000_000

func measure(maxProcs int) (total time.Duration, avg time.Duration) {
	runtime.GOMAXPROCS(maxProcs)

	ch := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	start := time.Now()

	// Goroutine A: send then receive (1 "round trip" per loop)
	go func() {
		defer wg.Done()
		for i := 0; i < roundTrips; i++ {
			ch <- struct{}{} // A -> B
			<-ch             // B -> A
		}
	}()

	// Goroutine B: receive then send
	for i := 0; i < roundTrips; i++ {
		<-ch             // A -> B
		ch <- struct{}{} // B -> A
	}

	wg.Wait()
	total = time.Since(start)
	avg = total / time.Duration(2*roundTrips) // 2 hand-offs per round trip
	return
}

func main() {
	total1, avg1 := measure(1)
	fmt.Printf("GOMAXPROCS(1): total=%v avg_switch=%v\n", total1, avg1)

	n := runtime.NumCPU()
	totalN, avgN := measure(n)
	fmt.Printf("GOMAXPROCS(%d): total=%v avg_switch=%v\n", n, totalN, avgN)
}
