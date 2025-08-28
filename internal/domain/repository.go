package domain

import "time"

// Repository defines all DB operations the app needs.
type Repository interface {
	// Feeds
	AddFeed(feed Feed) error
	ListFeeds(limit int) ([]Feed, error)
	DeleteFeed(name string) error
	UpdateFeedTimestamp(feedID string, updatedAt time.Time) error

	// Articles
	AddArticle(article Article) error
	ListArticlesByFeed(feedID string, limit int) ([]Article, error)

	// Share
	FetchInterval() (string, error)
	SetInterval(interval string) error
	SetDefaultCLIInterval(interval string) error

	// Shutdown
	Close() error
}
