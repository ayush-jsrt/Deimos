package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

// DB connection configuration
const (
	dbUser     = "root"
	dbPassword = "root"
	dbHost     = "crud-mysql.deimos.svc.cluster.local"
	dbName     = "NOTES"
)

func connectDB(retries int, delay time.Duration) *sql.DB {
	var conn *sql.DB
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName)

	for attempt := 1; attempt <= retries; attempt++ {
		conn, err = sql.Open("mysql", dsn)
		if err == nil {
			err = conn.Ping()
			if err == nil {
				log.Println("[DB] Connected successfully")
				return conn
			}
		}
		log.Printf("[DB] Attempt %d/%d failed: %v", attempt, retries, err)
		time.Sleep(delay)
	}

	log.Fatalf("[DB] Could not connect after %d attempts", retries)
	return nil
}

func initDB() {
	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) UNIQUE NOT NULL,
		content TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatalf("[DB] Error initializing table: %v", err)
	}
}

type Note struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func getNotes(c *gin.Context) {
	rows, err := db.Query("SELECT name, content FROM notes")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	notes := []Note{} // ensures [] instead of null
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.Name, &n.Content); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		notes = append(notes, n)
	}
	c.JSON(http.StatusOK, notes)
}

func createNote(c *gin.Context) {
	var n Note
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	_, err := db.Exec("INSERT INTO notes (name, content) VALUES (?, ?)", n.Name, n.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Note created successfully"})
}

func updateNote(c *gin.Context) {
	name := c.Param("name")
	var n Note
	if err := c.ShouldBindJSON(&n); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}
	_, err := db.Exec("UPDATE notes SET content=? WHERE name=?", n.Content, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Note updated successfully"})
}

func deleteNote(c *gin.Context) {
	name := c.Param("name")
	_, err := db.Exec("DELETE FROM notes WHERE name=?", name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Note deleted successfully"})
}

func main() {
	db = connectDB(30, 2*time.Second)
	initDB()

	r := gin.Default()

	// Enable CORS for all origins
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept"},
	}))

	// Routes
	r.GET("/notes", getNotes)
	r.POST("/notes", createNote)
	r.PUT("/notes/:name", updateNote)
	r.DELETE("/notes/:name", deleteNote)

	log.Println("[Server] Running on port 5000")
	r.Run(":5000") // listen on port 5000
}
