package db

import (
	"RSSHub/internal/domain"
	"RSSHub/pkg/logger"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var _ domain.Repository = (*PostgresRepository)(nil)

type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new Postgres repo
func NewPostgresRepository() (*PostgresRepository, error) {
	connStr := "host=localhost port=5432 user=postgres password=changeme dbname=rsshub sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	logger.Debug("Succefully connected to Database!")

	return &PostgresRepository{db: db}, nil
}

// Close DB connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// -------------------------------------------------------------Feeds--------------------------------------------------------------------

func (r *PostgresRepository) AddFeed(feed domain.Feed) error {
	query := `
		INSERT INTO feeds (name, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (name) DO NOTHING;
	`
	_, err := r.db.Exec(query, feed.Name, feed.URL, feed.CreatedAt, feed.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListFeedByName(feedName string) (domain.Feed, error) {
	feed := domain.Feed{}
	query := `
		SELECT id, name, url, created_at, updated_at
		FROM feeds
		WHERE name = $1
	`
	err := r.db.QueryRow(query, feedName).Scan(&feed.ID, &feed.Name, &feed.URL, &feed.CreatedAt, &feed.UpdatedAt)
	if err != nil {
		return domain.Feed{}, err
	}
	return feed, nil
}

func (r *PostgresRepository) ListFeeds(limit int) ([]domain.Feed, error) {
	var rows *sql.Rows
	var err error
	query := `
		SELECT id, name, url, created_at, updated_at
		FROM feeds
		ORDER BY created_at DESC
	`
	if limit < 0 {
		return nil, fmt.Errorf("--num parameter cannot be negative")
	} else if limit != 0 {
		query += "LIMIT $1"
		rows, err = r.db.Query(query, limit)
	} else {
		rows, err = r.db.Query(query)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []domain.Feed
	for rows.Next() {
		var f domain.Feed
		err := rows.Scan(&f.ID, &f.Name, &f.URL, &f.CreatedAt, &f.UpdatedAt)
		if err != nil {
			return nil, err
		}
		feeds = append(feeds, f)
	}
	return feeds, nil
}

func (r *PostgresRepository) UpdateFeedTimestamp(feedID string, updatedAt time.Time) error {
	_, err := r.db.Exec(`
		UPDATE feeds 
		SET updated_at = $1 
		WHERE id = $2
	`, updatedAt, feedID)
	return err
}

func (r *PostgresRepository) DeleteFeed(name string) error {
	query := `DELETE FROM feeds WHERE name = $1`
	result, err := r.db.Exec(query, name)
	if err != nil {
		logger.Error("Error deleting feed", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error("Error getting rows affected after delete", "error", err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("The feed is not present in db!")
	}

	return nil
}

// -------------------------------------------------------------Articles--------------------------------------------------------------------

// AddArticle inserts a new article (ignores duplicates by link)
func (r *PostgresRepository) AddArticle(article domain.Article) error {
	query := `
		INSERT INTO articles (created_at, updated_at, title, link, description, published_at, feed_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (link) DO NOTHING;
	`
	_, err := r.db.Exec(query,
		article.CreatedAt,
		article.UpdatedAt,
		article.Title,
		article.Link,
		article.Description,
		article.PublishedAt,
		article.FeedID,
	)
	return err
}

// ListArticles returns the N latest articles for a feed
func (r *PostgresRepository) ListArticles(feedName string, num int) ([]domain.Article, error) {
	query := `
		SELECT a.id, a.created_at, a.updated_at, a.title, a.link, a.description, a.published_at, a.feed_id
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE f.name = $1
		ORDER BY a.published_at DESC
		LIMIT $2;
	`

	rows, err := r.db.Query(query, feedName, num)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var a domain.Article
		err := rows.Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt, &a.Title, &a.Link, &a.Description, &a.PublishedAt, &a.FeedID)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}
	return articles, nil
}

// ListArticlesByFeed returns the N most recent articles for a feed
func (r *PostgresRepository) ListArticlesByFeed(feedID string, limit int) ([]domain.Article, error) {
	query := `
		SELECT id, feed_id, title, link, description, published_at, created_at, updated_at
		FROM articles
		WHERE feed_id = $1
		ORDER BY published_at DESC
		LIMIT $2;`

	rows, err := r.db.Query(query, feedID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []domain.Article
	for rows.Next() {
		var a domain.Article
		err := rows.Scan(
			&a.ID, &a.FeedID, &a.Title, &a.Link,
			&a.Description, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}

	return articles, nil
}

// -------------------------------------------------------------Share--------------------------------------------------------------------

func (r *PostgresRepository) FetchCliInterval() (string, error) {
	query := `SELECT interval FROM share`
	var interval string
	err := r.db.QueryRow(query).Scan(&interval)
	if err == sql.ErrNoRows {
		return "", err
	}
	return interval, nil
}

func (r *PostgresRepository) SetInterval(interval string) error {
	query := `UPDATE share SET interval = $1 WHERE id = 1`
	_, err := r.db.Exec(query, interval)
	return err
}

func (r *PostgresRepository) SetDefaultCliIntervalAndWorkersNum(interval string, workersNum int) error {
	query := `
		INSERT INTO share (id, interval, workers_num)
		VALUES (1, $1, $2)
		ON CONFLICT (id)
		DO UPDATE SET interval = EXCLUDED.interval, workers_num = EXCLUDED.workers_num;
	`
	_, err := r.db.Exec(query, interval, workersNum)
	return err
}

func (r *PostgresRepository) SetWorkers(workersNum int) error {
	query := `UPDATE share SET workers_num = $1 WHERE id = 1`
	_, err := r.db.Exec(query, workersNum)
	return err
}

func (r *PostgresRepository) FetchWorkersNumber() (int, error) {
	query := `SELECT workers_num FROM share`
	var workersNum int
	err := r.db.QueryRow(query).Scan(&workersNum)
	if err == sql.ErrNoRows {
		return 0, err
	}
	return workersNum, nil
}
