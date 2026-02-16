package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
OpenAPI Product schema:
required: product_id, sku, manufacturer, category_id, weight, some_other_id
*/

type Product struct {
	ProductID    int32  `json:"product_id"`
	SKU          string `json:"sku"`
	Manufacturer string `json:"manufacturer"`
	CategoryID   int32  `json:"category_id"`
	Weight       int32  `json:"weight"`
	SomeOtherID  int32  `json:"some_other_id"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type productStore struct {
	mu       sync.RWMutex
	products map[int32]Product
}

func newProductStore() *productStore {
	return &productStore{products: make(map[int32]Product)}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func internalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeJSON(w, http.StatusInternalServerError, ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: "Internal server error",
	})
}

// Routes supported:
// GET  /products/{productId}
// POST /products/{productId}/details
func parseProductID(path string) (id int32, ok bool, isProductRoute bool) {
	trim := strings.Trim(path, "/")
	parts := strings.Split(trim, "/")

	if len(parts) < 2 || parts[0] != "products" {
		return 0, false, false
	}
	isProductRoute = true

	v, err := strconv.ParseInt(parts[1], 10, 32)
	if err != nil || v < 1 {
		return 0, false, true
	}
	return int32(v), true, true
}

func validateProductBody(p Product, pathID int32) error {
	if p.ProductID < 1 {
		return errors.New("product_id must be a positive integer")
	}
	// Strong contract behavior: path productId must match body product_id
	if p.ProductID != pathID {
		return fmt.Errorf("product_id (%d) must match path productId (%d)", p.ProductID, pathID)
	}
	sku := strings.TrimSpace(p.SKU)
	if len(sku) < 1 || len(sku) > 100 {
		return errors.New("sku must be 1-100 characters")
	}
	m := strings.TrimSpace(p.Manufacturer)
	if len(m) < 1 || len(m) > 200 {
		return errors.New("manufacturer must be 1-200 characters")
	}
	if p.CategoryID < 1 {
		return errors.New("category_id must be a positive integer")
	}
	if p.Weight < 0 {
		return errors.New("weight must be >= 0")
	}
	if p.SomeOtherID < 1 {
		return errors.New("some_other_id must be a positive integer")
	}
	return nil
}

func (s *productStore) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok, isProductRoute := parseProductID(r.URL.Path)
	if !isProductRoute {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Route not found",
		})
		return
	}
	if !ok {
		// spec only lists 404/500 for GET. Invalid productId -> treat as not found.
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "PRODUCT_NOT_FOUND",
			Message: "Product not found",
			Details: "productId must be an integer >= 1",
		})
		return
	}

	s.mu.RLock()
	p, exists := s.products[id]
	s.mu.RUnlock()

	if !exists {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "PRODUCT_NOT_FOUND",
			Message: "Product not found",
			Details: fmt.Sprintf("No product with id=%d", id),
		})
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (s *productStore) handleAddProductDetails(w http.ResponseWriter, r *http.Request) {
	id, ok, isProductRoute := parseProductID(r.URL.Path)
	if !isProductRoute || !strings.HasSuffix(r.URL.Path, "/details") {
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Route not found",
		})
		return
	}
	if !ok {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "The provided input data is invalid",
			Details: "productId must be an integer >= 1",
		})
		return
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var p Product
	if err := dec.Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "The provided input data is invalid",
			Details: "Body must be valid JSON matching Product schema",
		})
		return
	}

	if err := validateProductBody(p, id); err != nil {
		writeJSON(w, http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_INPUT",
			Message: "The provided input data is invalid",
			Details: err.Error(),
		})
		return
	}

	// Spec says to add or update detailed information but also list 404.
	// Implement:
	// - If product exists: update and return 204
	// - If product does NOT exist: create it and return 204 (simpler "store bootstrap")
	// If you want strict 404-on-missing, flip createMissing to false.
	createMissing := true

	s.mu.Lock()
	_, exists := s.products[id]
	if !exists && !createMissing {
		s.mu.Unlock()
		writeJSON(w, http.StatusNotFound, ErrorResponse{
			Error:   "PRODUCT_NOT_FOUND",
			Message: "Product not found",
			Details: fmt.Sprintf("No product with id=%d", id),
		})
		return
	}
	s.products[id] = p
	s.mu.Unlock()

	// 204 No Content per spec
	w.WriteHeader(http.StatusNoContent)
}

func methodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, ErrorResponse{
		Error:   "METHOD_NOT_ALLOWED",
		Message: "Method not allowed",
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	store := newProductStore()
	mux := http.NewServeMux()

	// Products routes
	mux.HandleFunc("/products/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				internalError(w, fmt.Errorf("panic: %v", rec))
			}
		}()

		switch {
		case r.Method == http.MethodGet && !strings.HasSuffix(r.URL.Path, "/details"):
			store.handleGetProduct(w, r)
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/details"):
			store.handleAddProductDetails(w, r)
		default:
			methodNotAllowed(w)
		}
	})

	// Basic health endpoint (helps on ECS)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":8080"
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, withLogging(mux)))
}
