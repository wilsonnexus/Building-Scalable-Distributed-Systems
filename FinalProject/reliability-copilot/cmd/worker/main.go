package main

import (
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"reliability-copilot/internal/classifier"
	"reliability-copilot/internal/store"
)

func main() {
	dbPath := getenv("DB_PATH", "reliability.db")
	st, err := store.New(dbPath)
	if err != nil {
		log.Fatal(err)
	}

	workers := atoi(getenv("WORKERS", "1"))
	inferenceDelayMs := atoi(getenv("INFERENCE_DELAY_MS", "0"))
	dbDelayMs := atoi(getenv("DB_DELAY_MS", "0"))
	failRatePct := atoi(getenv("FAIL_RATE_PCT", "0"))
	protectionEnabled := getenv("PROTECTION_ENABLED", "true") == "true"

	log.Printf("worker starting with WORKERS=%d INFERENCE_DELAY_MS=%d DB_DELAY_MS=%d FAIL_RATE_PCT=%d PROTECTION_ENABLED=%v",
		workers, inferenceDelayMs, dbDelayMs, failRatePct, protectionEnabled)

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for {
				event, err := st.ClaimNextJob()
				if err != nil {
					log.Printf("worker %d claim error: %v", id, err)
					time.Sleep(500 * time.Millisecond)
					continue
				}
				if event == nil {
					time.Sleep(300 * time.Millisecond)
					continue
				}

				mode := "rule-based-worker"

				if dbDelayMs > 0 {
					time.Sleep(time.Duration(dbDelayMs) * time.Millisecond)
				}

				if failRatePct > 0 && time.Now().UnixNano()%100 < int64(failRatePct) {
					if protectionEnabled {
						_ = st.Complete(event.EventID, "unknown_failure", "Fallback recommendation due to temporary worker failure.", 0.50, "fallback-protected")
					} else {
						_ = st.Fail(event.EventID, "simulated worker failure", "unprotected")
					}
					continue
				}

				if inferenceDelayMs > 0 {
					if protectionEnabled && inferenceDelayMs > 1500 {
						_ = st.Complete(event.EventID, "unknown_failure", "Fallback recommendation due to slow inference path.", 0.52, "fallback-timeout")
						continue
					}
					time.Sleep(time.Duration(inferenceDelayMs) * time.Millisecond)
				}

				category, recommendation, score := classifier.Classify(event.LogText)
				if err := st.Complete(event.EventID, category, recommendation, score, mode); err != nil {
					log.Printf("worker %d complete error: %v", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
