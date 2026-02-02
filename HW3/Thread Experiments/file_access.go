package main

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

const (
	N              = 100000
	unbufferedPath = "unbuffered_output.txt"
	bufferedPath   = "buffered_output.txt"
)

func writeUnbuffered(path string) (time.Duration, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	start := time.Now()
	for i := 0; i < N; i++ {
		// one write call per line
		if _, err := f.Write([]byte(fmt.Sprintf("line %d\n", i))); err != nil {
			return 0, err
		}
	}
	return time.Since(start), nil
}

func writeBuffered(path string) (time.Duration, error) {
	f, err := os.Create(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)

	start := time.Now()
	for i := 0; i < N; i++ {
		// one write call per line (to the buffer)
		if _, err := w.WriteString(fmt.Sprintf("line %d\n", i)); err != nil {
			return 0, err
		}
	}
	if err := w.Flush(); err != nil {
		return 0, err
	}
	return time.Since(start), nil
}

func main() {
	d1, err := writeUnbuffered(unbufferedPath)
	if err != nil {
		fmt.Println("unbuffered error:", err)
		return
	}

	d2, err := writeBuffered(bufferedPath)
	if err != nil {
		fmt.Println("buffered error:", err)
		return
	}

	fmt.Printf("unbuffered: %v\nbuffered:   %v\n", d1, d2)
}
