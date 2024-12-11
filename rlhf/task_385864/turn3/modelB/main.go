package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	addr      = flag.String("addr", ":8080", "HTTP server address")
	workQueue = make(chan string, 1000)
	wg        sync.WaitGroup
)

func worker() {
	defer wg.Done()
	for url := range workQueue {
		response, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching %s: %v", url, err)
			continue
		}
		data, err := io.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			log.Printf("Error reading body from %s: %v", url, err)
			continue
		}
		log.Printf("Fetched %s: %d bytes", url, len(data))
	}
}

func asyncHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	if _, err := http.NewRequest("GET", url, nil); err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}
	select {
	case workQueue <- url:
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "Request accepted. Processing URL %s asynchronously\n", url)
	default:
		http.Error(w, "Server busy. Try again later.", http.StatusServiceUnavailable)
	}
}

func main() {
	flag.Parse()

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker()
	}

	http.HandleFunc("/", asyncHandler)

	srv := &http.Server{
		Addr:           *addr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s...", *addr)
		if err := srv.ListenAndServeTLS("cert.pem", "key.pem"); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	close(workQueue)
	srv.Close()
	wg.Wait()
	log.Println("Server stopped.")
}
