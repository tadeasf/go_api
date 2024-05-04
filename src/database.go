package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	// Make sure to import the Postgres driver
	_ "github.com/lib/pq"
)

// getDbConnection establishes a connection to the database and returns the connection object.
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

// initDB checks for the existence of the 'news_sources' table and creates it if it does not exist.
// initDB checks for the existence of the 'news_sources' and 'articles' tables and creates them if they do not exist.
func initDB() {
    db, err := getDbConnection()
    if err != nil {
        log.Fatalf("Could not connect to the database: %v", err)
    }
    defer db.Close()

    // Create news_sources table if it does not exist
    createNewsSourcesTableSQL := `CREATE TABLE IF NOT EXISTS news_sources (
        id SERIAL PRIMARY KEY,
        root_url VARCHAR(255) NOT NULL,
        name VARCHAR(100) NOT NULL
    );`
    _, err = db.Exec(createNewsSourcesTableSQL)
    if err != nil {
        log.Fatalf("Could not create news_sources table: %v", err)
    }

    // Create articles table if it does not exist
    createArticlesTableSQL := `CREATE TABLE IF NOT EXISTS articles (
        id SERIAL PRIMARY KEY,
        published_at TIMESTAMP WITH TIME ZONE NOT NULL,
        title TEXT NOT NULL,
        author TEXT,  -- Can be empty
        content TEXT NOT NULL,
        source_id INTEGER REFERENCES news_sources(id)  -- Foreign Key to news_sources table
    );`
    _, err = db.Exec(createArticlesTableSQL)
    if err != nil {
        log.Fatalf("Could not create articles table: %v", err)
    }

    fmt.Println("Tables creation/check completed successfully")
}