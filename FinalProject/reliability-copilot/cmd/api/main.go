package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"

	"reliability-copilot/internal/model"
	"reliability-copilot/internal/store"
)

func main() {
	region := getenv("AWS_REGION", "us-east-1")
	tableName := getenv("DDB_TABLE_NAME", "reliability-copilot-events")
	queueURL := getenv("SQS_QUEUE_URL", "")

	st, err := store.New(tableName, queueURL, region)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	mux.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var event model.FailureEvent
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}

		if event.EventID == "" {
			event.EventID = uuid.NewString()
		}

		if err := st.Enqueue(event); err != nil {
			http.Error(w, "failed to enqueue event", http.StatusInternalServerError)
			return
		}

		writeJSON(w, http.StatusAccepted, map[string]any{
			"event_id": event.EventID,
			"status":   "queued",
		})
	})

	mux.HandleFunc("/results/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		eventID := strings.TrimPrefix(r.URL.Path, "/results/")
		result, err := st.GetResult(eventID)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, result)
	})

	port := getenv("PORT", "8080")
	log.Printf("API listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
