package main

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"os"
	"strings"
	"time"
)

type FailureEvent struct {
	PipelineID string `json:"pipeline_id"`
	LogText    string `json:"log_text"`
	Source     string `json:"source"`
	Timestamp  string `json:"timestamp"`
}

type AnalysisResponse struct {
	FailureCategory string  `json:"failure_category"`
	Recommendation  string  `json:"recommendation"`
	RiskScore       float64 `json:"risk_score"`
	ProcessingMode  string  `json:"processing_mode"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/analyze", analyzeHandler)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	log.Printf("API listening on :%s", port)
	log.Fatal(server.ListenAndServe())
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func analyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event FailureEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	category, recommendation, score := classify(event.LogText)

	resp := AnalysisResponse{
		FailureCategory: category,
		Recommendation:  recommendation,
		RiskScore:       score,
		ProcessingMode:  "rule-based-baseline",
	}

	writeJSON(w, http.StatusOK, resp)
}

func classify(logText string) (string, string, float64) {
	text := strings.ToLower(logText)

	switch {
	case strings.Contains(text, "timeout"):
		return "timeout", "Check service latency, retry settings, and downstream availability.", 0.82
	case strings.Contains(text, "connection refused") || strings.Contains(text, "dns") || strings.Contains(text, "network"):
		return "infrastructure_issue", "Check networking, DNS, service discovery, and container health.", 0.78
	case strings.Contains(text, "dependency") || strings.Contains(text, "package") || strings.Contains(text, "artifact"):
		return "dependency_failure", "Verify dependency versions, artifact availability, and external service health.", 0.74
	case strings.Contains(text, "flaky") || strings.Contains(text, "intermittent"):
		return "flaky_test", "Rerun the job, isolate unstable tests, and review timing assumptions.", 0.67
	default:
		base := 0.55 + 0.1*math.Sin(float64(time.Now().UnixNano()%1000))
		if base < 0.5 {
			base = 0.5
		}
		if base > 0.75 {
			base = 0.75
		}
		return "unknown_failure", "Review recent logs and route for manual triage or LLM-assisted summarization.", base
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
