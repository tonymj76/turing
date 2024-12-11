package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"

	"golang.org/x/net/html"
)

func main() {
	// Accept URL as a command-line argument
	url := flag.String("url", "https://example.com", "URL of the HTML document")
	flag.Parse()

	if *url == "" {
		log.Fatal("Please provide a URL using the -url flag")
	}

	res, err := http.Get(*url)
	if err != nil {
		log.Fatalf("Error fetching URL: %v", err)
	}
	defer res.Body.Close()

	// Use streaming parsing for large HTML documents
	tokenizer := html.NewTokenizer(res.Body)

	for {
		tt := tokenizer.Next()
		switch {
		case tt == html.ErrorToken:
			if err := tokenizer.Err(); err != nil && err != io.EOF {
				log.Fatalf("Error parsing HTML: %v", err)
			}
			return // End of document

		case tt == html.StartTagToken:
			tok := tokenizer.Token()
			fmt.Printf("<%s ", tok.Data)
			for _, attr := range tok.Attr {
				fmt.Printf("%s=%q ", attr.Key, attr.Val)
			}
			fmt.Println(">")
		}
	}
}
