package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"
)

func asyncHandler(w http.ResponseWriter, r *http.Request) {
	// Start the asynchronous work
	go doWorkAsynchronously()

	// Respond to the client immediately
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintln(w, "Work has been started asynchronously. Check the logs for updates.")
}

func doWorkAsynchronously() {
	// Simulate a long-running task
	time.Sleep(2 * time.Second)
	log.Println("Work completed asynchronously.")
}

func main() {
	http.HandleFunc("/async", asyncHandler)
	// Configure the server to use HTTP/2
	srv := &http.Server{
		Addr:           ":8080",
		Handler:        http.DefaultServeMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Enable HTTP/2
	srv.TLSConfig = &tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.CurveP256},
	}
	fmt.Println("starting server")
	log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
}
