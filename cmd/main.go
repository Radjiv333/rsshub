package main

import (
	"RSSHub/internal/adapters/api"
	"RSSHub/internal/adapters/db"
	"RSSHub/internal/domain"
	"RSSHub/internal/domain/utils"
	"RSSHub/pkg/lock"
	"RSSHub/pkg/logger"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	logger.Init()
	var agg *api.Aggregator

	// Establishing DB connection
	repo, err := db.NewPostgresRepository()
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer repo.Close()

	if len(os.Args) >= 2 && (os.Args[1] == "--help" || os.Args[1] == "-h" || os.Args[1] == "help" || os.Args[1] == "-help") {
		utils.PrintHelp()
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: rsshub COMMAND [OPTIONS]")
		fmt.Println("Commands: add, list, delete")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "fetch":
		// lock.Release()
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
		workersNum, err := utils.GetAndParseWorkersNum()
		if err != nil {
			stop()
			log.Fatalf("failed to fetch workers number from env file: %v", err)
		}
		if workersNum == 0 {
			stop()
			log.Fatalf("number of workers cannot be 0")
		}

		agg = api.NewAggregator(cliInterval, workersNum, repo)

		// Starting feed fetch
		if err := agg.Start(ctx); err != nil {
			stop()
			log.Fatalf("failed to start aggregator: %v", err)
		}
		fmt.Printf("The background process for fetching feeds has started (interval = %v, workers = %d)\n", cliInterval, workersNum)

		// Introducing Sharegator
		dbInterval, err := utils.GetAndParseDBInterval()
		if err != nil {
			stop()
			log.Fatalf("failed to fetch DB interval value from env file: %v", err)
		}
		share := api.NewShareVar(repo, agg)

		// Update the current feed fetch interval
		share.UpdateShare(dbInterval, workersNum, ctx)

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

		resp, err := http.Get(*feedURL)
		if err != nil {
			fmt.Printf("Could not access the site through url: %s\n", *feedURL)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Got an unexpected code from url: %d\n", resp.StatusCode)
			os.Exit(1)
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Could not read the RSS body :(\n")
			os.Exit(1)
		}

		var testFeed domain.RSSFeed
		if err := xml.Unmarshal(data, &testFeed); err != nil {
			fmt.Printf("Could not parse the RSS body :(\n")
			os.Exit(1)
		}

		if testFeed.Channel.Title == "" {
			fmt.Printf("Our struct does not work with this site!\n")
			os.Exit(1)
		}

		feed := domain.Feed{
			Name:      *feedName,
			URL:       *feedURL,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		logger.Debug("Adding feed to the DB...", "feed", feed)
		err = repo.AddFeed(feed)
		if err != nil {
			log.Fatalf("failed to insert feed: %v", err)
		}

		fmt.Printf("Feed '%s' added successfully!\n", *feedName)

	case "list":
		listCmd := flag.NewFlagSet("list", flag.ExitOnError)
		feedNum := listCmd.Int("num", 0, "Number of feeds to display (default: all)")
		listCmd.Parse(os.Args[2:])

		if *feedNum < 0 {
			fmt.Println("The number of feeds should be more than 0")
			os.Exit(1)
		}

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
			fmt.Println(err.Error())
			os.Exit(1)
		}

		fmt.Printf("Feed '%s' deleted successfully\n", *feedName)

	case "articles":
		articlesCmd := flag.NewFlagSet("articles", flag.ExitOnError)
		feedName := articlesCmd.String("feed-name", "", "Feed name to list articles for")
		num := articlesCmd.Int("num", 3, "Number pkgof articles to show")
		articlesCmd.Parse(os.Args[2:])

		if *feedName == "" {
			fmt.Println("Usage: rsshub articles --feed-name <name> [--num N]")
			os.Exit(1)
		}

		if *num <= 0 || *num > 20 {
			fmt.Println("The number of articles should be more than 0 and less than 20")
			os.Exit(1)
		}

		feed, err := repo.ListFeedByName(*feedName)
		if err != nil {
			fmt.Printf("This feed have not yet been uploaded to the program or was deleted!\n")
			os.Exit(1)
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
			fmt.Printf("Usage: rsshub set-interval --duration <duration>\n")
			os.Exit(1)
		}

		dur, err := utils.ParseIntervalToDuration(*duration)
		if err != nil {
			log.Fatalf("invalid duration: %v\n", err)
		}

		// Set the new interval
		err = repo.SetInterval(*duration)
		if err != nil {
			log.Fatalf("error updating interval in db: %v", err)
		}
		fmt.Printf("The interval of fetching feeds changed to %v\n", dur)

	case "set-workers":
		if len(os.Args) < 3 {
			fmt.Println("Usage: rsshub set-workers <number>")
			os.Exit(1)
		}

		workersNum, err := strconv.Atoi(os.Args[2]) // Convert the argument to an integer
		if err != nil || workersNum <= 0 || workersNum >= 100 {
			fmt.Println("Usage: rsshub set-workers <number> (number should be greater than 0 or less than or equal to 100)")
			os.Exit(1)
		}

		err = repo.SetWorkers(workersNum)
		if err != nil {
			log.Fatalf("error updating interval in db: %v", err)
		}
		fmt.Printf("The number of workers have changed to %v\n", workersNum)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
