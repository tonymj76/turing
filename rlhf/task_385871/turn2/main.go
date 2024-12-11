package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func main() {
	url := "https://example.com" // Replace with the URL of your large HTML document

	// HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Go HTML Scraper/1.0")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer res.Body.Close()

	links := extractLinksFromReader(res.Body)
	for _, link := range links {
		fmt.Println(link)
	}
}

func extractLinksFromReader(r io.Reader) []string {
	var links []string
	seen := make(map[string]struct{})
	tokenizer := html.NewTokenizer(r)

	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			if tokenizer.Err() != nil {
				log.Printf("Error parsing HTML: %v", tokenizer.Err())
			}
			return links
		case tt == html.StartTagToken:
			tok := tokenizer.Token()
			if tok.Data == "a" {
				for _, attr := range tok.Attr {
					if attr.Key == "href" {
						link := strings.TrimSpace(attr.Val)
						if _, exists := seen[link]; !exists {
							seen[link] = struct{}{}
							links = append(links, link)
						}
					}
				}
			}
		}
	}
}
