package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

type album struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("postgres", "host=localhost port=5432 dbname=demo user=postgres password=14022014 sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)
	router.PUT("/albums/:id", updateAlbums)
	router.DELETE("/albums/:id", deleteAlbum)

	router.Run("localhost:8080")
}

func getAlbums(c *gin.Context) {
	rows, err := db.Query("SELECT id, title, artist, price FROM albums")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var albums []album
	for rows.Next() {
		var alb album
		if err := rows.Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		albums = append(albums, alb)
	}
	c.JSON(http.StatusOK, albums)
}

func getAlbumByID(c *gin.Context) {
    id := c.Param("id")

    var alb album
    err := db.QueryRow("SELECT id, title, artist, price FROM albums WHERE id = $1", id).Scan(&alb.ID, &alb.Title, &alb.Artist, &alb.Price)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
            log.Println("Error retrieving album:", err)
        }
        return
    }

    c.JSON(http.StatusOK, alb)
}


func postAlbums(c *gin.Context) {
	var newAlbum album
	if err := c.BindJSON(&newAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	result, err := db.Exec("INSERT INTO albums (title, artist, price) VALUES ($1, $2, $3)", newAlbum.Title, newAlbum.Artist, newAlbum.Price)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	newAlbum.ID = int(newID)
	c.JSON(http.StatusCreated, newAlbum)
}

func updateAlbums(c *gin.Context) {
	id := c.Param("id")

	var updatedAlbum album
	if err := c.BindJSON(&updatedAlbum); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	_, err := db.Exec("UPDATE albums SET title = $1, artist = $2, price = $3 WHERE id = $4", updatedAlbum.Title, updatedAlbum.Artist, updatedAlbum.Price, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updatedAlbum)
}

func deleteAlbum(c *gin.Context) {
	id := c.Param("id")

	_, err := db.Exec("DELETE FROM albums WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
