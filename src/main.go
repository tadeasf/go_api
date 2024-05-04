package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

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

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello World")
}

func main() {
    // Register the helloWorldHandler function to the "/hello" route
    http.HandleFunc("/hello", helloWorldHandler)

    // Start an HTTP server listening on port 8080
    fmt.Println("Server is listening on port 8080...")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        log.Fatal(err)
    }
}