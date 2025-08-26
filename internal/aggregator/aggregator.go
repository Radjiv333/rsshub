package aggregator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"RSSHub/internal/adapters/rss"
	"RSSHub/internal/domain"
	"RSSHub/pkg/lock"
	"RSSHub/pkg/logger"
)

type Aggregator struct {
	interval time.Duration
	ticker   *time.Ticker
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	running  bool
	mu       sync.Mutex

	jobs    chan domain.Feed
	workers int
	repo    domain.Repository

	stopWorkers chan struct{}
}

func NewAggregator(defaultInterval time.Duration, repo domain.Repository) *Aggregator {
	return &Aggregator{
		interval:    defaultInterval,
		workers:     3,
		jobs:        make(chan domain.Feed, 100),
		repo:        repo,
		stopWorkers: make(chan struct{}),
	}
}



func (a *Aggregator) Start(ctx context.Context) error {
	logger.Debug("'Start' function", "file", "aggregator.go")

	a.mu.Lock()
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.ticker = time.NewTicker(a.interval)
	a.running = true
	a.mu.Unlock()
	
	

	// Start the worker pool with the desired number of workers
	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go a.worker(ctx, i)
	}

	// Ticker loop for loading and processing feeds at regular intervals
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.ticker.C:
				fmt.Println("Tick: loading feeds…")
				feeds, err := a.repo.ListFeeds(5) // Take 5 oldest feeds
				if err != nil {
					fmt.Printf("error loading feeds: %v\n", err)
					continue
				}
				for _, f := range feeds {
					select {
					case a.jobs <- f:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return nil
}

func (a *Aggregator) Stop() error {
	lock.Release()
	a.mu.Lock()
	if !a.running {
		a.mu.Unlock()
		return fmt.Errorf("aggregator not running")
	}
	a.cancel()
	a.ticker.Stop()
	a.mu.Unlock()

	a.wg.Wait()
	a.running = false
	return nil
}

func (a *Aggregator) SetInterval(d time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Only stop the ticker if it's running
	if a.running {
		fmt.Printf("Changing interval from %v to %v\n", a.interval, d)
		a.ticker.Stop() // Stop the old ticker

		// Create a new ticker with the new duration
		a.ticker = time.NewTicker(d)
	}

	a.interval = d
	fmt.Printf("Interval changed to %v\n", d)
}

// --- Worker function ---
func (a *Aggregator) worker(ctx context.Context, id int) {
	defer a.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case <-a.stopWorkers:
			fmt.Printf("[worker %d] stopping\n", id)
			return
		case feed := <-a.jobs:
			// Fetch and parse RSS for the feed
			fmt.Printf("[worker %d] fetching %s (%s)\n", id, feed.Name, feed.URL)

			parsed, err := rss.FetchAndParse(feed.URL)
			if err != nil {
				fmt.Printf("[worker %d] error fetching %s: %v\n", id, feed.Name, err)
				continue
			}

			// Process each article and save it to the database
			for _, item := range parsed.Channel.Items {
				article := domain.Article{
					FeedID:      feed.ID,
					Title:       item.Title,
					Link:        item.Link,
					Description: item.Description,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}

				// Parse pubDate if possible
				parsedTime, err := rss.ParsePubDate(item.PubDate)
				if err == nil {
					article.PublishedAt = parsedTime
				} else {
					fmt.Printf("[worker %d] warning: could not parse date '%s': %v\n", id, item.PubDate, err)
					article.PublishedAt = time.Now()
				}

				// Save to DB
				err = a.repo.AddArticle(article)
				if err != nil {
					fmt.Printf("[worker %d] skipping article '%s': %v\n", id, article.Title, err)
				} else {
					fmt.Printf("[worker %d] saved: %s\n", id, article.Title)
				}
			}

			// Update the feed timestamp after processing
			feed.UpdatedAt = time.Now()
			if err := a.repo.UpdateFeedTimestamp(feed.ID, feed.UpdatedAt); err != nil {
				fmt.Printf("[worker %d] failed to update feed timestamp: %v\n", id, err)
			}
		}
	}
}

func (a *Aggregator) Resize(workers int) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if workers == a.workers {
		return nil // no change
	}

	diff := workers - a.workers
	if diff > 0 {
		// Scale up: add more workers
		for i := 0; i < diff; i++ {
			a.wg.Add(1)
			go a.worker(context.Background(), a.workers+i)
		}
	} else {
		// Scale down: stop some workers
		for i := 0; i < -diff; i++ {
			a.stopWorkers <- struct{}{}
		}
	}

	a.workers = workers
	return nil
}
