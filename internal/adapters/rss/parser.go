package rss

import (
	"RSSHub/internal/domain"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- Structs for XML mapping ---

// --- Parser ---

// FetchAndParse retrieves and parses an RSS feed
func FetchAndParse(url string) (*domain.RSSFeed, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read RSS body: %w", err)
	}

	var feed domain.RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	return &feed, nil
}

func ParsePubDate(pubDate string) (time.Time, error) {
	for _, layout := range domain.TimeLayouts {
		if t, err := time.Parse(layout, pubDate); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("could not parse date: %s", pubDate)
}
