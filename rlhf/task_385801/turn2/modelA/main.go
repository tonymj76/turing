package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
)

// User represents a user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	Timestamp time.Time `json:"timestamp"`
	UserID    string    `json:"user_id"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	RemoteIP  string    `json:"remote_ip"`
	UserAgent string    `json:"user_agent"`
	Status    int       `json:"status"`
}

// Token represents a JWT token
type Token struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Users stores in-memory users
var Users = []User{
	{
		ID:       "user1",
		Username: "example",
		Password: "example",
	},
}

// Authenticate handles authentication
func Authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	for _, u := range Users {
		if u.Username == user.Username && u.Password == user.Password {
			token := generateToken(u.ID)
			http.SetCookie(w, &http.Cookie{
				Name:     "token",
				Value:    token,
				HttpOnly: true,
				Secure:   true,
				Path:     "/",
				MaxAge:   60 * 60 * 24 * 7, // Token expires in 1 week
			})
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

// generateToken generates a JWT token
func generateToken(userID string) string {
	token := &Token{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)), // Token expires in 1 week
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	secretKey := []byte("super-secret") // Replace this with a secure key
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, token)
	signedToken, err := jwtToken.SignedString(secretKey)
	if err != nil {
		log.Fatalf("Error generating token: %v", err)
	}
	return signedToken
}

// requiresAuth is a middleware for authentication
func requiresAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenString := cookie.Value
		secretKey := []byte("super-secret") // Replace this with a secure key
		claims := &Token{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Method)
			}
			return secretKey, nil
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Token)
		if !ok || !token.Valid {
			http.Error(w, "Token expired", http.StatusUnauthorized)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), "userID", claims.UserID))
		next.ServeHTTP(w, r)
	})
}

type ResponseWriterWithStatus struct {
	http.ResponseWriter
	StatusCode int
}

func (rw *ResponseWriterWithStatus) WriteHeader(code int) {
	rw.StatusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Authorize is a middleware for authorization
func Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value("userID").(string)
		// Add your authorization logic here. For simplicity, we'll just log the userID
		log.Printf("User %s accessed path: %s", userID, r.URL.Path)
		next.ServeHTTP(w, r)
	})
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

func protectedResource(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("This is a protected resource."))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/auth", Authenticate).Methods("POST")
	r.PathPrefix("/protected/").Methods("GET").Handler(
		requiresAuth(
			Authorize(
				LoggingMiddleware(http.HandlerFunc(protectedResource)),
			),
		),
	)

	log.Fatal(http.ListenAndServe(":8080", r))
}
