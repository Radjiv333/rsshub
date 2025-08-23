package rss

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- Structs for XML mapping ---

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// --- Parser ---

// FetchAndParse retrieves and parses an RSS feed
func FetchAndParse(url string) (*RSSFeed, error) {
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

	var feed RSSFeed
	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("failed to parse RSS XML: %w", err)
	}

	return &feed, nil
}

// ParseTime safely converts pubDate into Go's time.Time
func ParseTime(pubDate string) (time.Time, error) {
	// RSS pubDate example: Mon, 06 Sep 2021 12:00:00 GMT
	layout := time.RFC1123Z // covers "Mon, 02 Jan 2006 15:04:05 -0700"
	t, err := time.Parse(layout, pubDate)
	if err != nil {
		// fallback: try without timezone offset
		t, err = time.Parse(time.RFC1123, pubDate)
	}
	return t, err
}
