package domain

import "time"

type Article struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Link        string
	Description string
	PublishedAt time.Time
	FeedID      string
}
