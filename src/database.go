package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Make sure to import the Postgres driver
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

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
    db.SetMaxOpenConns(25) // Set max. number of open connections to the database.
    db.SetMaxIdleConns(25) // Set max. number of connections in the idle connection pool.
    db.SetConnMaxLifetime(time.Minute * 5) // Set the max. amount of time a connection may be reused.

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

func initDB() {
    db, err := getDbConnection()
    if err != nil {
        log.Fatalf("Could not connect to the database: %v", err)
    }
    defer db.Close()

    // Create the tables

    // Create the users table

    fmt.Println("Tables creation/check completed successfully")
}

func dbCheckHandler(c *gin.Context) {
    db, err := getDbConnection()
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed to connect to database")
        return
    }
    defer db.Close()

    // Simple check if we can retrieve something from the database
    var count int
    err = db.QueryRow("SELECT 1").Scan(&count)
    if err != nil {
        c.String(http.StatusInternalServerError, "Database error")
        return
    }

    c.String(http.StatusOK, "Successfully connected to database")
}

func initializeDatabaseRoutes(r *gin.Engine) {
    r.GET("/db-check", dbCheckHandler)
}