package aggregator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"RSSHub/internal/adapters/rss"
	"RSSHub/internal/domain"
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
}

func NewAggregator(defaultInterval time.Duration, repo domain.Repository) *Aggregator {
	return &Aggregator{
		interval: defaultInterval,
		workers:  3,                          // default pool size
		jobs:     make(chan domain.Feed, 100), // buffered channel
		repo:     repo,
	}
}

func (a *Aggregator) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("aggregator already running")
	}
	ctx, cancel := context.WithCancel(ctx)
	a.cancel = cancel
	a.ticker = time.NewTicker(a.interval)
	a.running = true
	a.mu.Unlock()

	// start workers
	for i := 0; i < a.workers; i++ {
		a.wg.Add(1)
		go a.worker(ctx, i)
	}

	// ticker loop
	a.wg.Add(1)
	go func() {
		defer a.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case <-a.ticker.C:
				fmt.Println("Tick: loading feeds…")
				feeds, err := a.repo.ListFeeds(5) // take 5 oldest feeds
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

	if a.running {
		a.ticker.Stop()
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
		case feed := <-a.jobs:
			fmt.Printf("[worker %d] fetching %s (%s)\n", id, feed.Name, feed.URL)

			parsed, err := rss.FetchAndParse(feed.URL)
			if err != nil {
				fmt.Printf("[worker %d] error fetching %s: %v\n", id, feed.Name, err)
				continue
			}

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
				parsedTime, err := time.Parse(time.RubyDate, item.PubDate)
				if err == nil {
					article.PublishedAt = parsedTime
				} else {
					article.PublishedAt = time.Now() // fallback
				}

				// Save to DB
				err = a.repo.AddArticle(article)
				if err != nil {
					fmt.Printf("[worker %d] skipping article '%s': %v\n", id, article.Title, err)
				} else {
					fmt.Printf("[worker %d] saved: %s\n", id, article.Title)
				}
			}

			// Update feed timestamp
			feed.UpdatedAt = time.Now()
			if err := a.repo.UpdateFeedTimestamp(feed.ID, feed.UpdatedAt); err != nil {
				fmt.Printf("[worker %d] failed to update feed timestamp: %v\n", id, err)
			}
		}
	}
}
