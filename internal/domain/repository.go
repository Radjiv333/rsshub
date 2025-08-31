package domain

import (
	"time"
)

// Repository defines all DB operations the app needs.
type Repository interface {
	// Feeds
	AddFeed(feed Feed) error
	ListFeeds(limit int) ([]Feed, error)
	ListFeedByName(feedName string) (Feed, error)
	DeleteFeed(name string) error
	UpdateFeedTimestamp(feedID string, updatedAt time.Time) error

	// Articles
	AddArticle(article Article) error
	ListArticlesByFeed(feedID string, limit int) ([]Article, error)
	ListArticles(feedName string, num int) ([]Article, error)

	// Share
	FetchCliInterval() (string, error)
	SetInterval(interval string) error
	SetDefaultCliIntervalAndWorkersNum(interval string, workersNum int) error
	SetWorkers(workersNum int) error
	FetchWorkersNumber() (int, error)

	// Shutdown
	Close() error
}
