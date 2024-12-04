package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type auditEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IP        string    `json:"ip"`
	User      string    `json:"user"`
	Status    int       `json:"status"`
}

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		// Process request
		c.Next()

		user, exist := c.Get("user")
		if !exist {
			user = "some User"
		}
		entry := auditEntry{
			Timestamp: start,
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			IP:        c.ClientIP(),
			User:      user.(string), // Get the user from the context
			Status:    c.Writer.Status(),
		}

		// Write audit entry to log file or other storage
		auditLog(entry)
	}
}

func auditLog(entry auditEntry) {
	fmt.Println("Audit Log:", entry)
	log.Printf("%s %s %s %s %s %d", entry.Timestamp.Format(time.RFC3339), entry.Method, entry.Path, entry.IP, entry.User, entry.Status)
}

func main() {
	router := gin.Default()
	router.Use(AuditMiddleware())
	// Initialize user authentication here
	router.GET("/", func(c *gin.Context) {
		c.Set("user", "testUser") // Set a test user
		c.String(200, "Hello, World!")
	})
	router.Run(":8080")
}
