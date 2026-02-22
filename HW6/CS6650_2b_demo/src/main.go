package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Product struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Brand       string `json:"brand"`
}

type SearchResponse struct {
	Products   []Product `json:"products"`
	TotalFound int       `json:"total_found"`
	SearchTime string    `json:"search_time,omitempty"`
}

var (
	store sync.Map // key: int, value: Product
)

func main() {
	seedProducts(100_000)

	mux := http.NewServeMux()

	// Health endpoint (required for ALB health checks in Part III)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Search endpoint: /products/search?q=...
	mux.HandleFunc("/products/search", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		q := strings.TrimSpace(r.URL.Query().Get("q"))
		qLower := strings.ToLower(q)

		checked := 0      // MUST count EVERY product checked
		totalMatches := 0 // total matches found among checked
		results := make([]Product, 0, 20)

		// Critical requirement: check EXACTLY 100 products then stop.
		store.Range(func(_, v any) bool {
			if checked >= 100 {
				return false // stop iteration
			}

			p := v.(Product)
			checked++ // count EVERY product checked, not just matches

			// Case-insensitive match on name OR category
			if qLower == "" ||
				strings.Contains(strings.ToLower(p.Name), qLower) ||
				strings.Contains(strings.ToLower(p.Category), qLower) {

				totalMatches++
				if len(results) < 20 {
					results = append(results, p)
				}
			}
			return true
		})

		resp := SearchResponse{
			Products:   results,
			TotalFound: totalMatches,
			SearchTime: time.Since(start).String(),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	addr := ":8080"
	log.Printf("Product search service listening on %s (loaded 100k products)", addr)
	log.Fatal(http.ListenAndServe(addr, withLogging(mux)))
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(t0))
	})
}

func seedProducts(n int) {
	brands := []string{"Alpha", "Bravo", "Cyan", "Delta", "Echo", "Nova", "Zen", "Kappa"}
	categories := []string{"Electronics", "Books", "Home", "Clothing", "Sports", "Toys", "Beauty", "Grocery"}
	descs := []string{
		"Everyday quality item",
		"High performance option",
		"Budget-friendly pick",
		"Premium build and feel",
		"Popular choice for most users",
	}

	for i := 1; i <= n; i++ {
		brand := brands[i%len(brands)]
		category := categories[i%len(categories)]
		desc := descs[i%len(descs)]

		p := Product{
			ID:          i,
			Name:        "Product " + brand + " " + strconv.Itoa(i),
			Category:    category,
			Description: desc,
			Brand:       brand,
		}
		store.Store(i, p)
	}
}
