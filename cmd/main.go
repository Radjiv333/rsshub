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
	"RSSHub/internal/aggregator"
	"RSSHub/internal/domain"
	"RSSHub/internal/domain/utils"
	"RSSHub/internal/share"
	"RSSHub/pkg/lock"
	"RSSHub/pkg/logger"
)

func main() {
	logger.Init()
	var agg *aggregator.Aggregator

	// Establishing DB connection
	repo, err := db.NewPostgresRepository()
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer repo.Close()

	if len(os.Args) < 2 {
		fmt.Println("Usage: rsshub COMMAND [OPTIONS]")
		fmt.Println("Commands: add, list, delete")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch":
		// Locking the fetch command, so that that there would not be 2 'fetch' funning apps
		if err := lock.Acquire(); err != nil {
			log.Fatalf("cannot start fetch: %v", err)
		}
		defer lock.Release()

		// Introducing Ctrl+C signal
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()

		// Introducing aggregator
		cliInterval, err := utils.GetAndParseCliInterval()
		if err != nil {
			stop()
			log.Fatalf("failed to fetch interval value from env file: %v", err)
		}
		// workersNum, err := utils.GetAndParseWorkersNum()
		// if err != nil {
		// 	stop()
		// 	log.Fatalf("failed to fetch interval value from env file: %v", err)
		// }

		agg = aggregator.NewAggregator(cliInterval, repo)

		// Starting feed fetch
		if err := agg.Start(ctx); err != nil {
			stop()
			log.Fatalf("failed to start aggregator: %v", err)
		}
		fmt.Printf("The background process for fetching feeds has started (interval = %v, workers = 3)")

		// Introducing Sharegator
		dbInterval, err := utils.GetAndParseDBInterval()
		if err != nil {
			stop()
			log.Fatalf("failed to fetch DB interval value from env file: %v", err)
		}
		share := share.NewShareVar(repo, agg)

		// Update the current feed fetch interval
		share.UpdateShare(dbInterval, ctx)

		// Waiting for Ctrl+C
		<-ctx.Done()
		agg.Stop()
		logger.Debug("Aggregator stopped cleanly")
		share.Stop()
		logger.Debug("Sharegator stopped cleanly")
		fmt.Println("Graceful shutdown: aggregator stopped")

	case "add":
		addCmd := flag.NewFlagSet("add", flag.ExitOnError)
		feedName := addCmd.String("name", "", "Feed name")
		feedURL := addCmd.String("url", "", "Feed URL")
		addCmd.Parse(os.Args[2:])

		if *feedName == "" || *feedURL == "" {
			fmt.Println("Usage: rsshub add --name <feed-name> --url <feed-url>")
			os.Exit(1)
		}

		feed := domain.Feed{
			Name:      *feedName,
			URL:       *feedURL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		logger.Debug("Adding feed to the DB...", "feed", feed)
		err := repo.AddFeed(feed)
		if err != nil {
			log.Fatalf("failed to insert feed: %v", err)
		}

		fmt.Printf("Feed '%s' added successfully!\n", *feedName)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		feedNum := listCmd.Int("num", 0, "Number of feeds to display (default: all)")
		listCmd.Parse(os.Args[2:])

		feeds, err := repo.ListFeeds(*feedNum)
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
		feedName := deleteCmd.String("name", "", "Feed name to delete")
		deleteCmd.Parse(os.Args[2:])

		if *feedName == "" {
			fmt.Println("Usage: rsshub delete --name <feed-name>")
			os.Exit(1)
		}

		err := repo.DeleteFeed(*feedName)
		if err != nil {
			log.Fatalf("failed to delete feed: %v", err)
		}

		fmt.Printf("Feed '%s' deleted successfully\n", *feedName)

	case "articles":
		articlesCmd := flag.NewFlagSet("articles", flag.ExitOnError)
		feedName := articlesCmd.String("feed-name", "", "Feed name to list articles for")
		num := articlesCmd.Int("num", 3, "Number of articles to show")
		articlesCmd.Parse(os.Args[2:])

		if *feedName == "" {
			fmt.Println("Usage: rsshub articles --feed-name <name> [--num N]")
			os.Exit(1)
		}

		feed, err := repo.ListFeedByName(*feedName)
		if err != nil {
			log.Fatalf("failed to get feed by name: %v", err)
		}
		if feed.Name == "" {
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

	case "set-interval":
		intervalCmd := flag.NewFlagSet("set-interval", flag.ExitOnError)
		duration := intervalCmd.String("duration", "", "New interval for fetching feeds")
		intervalCmd.Parse(os.Args[2:])

		if *duration == "" {
			log.Fatal("Usage: rsshub set-interval --duration <duration>")
		}

		_, err := utils.ParseIntervalToDuration(*duration)
		if err != nil {
			log.Fatalf("invalid duration: %v\n", err)
		}

		// Set the new interval
		err = repo.SetInterval(*duration)
		if err != nil {
			log.Fatalf("error updating interval in db: %v", err)
		}

	// case "set-workers":
	// 	workersCmd := flag.NewFlagSet("set-workers", flag.ExitOnError)
	// 	workers := workersCmd.Int("workers", 0, "Number of workers to use")
	// 	workersCmd.Parse(os.Args[2:])

	// 	if *workers <= 0 {
	// 		fmt.Println("Usage: rsshub set-workers --workers <number>")
	// 		os.Exit(1)
	// 	}

	// 	err := agg.Resize(*workers)
	// 	if err != nil {
	// 		fmt.Printf("Error resizing workers: %v\n", err)
	// 		os.Exit(1)
	// 	}

	// 	fmt.Printf("Number of workers changed to: %d\n", *workers)
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
