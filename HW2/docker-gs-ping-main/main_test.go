package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouterForTest() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/albums", getAlbums)
	r.GET("/albums/:id", getAlbumByID)
	r.POST("/albums", postAlbums)

	return r
}

// resetAlbums keeps tests independent (because albums is global).
func resetAlbums() {
	albums = []album{
		{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
		{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
		{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
	}
}

func TestGetAlbums(t *testing.T) {
	resetAlbums()
	r := setupRouterForTest()

	req := httptest.NewRequest(http.MethodGet, "/albums", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /albums status = %d, want %d", w.Code, http.StatusOK)
	}
	// Minimal check: ensure we got JSON-ish content with known seeded ID
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte(`"id": "1"`)) {
		t.Fatalf("GET /albums body missing seeded album: %s", body)
	}
}

func TestGetAlbumByIDFound(t *testing.T) {
	resetAlbums()
	r := setupRouterForTest()

	req := httptest.NewRequest(http.MethodGet, "/albums/2", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /albums/2 status = %d, want %d", w.Code, http.StatusOK)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`"id": "2"`)) {
		t.Fatalf("GET /albums/2 body = %s, want id 2", w.Body.String())
	}
}

func TestGetAlbumByIDNotFound(t *testing.T) {
	resetAlbums()
	r := setupRouterForTest()

	req := httptest.NewRequest(http.MethodGet, "/albums/999", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /albums/999 status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPostAlbumsValid(t *testing.T) {
	resetAlbums()
	r := setupRouterForTest()

	payload := []byte(`{"id":"4","title":"The Modern Sound of Betty Carter","artist":"Betty Carter","price":49.99}`)
	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("POST /albums status = %d, want %d. body=%s", w.Code, http.StatusCreated, w.Body.String())
	}

	// Confirm it actually got appended
	if len(albums) != 4 {
		t.Fatalf("albums length = %d, want 4", len(albums))
	}
}

func TestPostAlbumsInvalidJSON(t *testing.T) {
	resetAlbums()
	r := setupRouterForTest()

	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader([]byte(`not-json`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("POST /albums invalid json status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostAlbumsDuplicateID(t *testing.T) {
	resetAlbums()
	r := setupRouterForTest()

	payload := []byte(`{"id":"1","title":"Dup","artist":"X","price":10}`)
	req := httptest.NewRequest(http.MethodPost, "/albums", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("POST /albums duplicate id status = %d, want %d. body=%s", w.Code, http.StatusConflict, w.Body.String())
	}
}
