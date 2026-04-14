package store

import (
	"database/sql"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"reliability-copilot/internal/model"
)

type Store struct {
	DB *sql.DB
}

func New(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	s := &Store{DB: db}
	if err := s.Init(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Init() error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS events (
			event_id TEXT PRIMARY KEY,
			pipeline_id TEXT,
			log_text TEXT,
			source TEXT,
			ts TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS jobs (
			event_id TEXT PRIMARY KEY,
			status TEXT,
			created_at TEXT,
			updated_at TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS results (
			event_id TEXT PRIMARY KEY,
			failure_category TEXT,
			recommendation TEXT,
			risk_score REAL,
			status TEXT,
			processing_mode TEXT,
			error_message TEXT
		);`,
	}
	for _, stmt := range stmts {
		if _, err := s.DB.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Enqueue(e model.FailureEvent) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`INSERT OR REPLACE INTO events(event_id, pipeline_id, log_text, source, ts) VALUES (?, ?, ?, ?, ?)`,
		e.EventID, e.PipelineID, e.LogText, e.Source, e.Timestamp,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT OR REPLACE INTO jobs(event_id, status, created_at, updated_at) VALUES (?, 'queued', ?, ?)`,
		e.EventID, now, now,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`INSERT OR REPLACE INTO results(event_id, failure_category, recommendation, risk_score, status, processing_mode, error_message)
		 VALUES (?, '', '', 0, 'queued', '', '')`,
		e.EventID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Store) ClaimNextJob() (*model.FailureEvent, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	row := tx.QueryRow(`SELECT event_id FROM jobs WHERE status = 'queued' ORDER BY created_at LIMIT 1`)
	var eventID string
	if err := row.Scan(&eventID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = tx.Exec(`UPDATE jobs SET status = 'processing', updated_at = ? WHERE event_id = ?`, now, eventID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec(`UPDATE results SET status = 'processing' WHERE event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}

	row = tx.QueryRow(`SELECT event_id, pipeline_id, log_text, source, ts FROM events WHERE event_id = ?`, eventID)
	var e model.FailureEvent
	if err := row.Scan(&e.EventID, &e.PipelineID, &e.LogText, &e.Source, &e.Timestamp); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *Store) Complete(eventID, category, recommendation string, risk float64, mode string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE jobs SET status = 'completed', updated_at = ? WHERE event_id = ?`, now, eventID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`UPDATE results SET failure_category = ?, recommendation = ?, risk_score = ?, status = 'completed', processing_mode = ?, error_message = '' WHERE event_id = ?`,
		category, recommendation, risk, mode, eventID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) Fail(eventID, msg, mode string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE jobs SET status = 'failed', updated_at = ? WHERE event_id = ?`, now, eventID)
	if err != nil {
		return err
	}
	_, err = tx.Exec(
		`UPDATE results SET status = 'failed', processing_mode = ?, error_message = ? WHERE event_id = ?`,
		mode, msg, eventID,
	)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) GetResult(eventID string) (*model.AnalysisResult, error) {
	row := s.DB.QueryRow(
		`SELECT event_id, failure_category, recommendation, risk_score, status, processing_mode, error_message
		 FROM results WHERE event_id = ?`,
		eventID,
	)

	var r model.AnalysisResult
	if err := row.Scan(&r.EventID, &r.FailureCategory, &r.Recommendation, &r.RiskScore, &r.Status, &r.ProcessingMode, &r.ErrorMessage); err != nil {
		return nil, err
	}
	return &r, nil
}
