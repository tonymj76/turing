package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var addr = flag.String("addr", ":8080", "HTTP server address")

func asyncHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate work
	time.Sleep(2 * time.Second)

	fmt.Fprintf(w, "Hello, %s!", r.RemoteAddr)
}

func main() {
	flag.Parse()

	http.HandleFunc("/", asyncHandler)

	srv := &http.Server{
		Addr:           *addr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig: &tls.Config{
			MinVersion:       tls.VersionTLS12,
			CurvePreferences: []tls.CurveID{tls.CurveP256},
		},
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s...", *addr)
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %s", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")
	srv.Close() // Graceful shutdown
	log.Println("Server stopped")
}
