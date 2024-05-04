package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    }
}

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func getDbConnection() (*sql.DB, error) {
	// Retrieve environment variables
	psqlUser := os.Getenv("PSQL_USER")
	psqlPassword := os.Getenv("PSQL_PASSWORD")
	psqlHost := os.Getenv("PSQL_HOST")
	psqlDB := os.Getenv("PSQL_DB")
	psqlPort := os.Getenv("PSQL_PORT")

	// Construct the connection string
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		psqlHost, psqlPort, psqlUser, psqlPassword, psqlDB)

	// Open the connection
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// Check the connection
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func main() {
	r := gin.Default()
    r.Use(CORSMiddleware())
    
	r.GET("/hello", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World")
	})

	r.GET("/dbcheck", func(c *gin.Context) {
		db, err := getDbConnection()
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to connect to database")
			log.Println("Failed to connect to database:", err)
			return
		}
		defer db.Close()
		c.String(http.StatusOK, "Successfully connected to database")
	})

	// Start an HTTP server listening on port 8080
	fmt.Println("Server is listening on port 8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}