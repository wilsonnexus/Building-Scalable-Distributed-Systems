package model

type FailureEvent struct {
	EventID    string `json:"event_id"`
	PipelineID string `json:"pipeline_id"`
	LogText    string `json:"log_text"`
	Source     string `json:"source"`
	Timestamp  string `json:"timestamp"`
}

type AnalysisResult struct {
	EventID         string  `json:"event_id"`
	FailureCategory string  `json:"failure_category"`
	Recommendation  string  `json:"recommendation"`
	RiskScore       float64 `json:"risk_score"`
	Status          string  `json:"status"`
	ProcessingMode  string  `json:"processing_mode"`
	ErrorMessage    string  `json:"error_message,omitempty"`
}
