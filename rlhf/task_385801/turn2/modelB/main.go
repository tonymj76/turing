package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite" // Replace with your preferred DB dialect
)

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuditLog struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	RemoteIP  string    `json:"remote_ip"`
	UserAgent string    `json:"user_agent"`
	UserID    uint      `json:"user_id"`
	Username  string    `json:"username"`
}

// Sample in-memory authorization rules
var authRules = map[string][]string{
	"/admin": {"admin"},
	"/data":  {"user", "admin"},
}

func authenticate(r *http.Request) (uint, string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return 0, "", fmt.Errorf("token required")
	}

	// Simplified token validation for demonstration purposes
	// In a real application, use a proper JWT library for validation
	parts := strings.SplitN(token, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return 0, "", fmt.Errorf("invalid token format")
	}

	// Extract username from the token (in a real app, this would be from the decoded JWT)
	username := parts[1]

	db, err := gorm.Open("sqlite3", ":memory:") // Replace with your database connection string
	if err != nil {
		return 0, "", err
	}
	defer db.Close()

	var user User
	if err := db.First(&user, "username = ?", username).Error; err != nil {
		return 0, "", fmt.Errorf("user not found")
	}

	return user.ID, user.Username, nil
}

func authorize(path string, userID uint, username string) bool {
	allowedRoles, ok := authRules[path]
	if !ok {
		return false // No rule defined for this path, deny by default
	}

	if username == "admin" {
		return true // Admin has access to all
	}

	for _, role := range allowedRoles {
		if role == "user" {
			return true // User has access to specified paths
		}
	}

	return false // User denied access
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		userID, username, err := authenticate(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if !authorize(r.URL.Path, userID, username) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		// Handle the request
		next.ServeHTTP(w, r)

		logEntry := &AuditLog{
			Timestamp: start,
			Method:    r.Method,
			Path:      r.URL.Path,
			RemoteIP:  strings.SplitN(r.RemoteAddr, ":", 2)[0],
			UserAgent: r.UserAgent(),
			UserID:    userID,
			Username:  username,
		}

		logAudit(logEntry)
	})
}

func logAudit(logEntry *AuditLog) {
	b, err := json.Marshal(logEntry)
	if err != nil {
		log.Printf("Error marshaling audit log: %v", err)
		return
	}
	log.Println(string(b))
}

func helloWorld(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, world!"))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/", helloWorld).Methods("GET")
	r.HandleFunc("/admin", helloWorld).Methods("GET")
	r.HandleFunc("/data", helloWorld).Methods("GET")

	// Wrap the router with the logging middleware
	loggedRouter := LoggingMiddleware(r)

	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
