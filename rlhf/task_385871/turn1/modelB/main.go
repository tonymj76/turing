package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/net/html"
)

// Node represents a parsed HTML node
type Node struct {
	Type     html.NodeType
	Data     string
	Attr     []html.Attribute
	Children []*Node
}

func main() {
	// Open the HTML file
	filename := "large_document.html"
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	// Parse the HTML document
	doc, err := html.Parse(file)
	if err != nil {
		log.Fatalf("Error parsing HTML: %v", err)
	}

	// Process the nodes
	processNodes(doc)
}

func processNodes(n *html.Node) {
	// Example: Print the content of <title> tags
	if n.Type == html.ElementNode && n.Data == "title" {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				fmt.Println("Title:", c.Data)
			}
		}
	}

	// Traverse child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processNodes(c)
	}
}
