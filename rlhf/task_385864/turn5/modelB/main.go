package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
)

var (
	addr                  = flag.String("addr", ":8080", "HTTP server address")
	maxConcurrentRequests = int32(100)
	workQueue             = make(chan struct{}, 100)
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func asyncHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&maxConcurrentRequests) <= int32(len(workQueue)) {
		http.Error(w, "Too many concurrent requests", http.StatusTooManyRequests)
		return
	}

	workQueue <- struct{}{}
	defer func() { <-workQueue }()

	go func() {
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
		log.Printf("Processed request from %s", r.RemoteAddr)
	}()

	fmt.Fprintf(w, "Hello, %s! Your request is being processed asynchronously.\n", r.RemoteAddr)
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
	}

	err := http2.ConfigureServer(srv, nil)
	if err != nil {
		log.Fatal("Error configuring HTTP/2:", err)
	}

	log.Printf("Server starting on %s...", *addr)
	log.Fatal(srv.ListenAndServeTLS("cert.pem", "key.pem"))
}
