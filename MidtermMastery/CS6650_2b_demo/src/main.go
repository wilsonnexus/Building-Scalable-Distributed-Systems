package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ---------- Product Model ----------
type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

// ---------- In-memory store ----------
var products []Product

func generateProducts(n int) {
	brands := []string{"Alpha", "Beta", "Gamma", "Delta", "Omega"}
	cats := []string{"Electronics", "Books", "Home", "Sports", "Toys"}
	descs := []string{"Great product", "High quality", "Budget option", "Premium build", "Popular item"}

	products = make([]Product, n)
	for i := 0; i < n; i++ {
		brand := brands[i%len(brands)]
		cat := cats[i%len(cats)]
		products[i] = Product{
			ID:          i + 1,
			Name:        fmt.Sprintf("Product %s %d", brand, i+1),
			Category:    cat,
			Description: descs[i%len(descs)],
			Brand:       brand,
		}
	}
}

// ---------- Downstream simulation ----------
func downstreamHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode") // slow | fail | normal

	switch mode {
	case "fail":
		http.Error(w, "downstream forced fail", http.StatusServiceUnavailable)
		return
	case "slow":
		time.Sleep(600 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
		return
	default:
		// keep existing random behavior if wanted
		p := rand.Float64()
		if p < 0.20 {
			http.Error(w, "downstream error", http.StatusServiceUnavailable)
			return
		}
		if p < 0.50 {
			time.Sleep(time.Duration(250+rand.Intn(350)) * time.Millisecond)
		} else {
			time.Sleep(time.Duration(10+rand.Intn(30)) * time.Millisecond)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}
}

// ---------- Search ----------
type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	Checked    int       `json:"checked"`
	SearchTime string    `json:"search_time,omitempty"`
	Downstream string    `json:"downstream,omitempty"`
	Mode       string    `json:"mode,omitempty"`
}

var activeRequests int32

func searchHandler_BAD(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))
	mode := r.URL.Query().Get("mode") // e.g., mode=crash

	// Track concurrency; optionally crash if we overload (shows task restart)
	cur := atomic.AddInt32(&activeRequests, 1)
	defer atomic.AddInt32(&activeRequests, -1)

	if mode == "crash" && cur >= 20 {
		// Simulated crash: like a fatal error/OOM/panic in production
		panic("simulated crash: too many concurrent requests")
	}

	// BAD: Call downstream with no timeout and no concurrency limit.
	// Under load, goroutines pile up waiting on downstream => latency/CPU spikes.
	dsURL := "http://127.0.0.1:8080/downstream?mode=slow"
	resp, err := http.Get(dsURL)
	for i := 0; i < 9; i++ { // fan-out 10x per request (simulates dependency amplification)
		resp, err := http.Get(dsURL)
		if err == nil && resp != nil {
			_ = resp.Body.Close()
		}
	}
	downstreamStatus := "slow_fanout_10x"
	if err != nil {
		downstreamStatus = "network_error"
	} else {
		_ = resp.Body.Close()
		if resp.StatusCode >= 500 || resp.StatusCode == 429 || resp.StatusCode == 503 {
			downstreamStatus = "error_" + strconv.Itoa(resp.StatusCode)
		}
	}

	// Requirement: always check exactly 100 products
	checked := 0
	results := make([]Product, 0, 20)
	totalFound := 0

	for i := 0; i < len(products) && checked < 100; i++ {
		checked++
		p := products[i]
		if q == "" {
			continue
		}
		if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(strings.ToLower(p.Category), q) {
			totalFound++
			if len(results) < 20 {
				results = append(results, p)
			}
		}
	}

	out := SearchResponse{
		Products:   results,
		TotalFound: totalFound,
		Checked:    checked,
		SearchTime: time.Since(start).String(),
		Downstream: downstreamStatus,
		Mode:       "bad_no_timeout_no_bulkhead",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

// ---------- Bulkhead (limit concurrent downstream work) ----------
var bulkhead = make(chan struct{}, 30) // allow at most 30 concurrent downstream calls

// ---------- Simple Circuit Breaker ----------
type CircuitBreaker struct {
	mu           sync.Mutex
	state        string // "closed" | "open" | "half_open"
	failCount    int
	successCount int
	openUntil    time.Time
}

func NewBreaker() *CircuitBreaker {
	return &CircuitBreaker{state: "closed"}
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "open" {
		if time.Now().After(cb.openUntil) {
			cb.state = "half_open"
			cb.successCount = 0
			return true
		}
		return false
	}
	return true
}

func (cb *CircuitBreaker) OnSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == "half_open" {
		cb.successCount++
		if cb.successCount >= 5 {
			cb.state = "closed"
			cb.failCount = 0
		}
		return
	}
	cb.failCount = 0
}

func (cb *CircuitBreaker) OnFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failCount++
	// Open breaker after 10 failures
	if cb.failCount >= 10 && cb.state != "open" {
		cb.state = "open"
		cb.openUntil = time.Now().Add(10 * time.Second) // cool down
	}
}

var breaker = NewBreaker()

func callDownstreamWithProtections() (string, error) {
	// Fail fast if breaker is open
	if !breaker.Allow() {
		return "breaker_open", fmt.Errorf("circuit breaker open")
	}

	// Bulkhead: limit concurrent downstream calls
	select {
	case bulkhead <- struct{}{}:
		defer func() { <-bulkhead }()
	default:
		return "bulkhead_reject", fmt.Errorf("bulkhead full")
	}

	// Fail fast timeout
	client := &http.Client{Timeout: 120 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:8080/downstream?mode=slow")
	if err != nil {
		breaker.OnFailure()
		return "timeout_or_net_error", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 || resp.StatusCode == 429 || resp.StatusCode == 503 {
		breaker.OnFailure()
		return "error_" + strconv.Itoa(resp.StatusCode), fmt.Errorf("downstream status %d", resp.StatusCode)
	}

	breaker.OnSuccess()
	return "ok", nil
}

func searchHandler_FIXED(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	q := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("q")))

	downstreamStatus, _ := callDownstreamWithProtections()
	// Even if downstream fails, we still respond quickly (graceful degradation)

	checked := 0
	results := make([]Product, 0, 20)
	totalFound := 0

	for i := 0; i < len(products) && checked < 100; i++ {
		checked++
		p := products[i]
		if q == "" {
			continue
		}
		if strings.Contains(strings.ToLower(p.Name), q) || strings.Contains(strings.ToLower(p.Category), q) {
			totalFound++
			if len(results) < 20 {
				results = append(results, p)
			}
		}
	}

	out := SearchResponse{
		Products:   results,
		TotalFound: totalFound,
		Checked:    checked,
		SearchTime: time.Since(start).String(),
		Downstream: downstreamStatus,
		Mode:       "fixed_bulkhead_cb_failfast",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "bad"
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"mode":  mode,
		"build": "step3-demo-v1",
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())
	generateProducts(100000)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/downstream", downstreamHandler)
	mux.HandleFunc("/version", versionHandler)
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("MODE"))) // "bad" or "fixed"
	if mode == "" {
		mode = "bad"
	}

	if mode == "fixed" {
		log.Println("MODE=fixed: using searchHandler_FIXED")
		mux.HandleFunc("/products/search", searchHandler_FIXED)
	} else {
		log.Println("MODE=bad: using searchHandler_BAD")
		mux.HandleFunc("/products/search", searchHandler_BAD)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("listening on :8080")
	log.Fatal(srv.ListenAndServe())
}
