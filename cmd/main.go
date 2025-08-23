package main

import (
	"fmt"
	"log"

	"RSSHub/internal/adapters/rss"
)

func main() {
	url := "https://techcrunch.com/feed/"
	feed, err := rss.FetchAndParse(url)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Feed Title:", feed.Channel.Title)
	for i, item := range feed.Channel.Items {
		if i >= 3 {
			break
		}
		fmt.Printf("%d. %s (%s)\n", i+1, item.Title, item.PubDate)
	}
}
