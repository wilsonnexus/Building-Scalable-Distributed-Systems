package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Record struct {
	Value   string `json:"value"`
	Version int    `json:"version"`
}

type SetRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ReplicateRequest struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Version int    `json:"version"`
}

type GetResponse struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Version  int    `json:"version"`
	ServedBy string `json:"served_by"`
}

var (
	nodeID = getenv("NODE_ID", "node1")
	peers  = splitCSV(os.Getenv("PEERS"))

	store   = map[string]Record{}
	storeMu sync.RWMutex

	httpClient = &http.Client{Timeout: 5 * time.Second}
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/set", setHandler)
	mux.HandleFunc("/get/", getHandler)
	mux.HandleFunc("/local_read/", localReadHandler)
	mux.HandleFunc("/internal/replicate", internalReplicateHandler)

	addr := ":8000"
	fmt.Printf("starting leaderless node=%s on %s\n", nodeID, addr)
	http.ListenAndServe(addr, mux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"node_id": nodeID,
	})
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Key == "" {
		http.Error(w, "key cannot be empty", http.StatusBadRequest)
		return
	}

	version := nextVersion(req.Key)
	setLocalRecord(req.Key, req.Value, version)

	ackCount := 1
	for _, peer := range peers {
		ok := sendReplication(peer, req.Key, req.Value, version)
		time.Sleep(200 * time.Millisecond)
		if ok {
			ackCount++
		}
	}

	if ackCount != 5 {
		http.Error(w, fmt.Sprintf("write failed to reach all nodes, only got %d/5", ackCount), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"key":               req.Key,
		"value":             req.Value,
		"version":           version,
		"write_coordinator": nodeID,
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/get/")
	rec, ok := getLocalRecord(key)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, GetResponse{
		Key:      key,
		Value:    rec.Value,
		Version:  rec.Version,
		ServedBy: nodeID,
	})
}

func localReadHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/local_read/")
	rec, ok := getLocalRecord(key)
	if !ok {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, GetResponse{
		Key:      key,
		Value:    rec.Value,
		Version:  rec.Version,
		ServedBy: nodeID,
	})
}

func internalReplicateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req ReplicateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	time.Sleep(100 * time.Millisecond)
	setLocalRecord(req.Key, req.Value, req.Version)
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "replicated",
		"node_id": nodeID,
	})
}

func sendReplication(peer, key, value string, version int) bool {
	body, _ := json.Marshal(ReplicateRequest{
		Key:     key,
		Value:   value,
		Version: version,
	})
	resp, err := httpClient.Post(peer+"/internal/replicate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.ReadAll(resp.Body)
	return resp.StatusCode == http.StatusOK
}

func getLocalRecord(key string) (Record, bool) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	rec, ok := store[key]
	return rec, ok
}

func setLocalRecord(key, value string, version int) {
	storeMu.Lock()
	defer storeMu.Unlock()
	current, ok := store[key]
	if !ok || version >= current.Version {
		store[key] = Record{Value: value, Version: version}
	}
}

func nextVersion(key string) int {
	storeMu.RLock()
	defer storeMu.RUnlock()
	current, ok := store[key]
	if !ok {
		return 1
	}
	return current.Version + 1
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := []string{}
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			out = append(out, strings.TrimSpace(p))
		}
	}
	return out
}