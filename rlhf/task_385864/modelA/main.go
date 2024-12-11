package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	addr = flag.String("addr", ":8080", "HTTP server address")
)

func asyncHandler(w http.ResponseWriter, r *http.Request) {
	// Simulate some asynchronous work using goroutines
	delayStr := r.FormValue("delay")
	dalay, err := strconv.Atoi(delayStr)
	if err != nil || dalay < 0 {
		http.Error(w, "invalid delay", http.StatusBadRequest)
		return
	}

	resultStream := make(chan string)

	go func() {
		time.Sleep(time.Duration(dalay) * time.Second)
		resultMsg := fmt.Sprintf("Hello, %s! The delay was: %s\n", r.RemoteAddr, r.FormValue("delay"))
		resultStream <- resultMsg
	}()

	// wait for the goroutine to finish
	select {
	case msg := <-resultStream:
		// Write the response in the main context
		fmt.Fprint(w, msg)
		w.WriteHeader(http.StatusOK)
	case <-time.After(20 * time.Second): // Timeout to prevent indefinite wait
		http.Error(w, "Request timeout", http.StatusRequestTimeout)
	}
}

func main() {
	flag.Parse()

	// Register the asynchronous handler
	http.HandleFunc("/", asyncHandler)

	// Configure the server to use HTTP/2
	srv := &http.Server{
		Addr:           *addr,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Enable HTTP/2
	// srv.TLSConfig = &tls.Config{
	// 	MinVersion:       tls.VersionTLS12,
	// 	CurvePreferences: []tls.CurveID{tls.CurveP256},
	// }
	fmt.Println("staring server")
	log.Fatal(srv.ListenAndServe())
}
