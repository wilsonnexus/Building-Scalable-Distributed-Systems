package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums seeds record album data (in-memory for this assignment).
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

func main() {
	router := gin.Default()

	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	// Local only: listen on your machine at localhost:8080
	if err := router.Run(":8080"); err != nil {
		// In practice you'd log.Fatalf; keeping it simple.
		panic(err)
	}
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// BindJSON parses and validates JSON format. If it fails, respond clearly.
	if err := c.BindJSON(&newAlbum); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "invalid JSON body"})
		return
	}

	// Minimal validation for maintainability + demo readiness.
	if newAlbum.ID == "" || newAlbum.Title == "" || newAlbum.Artist == "" || newAlbum.Price <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{
			"message": "missing/invalid fields: id, title, artist must be non-empty; price must be > 0",
		})
		return
	}

	// Avoid duplicate IDs (simple check since we're in-memory).
	for _, a := range albums {
		if a.ID == newAlbum.ID {
			c.IndentedJSON(http.StatusConflict, gin.H{"message": "album with that id already exists"})
			return
		}
	}

	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID matches the id parameter.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}
