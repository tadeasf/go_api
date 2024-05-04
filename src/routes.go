package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// initializeRoutes sets up the web routes.
func initializeRoutes(router *gin.Engine) {
    router.GET("/hello", func(c *gin.Context) {
        c.String(http.StatusOK, "Hello World")
    })

    router.GET("/dbcheck", dbCheckHandler)
    router.GET("/news-sources", listNewsSources)
    router.POST("/news-sources", addNewsSource) // New route for adding a source
    router.DELETE("/news-sources", removeNewsSource) // New route for removing a source
}

// dbCheckHandler checks the database connection.
func dbCheckHandler(c *gin.Context) {
	db, err := getDbConnection()
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to connect to database")
		return
	}
	defer db.Close()
	c.String(http.StatusOK, "Successfully connected to database")
}

// listNewsSources fetches and returns all sources from the news_sources table.
func listNewsSources(c *gin.Context) {
	db, err := getDbConnection()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
		return
	}
	defer db.Close()

	var sources []struct {
		ID      int    `json:"id"`
		RootURL string `json:"root_url"`
		Name    string `json:"name"`
	}
	rows, err := db.Query("SELECT id, root_url, name FROM news_sources")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query news sources"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var s struct {
			ID      int    `json:"id"`
			RootURL string `json:"root_url"`
			Name    string `json:"name"`
		}
		if err := rows.Scan(&s.ID, &s.RootURL, &s.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan news sources"})
			return
		}
		sources = append(sources, s)
	}

	c.JSON(http.StatusOK, sources)
}

// addNewsSource adds a new news source to the database.
func addNewsSource(c *gin.Context) {
    var source struct {
        RootURL string `json:"root_url" binding:"required"`
        Name    string `json:"name" binding:"required"`
    }

    if err := c.ShouldBindJSON(&source); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    db, err := getDbConnection()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }
    defer db.Close()

    insertSQL := `INSERT INTO news_sources (root_url, name) VALUES ($1, $2)`
    _, err = db.Exec(insertSQL, source.RootURL, source.Name)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert new source"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "News source added successfully"})
}

// removeNewsSource removes a news source from the database by its URL.
func removeNewsSource(c *gin.Context) {
    rootURL := c.Query("root_url")
    if rootURL == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "root_url is required"})
        return
    }

    db, err := getDbConnection()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to database"})
        return
    }
    defer db.Close()

    deleteSQL := `DELETE FROM news_sources WHERE root_url = $1`
    res, err := db.Exec(deleteSQL, rootURL)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete source"})
        return
    }

    count, err := res.RowsAffected()
    if err != nil || count == 0 {
        c.JSON(http.StatusNotFound, gin.H{"error": "Source not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "News source removed successfully"})
}