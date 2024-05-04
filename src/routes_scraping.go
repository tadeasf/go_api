package main

import (
	"database/sql"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gocolly/colly/v2"
)

func initializeScrapingRoutes(router *gin.Engine) {
    router.GET("/start-scraping", startScrapingHandler)
    router.GET("/articles", getArticlesHandler)
}

func startScrapingHandler(c *gin.Context) {
    go scrapeSources()
    c.JSON(http.StatusOK, gin.H{"message": "Scraping started. Check logs for progress."})
}

func scrapeSources() {
    db, err := getDbConnection()
    if err != nil {
        log.Printf("Failed to connect to the database: %v\n", err)
        return
    }
    defer db.Close()

    var sources []struct {
        ID      int
        RootURL string
    }
    rows, err := db.Query("SELECT id, root_url FROM news_sources")
    if err != nil {
        log.Printf("Failed to query sources: %v\n", err)
        return
    }
    defer rows.Close()

    for rows.Next() {
        var s struct {
            ID      int
            RootURL string
        }
        if err := rows.Scan(&s.ID, &s.RootURL); err != nil {
            log.Printf("Failed to scan source: %v\n", err)
            continue
        }
        sources = append(sources, s)
    }

    var wg sync.WaitGroup
    for _, source := range sources {
        wg.Add(1)
        go func(source struct{ ID int; RootURL string }) {
            defer wg.Done()
            scrapeSource(source.ID, source.RootURL, db)
        }(source)
    }
    wg.Wait()
}

func scrapeSource(sourceID int, url string, db *sql.DB) {
    c := colly.NewCollector()

    c.OnHTML("article", func(e *colly.HTMLElement) {
        title := e.ChildText("h1, h2, h3")
        author := e.ChildText(".author")
        content := e.ChildText("p")
        articleURL := e.Request.URL.String() // Assuming the article's URL is the page URL

        if title != "" {
            insertArticle(db, sourceID, title, author, content, articleURL)
        }
    })

    c.Visit(url)
}

func insertArticle(db *sql.DB, sourceID int, title, author, content, url string) {
    _, err := db.Exec("INSERT INTO articles (scraped_at, title, author, content, source_id, url) VALUES ($1, $2, $3, $4, $5, $6)", time.Now(), title, author, content, sourceID, url)
    if err != nil {
        log.Printf("Failed to insert article: %v\n", err)
    }
}

func getArticlesHandler(c *gin.Context) {
    db, err := getDbConnection()
    if err != nil {
        log.Printf("Failed to connect to the database: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }
    defer db.Close()

    var articles []struct {
        ID        int    `json:"id"`
        ScrapedAt string `json:"scraped_at"`
        Title     string `json:"title"`
        Author    string `json:"author"`
        Content   string `json:"content"`
        URL       string `json:"url"`
    }

    rows, err := db.Query("SELECT id, scraped_at, title, author, content, url FROM articles")
    if err != nil {
        log.Printf("Failed to query articles: %v\n", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query articles"})
        return
    }
    defer rows.Close()

    for rows.Next() {
        var a struct {
            ID        int    `json:"id"`
            ScrapedAt string `json:"scraped_at"`
            Title     string `json:"title"`
            Author    string `json:"author"`
            Content   string `json:"content"`
            URL       string `json:"url"`
        }
        err := rows.Scan(&a.ID, &a.ScrapedAt, &a.Title, &a.Author, &a.Content, &a.URL)
        if err != nil {
            log.Printf("Failed to scan article: %v\n", err)
            continue
        }
        articles = append(articles, a)
    }

    c.JSON(http.StatusOK, articles)
}