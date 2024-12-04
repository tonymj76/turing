package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp time.Time `json:"timestamp"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	RemoteIP  string    `json:"remote_ip"`
	UserAgent string    `json:"user_agent"`
	Status    int       `json:"status"`
}

type ResponseWriterWithStatus struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *ResponseWriterWithStatus) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware is a middleware to log HTTP requests
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap the ResponseWriter
		rw := &ResponseWriterWithStatus{
			ResponseWriter: w,
			StatusCode:     http.StatusOK, // Default to 200 (OK) if WriteHeader is not called
		}

		// Handle the request
		next.ServeHTTP(rw, r)

		// Create an audit log entry
		logEntry := &AuditLog{
			Timestamp: start,
			Method:    r.Method,
			Path:      r.URL.Path,
			RemoteIP:  strings.SplitN(r.RemoteAddr, ":", 2)[0],
			UserAgent: r.UserAgent(),
			Status:    rw.StatusCode,
		}

		// Log the audit entry
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

	// Wrap the router with the logging middleware
	loggedRouter := LoggingMiddleware(r)

	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
