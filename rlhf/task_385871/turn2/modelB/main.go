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
	url := "https://example.com" // Replace with the URL of the large HTML document

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("User-Agent", "Go HTML Scraper/1.0")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer resp.Body.Close()

	links := ParseHTMLLinksFromReader(resp.Body)
	fmt.Println("Extracted Links:")
	for _, link := range links {
		fmt.Println(link)
	}
}

func ParseHTMLLinksFromReader(r io.Reader) []string {
	var links []string
	seen := make(map[string]struct{})
	tokenizer := html.NewTokenizer(r)

	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			if err := tokenizer.Err(); err != nil && err.Error() != "EOF" {
				log.Printf("Error parsing HTML: %v", err)
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
