package domain

type FeedRepository interface {
	AddFeed(feed Feed) error
	DeleteFeed(name string) error
	ListFeeds(limit int) ([]Feed, error)
	GetFeedByName(name string) (Feed, error)
}

type ArticleRepository interface {
	AddArticles([]Article) error
	GetLatestArticles(feedName string, limit int) ([]Article, error)
}
