package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"RSSHub/internal/adapters/db"
	"RSSHub/internal/adapters/rss"
	"RSSHub/internal/aggregator"
	"RSSHub/internal/domain"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rsshub COMMAND [OPTIONS]")
		fmt.Println("Commands: add, list, delete")
		os.Exit(1)
	}

	repo, err := db.NewPostgresRepository()
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer repo.Close()

	switch os.Args[1] {
	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		name := addCmd.String("name", "", "Feed name")
		url := addCmd.String("url", "", "Feed URL")
		addCmd.Parse(os.Args[2:])

		if *name == "" || *url == "" {
			fmt.Println("Usage: rsshub add --name <feed-name> --url <feed-url>")
			os.Exit(1)
		}

		feed := domain.Feed{
			Name:      *name,
			URL:       *url,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		err := repo.AddFeed(feed)
		if err != nil {
			log.Fatalf("failed to insert feed: %v", err)
		}
		fmt.Printf("Feed '%s' added successfully\n", *name)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		num := listCmd.Int("num", 0, "Number of feeds to display (default: all)")
		listCmd.Parse(os.Args[2:])

		feeds, err := repo.ListFeeds(*num)
		if err != nil {
			log.Fatalf("failed to list feeds: %v", err)
		}

		fmt.Println("\n# Available RSS Feeds")
		for i, f := range feeds {
			fmt.Printf("%d. Name: %s\n   URL: %s\n   Added: %s\n\n",
				i+1, f.Name, f.URL, f.CreatedAt.Format("2006-01-02 15:04"),
			)
		}

	case "delete":
		deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
		name := deleteCmd.String("name", "", "Feed name to delete")
		deleteCmd.Parse(os.Args[2:])

		if *name == "" {
			fmt.Println("Usage: rsshub delete --name <feed-name>")
			os.Exit(1)
		}

		err := repo.DeleteFeed(*name)
		if err != nil {
			log.Fatalf("failed to delete feed: %v", err)
		}

		fmt.Printf("Feed '%s' deleted successfully\n", *name)

	case "articles":
		articlesCmd := flag.NewFlagSet("articles", flag.ExitOnError)
		feedName := articlesCmd.String("feed-name", "", "Feed name to list articles for")
		num := articlesCmd.Int("num", 3, "Number of articles to show")
		articlesCmd.Parse(os.Args[2:])

		if *feedName == "" {
			fmt.Println("Usage: rsshub articles --feed-name <name> [--num N]")
			os.Exit(1)
		}

		feeds, _ := repo.ListFeeds(100)
		var feed domain.Feed
		found := false
		for _, f := range feeds {
			if f.Name == *feedName {
				feed = f
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("feed '%s' not found", *feedName)
		}

		articles, err := repo.ListArticlesByFeed(feed.ID, *num)
		if err != nil {
			log.Fatalf("failed to fetch articles: %v", err)
		}

		fmt.Printf("Feed: %s\n\n", feed.Name)
		for i, a := range articles {
			fmt.Printf("%d. [%s] %s\n   %s\n\n",
				i+1,
				a.PublishedAt.Format("2006-01-02"),
				a.Title,
				a.Link,
			)
		}

	case "fetch-once":
		fetchCmd := flag.NewFlagSet("fetch-once", flag.ExitOnError)
		feedName := fetchCmd.String("feed-name", "", "Feed name to fetch")
		fetchCmd.Parse(os.Args[2:])

		if *feedName == "" {
			fmt.Println("Usage: rsshub fetch-once --feed-name <name>")
			os.Exit(1)
		}

		feeds, _ := repo.ListFeeds(50)
		var feed domain.Feed
		found := false
		for _, f := range feeds {
			if f.Name == *feedName {
				feed = f
				found = true
				break
			}
		}
		if !found {
			log.Fatalf("feed '%s' not found", *feedName)
		}

		parsed, err := rss.FetchAndParse(feed.URL)
		if err != nil {
			log.Fatalf("failed to fetch RSS: %v", err)
		}

		fmt.Println("Feed:", parsed.Channel.Title)

		for _, item := range parsed.Channel.Items {
			article := domain.Article{
				FeedID:      feed.ID,
				Title:       item.Title,
				Link:        item.Link,
				Description: item.Description,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}

			parsedTime, err := time.Parse(time.RubyDate, item.PubDate)
			if err == nil {
				article.PublishedAt = parsedTime
			} else {
				// fallback: use now if parsing fails
				article.PublishedAt = time.Now()
			}

			err = repo.AddArticle(article)
			if err != nil {
				log.Printf("skipping article '%s': %v\n", article.Title, err)
			} else {
				fmt.Printf("saved: %s\n", article.Title)
			}
		}

	case "fetch":
		fetchCmd := flag.NewFlagSet("fetch", flag.ExitOnError)
		interval := fetchCmd.Int("interval", 3, "Fetch interval in minutes")
		fetchCmd.Parse(os.Args[2:])

		agg := aggregator.NewAggregator(time.Duration(*interval)*time.Minute, repo)

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		if err := agg.Start(ctx); err != nil {
			log.Fatalf("failed to start aggregator: %v", err)
		}

		<-ctx.Done() // wait for Ctrl+C
		if err := agg.Stop(); err != nil {
			log.Printf("aggregator stopped with error: %v", err)
		} else {
			fmt.Println("Aggregator stopped cleanly")
		}

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
