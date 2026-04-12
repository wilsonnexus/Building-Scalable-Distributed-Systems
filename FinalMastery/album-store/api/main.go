package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	db        *sql.DB
	baseURL   string
	uploadDir string
	publicDir string
	maxMemory int64
}

type Album struct {
	AlbumID     string `json:"album_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Owner       string `json:"owner"`
}

type PhotoAccepted struct {
	PhotoID string `json:"photo_id"`
	Seq     int    `json:"seq"`
	Status  string `json:"status"`
}

type PhotoStatus struct {
	PhotoID string `json:"photo_id"`
	AlbumID string `json:"album_id"`
	Seq     int    `json:"seq"`
	Status  string `json:"status"`
	URL     string `json:"url,omitempty"`
}

type ErrorResp struct {
	Error string `json:"error"`
}

func main() {
	dbPath := getenv("DB_PATH", "/app/data/app.db")
	baseURL := getenv("PUBLIC_BASE_URL", "http://localhost")
	uploadDir := getenv("UPLOAD_DIR", "/app/uploads")
	publicDir := getenv("PUBLIC_DIR", "/app/public")

	_ = os.MkdirAll(filepath.Dir(dbPath), 0755)
	_ = os.MkdirAll(uploadDir, 0755)
	_ = os.MkdirAll(publicDir, 0755)

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(1)

	if err := initDB(db); err != nil {
		log.Fatal(err)
	}

	app := &App{
		db:        db,
		baseURL:   strings.TrimRight(baseURL, "/"),
		uploadDir: uploadDir,
		publicDir: publicDir,
		maxMemory: 32 << 20,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", app.health)
	mux.HandleFunc("/albums", app.albumsRoot)
	mux.HandleFunc("/albums/", app.albumsSubroutes)
	mux.HandleFunc("/media/", app.media)

	addr := ":8080"
	log.Printf("album-store api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, logging(mux)))
}

func initDB(db *sql.DB) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS albums (
			album_id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL,
			owner TEXT NOT NULL,
			next_seq INTEGER NOT NULL DEFAULT 0
		);`,
		`CREATE TABLE IF NOT EXISTS photos (
			photo_id TEXT PRIMARY KEY,
			album_id TEXT NOT NULL,
			seq INTEGER NOT NULL,
			status TEXT NOT NULL,
			url TEXT,
			upload_path TEXT NOT NULL,
			public_path TEXT,
			created_at TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_photos_album_id ON photos(album_id);`,
		`CREATE TABLE IF NOT EXISTS jobs (
			job_id INTEGER PRIMARY KEY AUTOINCREMENT,
			photo_id TEXT UNIQUE NOT NULL,
			album_id TEXT NOT NULL,
			upload_path TEXT NOT NULL,
			public_path TEXT NOT NULL,
			status TEXT NOT NULL,
			attempts INTEGER NOT NULL DEFAULT 0,
			error TEXT
		);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (a *App) albumsRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/albums" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rows, err := a.db.Query(`SELECT album_id, title, description, owner FROM albums`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		albums := []Album{}
		for rows.Next() {
			var al Album
			if err := rows.Scan(&al.AlbumID, &al.Title, &al.Description, &al.Owner); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			albums = append(albums, al)
		}
		writeJSON(w, http.StatusOK, albums)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) albumsSubroutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/albums/")
	parts := strings.Split(path, "/")

	if len(parts) == 1 {
		albumID := parts[0]
		switch r.Method {
		case http.MethodPut:
			a.putAlbum(w, r, albumID)
			return
		case http.MethodGet:
			a.getAlbum(w, r, albumID)
			return
		}
	}

	if len(parts) == 2 && parts[1] == "photos" && r.Method == http.MethodPost {
		a.uploadPhoto(w, r, parts[0])
		return
	}

	if len(parts) == 3 && parts[1] == "photos" {
		albumID := parts[0]
		photoID := parts[2]
		switch r.Method {
		case http.MethodGet:
			a.getPhoto(w, r, albumID, photoID)
			return
		case http.MethodDelete:
			a.deletePhoto(w, r, albumID, photoID)
			return
		}
	}

	writeError(w, http.StatusNotFound, "not found")
}

func (a *App) putAlbum(w http.ResponseWriter, r *http.Request, albumID string) {
	var req Album

	decErr := json.NewDecoder(r.Body).Decode(&req)
	if decErr != nil && !errors.Is(decErr, io.EOF) {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}

	// If album_id is omitted in the body, use the path parameter.
	if req.AlbumID == "" {
		req.AlbumID = albumID
	}

	// Reject only if the body album_id is present but mismatches the path.
	if req.AlbumID != albumID {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}

	var existing Album
	err := a.db.QueryRow(
		`SELECT album_id, title, description, owner FROM albums WHERE album_id = ?`,
		albumID,
	).Scan(&existing.AlbumID, &existing.Title, &existing.Description, &existing.Owner)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	// If the album already exists, keep old values for any missing fields.
	if err == nil {
		if req.Title == "" {
			req.Title = existing.Title
		}
		if req.Description == "" {
			req.Description = existing.Description
		}
		if req.Owner == "" {
			req.Owner = existing.Owner
		}

		_, err = a.db.Exec(
			`UPDATE albums SET title = ?, description = ?, owner = ? WHERE album_id = ?`,
			req.Title, req.Description, req.Owner, req.AlbumID,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		writeJSON(w, http.StatusOK, req)
		return
	}

	// For new albums, be permissive if fields are omitted.
	if req.Title == "" {
		req.Title = "untitled"
	}
	if req.Description == "" {
		req.Description = ""
	}
	if req.Owner == "" {
		req.Owner = ""
	}

	_, err = a.db.Exec(
		`INSERT INTO albums(album_id, title, description, owner, next_seq) VALUES (?, ?, ?, ?, 0)`,
		req.AlbumID, req.Title, req.Description, req.Owner,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusCreated, req)
}

func (a *App) getAlbum(w http.ResponseWriter, r *http.Request, albumID string) {
	var al Album
	err := a.db.QueryRow(
		`SELECT album_id, title, description, owner FROM albums WHERE album_id = ?`,
		albumID,
	).Scan(&al.AlbumID, &al.Title, &al.Description, &al.Owner)

	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, al)
}

func (a *App) uploadPhoto(w http.ResponseWriter, r *http.Request, albumID string) {
	if !albumExists(a.db, albumID) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	if err := r.ParseMultipartForm(a.maxMemory); err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		writeError(w, http.StatusBadRequest, "bad request")
		return
	}
	defer file.Close()

	photoID := uuid.NewString()
	uploadPath := filepath.Join(a.uploadDir, photoID+"_"+sanitizeFilename(header))
	publicPath := filepath.Join(a.publicDir, photoID)

	if err := saveUploadedFile(file, uploadPath); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	tx, err := a.db.Begin()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE albums SET next_seq = next_seq + 1 WHERE album_id = ?`, albumID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	var seq int
	err = tx.QueryRow(`SELECT next_seq FROM albums WHERE album_id = ?`, albumID).Scan(&seq)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	_, err = tx.Exec(
		`INSERT INTO photos(photo_id, album_id, seq, status, url, upload_path, public_path, created_at)
		 VALUES (?, ?, ?, 'processing', '', ?, '', ?)`,
		photoID, albumID, seq, uploadPath, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	_, err = tx.Exec(
		`INSERT INTO jobs(photo_id, album_id, upload_path, public_path, status, attempts, error)
		 VALUES (?, ?, ?, ?, 'queued', 0, '')`,
		photoID, albumID, uploadPath, publicPath,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusAccepted, PhotoAccepted{
		PhotoID: photoID,
		Seq:     seq,
		Status:  "processing",
	})
}

func (a *App) getPhoto(w http.ResponseWriter, r *http.Request, albumID, photoID string) {
	var ps PhotoStatus
	err := a.db.QueryRow(
		`SELECT photo_id, album_id, seq, status, url FROM photos WHERE photo_id = ? AND album_id = ?`,
		photoID, albumID,
	).Scan(&ps.PhotoID, &ps.AlbumID, &ps.Seq, &ps.Status, &ps.URL)

	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, ps)
}

