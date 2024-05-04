package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"strings"

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

    // Assuming BBC News is the source we want to scrape
    scrapeBBCNews(db)
}

func scrapeBBCNews(db *sql.DB) {
    sourceID := 1 // Assuming BBC News has ID 1 in your `news_sources` table
    rootURL := "https://www.bbc.com/news"

    c := colly.NewCollector()

    // Find and visit all links
    c.OnHTML("a[href]", func(e *colly.HTMLElement) {
        link := e.Attr("href")
        if strings.HasPrefix(link, "/news/") {
            articleURL := "https://www.bbc.com" + link
            go visitArticlePage(sourceID, articleURL, db)
        }
    })

    c.Visit(rootURL)
}

func visitArticlePage(sourceID int, articleURL string, db *sql.DB) {
    articleCollector := colly.NewCollector()

    var title, author, content string

    articleCollector.OnHTML("h1.sc-82e6a0ec-0.fxXQuy", func(e *colly.HTMLElement) {
        title = e.Text
    })

    articleCollector.OnHTML("span[data-testid='byline-name'].sc-53757630-5.jMdGt", func(e *colly.HTMLElement) {
        author = e.Text
    })

    articleCollector.OnHTML("p.sc-e1853509-0.bmLndb", func(e *colly.HTMLElement) {
        content += e.Text + "\n"
    })

    articleCollector.OnScraped(func(r *colly.Response) {
        if title != "" {
            insertArticle(db, sourceID, title, author, content, articleURL)
        }
    })

    articleCollector.Visit(articleURL)
}

func insertArticle(db *sql.DB, sourceID int, title, author, content, url string) {
    _, err := db.Exec("INSERT INTO articles (scraped_at, title, author, content, source_id, url) VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (url) DO NOTHING", time.Now(), title, author, content, sourceID, url)
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