package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
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
	Key            string `json:"key"`
	Value          string `json:"value"`
	Version        int    `json:"version"`
	ServedBy       string `json:"served_by"`
	ReadQuorumSize int    `json:"read_quorum_size,omitempty"`
}

var (
	nodeID    = getenv("NODE_ID", "node1")
	role      = getenv("ROLE", "leader")
	leaderURL = getenv("LEADER_URL", "http://node1:8000")
	followers = splitCSV(os.Getenv("FOLLOWERS"))
	defaultW  = atoi(getenv("DEFAULT_W", "5"))
	defaultR  = atoi(getenv("DEFAULT_R", "1"))

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
	mux.HandleFunc("/internal/read/", internalReadHandler)

	addr := ":8000"
	fmt.Printf("starting leader_follower node=%s role=%s on %s\n", nodeID, role, addr)
	http.ListenAndServe(addr, mux)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"node_id": nodeID,
		"role":    role,
	})
}

func setHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if role != "leader" {
		http.Error(w, "writes must go to the leader", http.StatusBadRequest)
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

	W := defaultW
	if rawW := r.URL.Query().Get("w"); rawW != "" {
		W = atoi(rawW)
	}

	version := nextVersion(req.Key)
	setLocalRecord(req.Key, req.Value, version)

	if W == 1 {
		go replicateToFollowers(req.Key, req.Value, version, 0)
		writeJSON(w, http.StatusCreated, map[string]any{
			"key":            req.Key,
			"value":          req.Value,
			"version":        version,
			"acknowledged_w": 1,
		})
		return
	}

	ackCount := 1
	lastIndex := 0

	for idx, follower := range followers {
		ok := sendReplication(follower, req.Key, req.Value, version)
		time.Sleep(200 * time.Millisecond)
		if ok {
			ackCount++
			lastIndex = idx + 1
		}
		if ackCount >= W {
			break
		}
	}

	if ackCount < W {
		http.Error(w, fmt.Sprintf("write failed to reach W=%d acknowledgements", W), http.StatusInternalServerError)
		return
	}

	if lastIndex < len(followers) {
		go replicateToFollowers(req.Key, req.Value, version, lastIndex)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"key":            req.Key,
		"value":          req.Value,
		"version":        version,
		"acknowledged_w": ackCount,
	})
}

func getHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, "/get/")
	if key == "" {
		http.Error(w, "missing key", http.StatusBadRequest)
		return
	}

	R := defaultR
	if rawR := r.URL.Query().Get("r"); rawR != "" {
		R = atoi(rawR)
	}

	if role != "leader" {
		resp, err := httpClient.Get(fmt.Sprintf("%s/get/%s?r=%d", leaderURL, key, R))
		if err != nil {
			http.Error(w, "failed to proxy read to leader", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
		return
	}

	if R == 1 {
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
		return
	}

	records := []GetResponse{}
	if rec, ok := getLocalRecord(key); ok {
		records = append(records, GetResponse{
			Key:      key,
			Value:    rec.Value,
			Version:  rec.Version,
			ServedBy: nodeID,
		})
	}

	followersNeeded := R - 1
	if followersNeeded > len(followers) {
		followersNeeded = len(followers)
	}

	for _, follower := range followers[:followersNeeded] {
		resp, err := httpClient.Get(fmt.Sprintf("%s/internal/read/%s", follower, key))
		if err != nil {
			continue
		}
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			continue
		}
		var gr GetResponse
		err = json.NewDecoder(resp.Body).Decode(&gr)
		resp.Body.Close()
		if err == nil {
			records = append(records, gr)
		}
	}

	if len(records) == 0 {
		http.Error(w, "key not found", http.StatusNotFound)
		return
	}

	newest := records[0]
	for _, rec := range records[1:] {
		if rec.Version > newest.Version {
			newest = rec
		}
	}
	newest.ReadQuorumSize = len(records)
	writeJSON(w, http.StatusOK, newest)
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

func internalReadHandler(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/internal/read/")
	time.Sleep(50 * time.Millisecond)
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

func sendReplication(follower, key, value string, version int) bool {
	body, _ := json.Marshal(ReplicateRequest{
		Key:     key,
		Value:   value,
		Version: version,
	})
	resp, err := httpClient.Post(follower+"/internal/replicate", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func replicateToFollowers(key, value string, version int, startIndex int) {
	for _, follower := range followers[startIndex:] {
		sendReplication(follower, key, value, version)
		time.Sleep(200 * time.Millisecond)
	}
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

func atoi(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