func (a *App) deletePhoto(w http.ResponseWriter, r *http.Request, albumID, photoID string) {
	var uploadPath, publicPath string
	err := a.db.QueryRow(
		`SELECT upload_path, public_path FROM photos WHERE photo_id = ? AND album_id = ?`,
		photoID, albumID,
	).Scan(&uploadPath, &publicPath)

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}

	_ = removeIfExists(uploadPath)
	_ = removeIfExists(publicPath)

	tx, err := a.db.Begin()
	if err == nil {
		_, _ = tx.Exec(`DELETE FROM jobs WHERE photo_id = ?`, photoID)
		_, _ = tx.Exec(`DELETE FROM photos WHERE photo_id = ? AND album_id = ?`, photoID, albumID)
		_ = tx.Commit()
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *App) media(w http.ResponseWriter, r *http.Request) {
	photoID := strings.TrimPrefix(r.URL.Path, "/media/")
	if photoID == "" {
		http.NotFound(w, r)
		return
	}
	path := filepath.Join(a.publicDir, photoID)
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}

func albumExists(db *sql.DB, albumID string) bool {
	var count int
	_ = db.QueryRow(`SELECT COUNT(1) FROM albums WHERE album_id = ?`, albumID).Scan(&count)
	return count > 0
}

func saveUploadedFile(src multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, src)
	return err
}

func removeIfExists(path string) error {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	return os.Remove(path)
}

func sanitizeFilename(h *multipart.FileHeader) string {
	name := filepath.Base(h.Filename)
	if name == "" || name == "." || name == "/" {
		return "upload.bin"
	}
	return strings.ReplaceAll(name, " ", "_")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResp{Error: msg})
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func getenv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
