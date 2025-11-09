# RSSHub - RSS Feed Aggregator

RSSHub is a CLI application that fetches, parses, and stores RSS feeds from various sources. It helps users stay informed by collecting publications from news sites, blogs, and forums in one centralized location.

## Features

- **RSS Feed Management**: Add, list, and delete RSS feeds
- **Background Processing**: Periodic fetching of RSS feeds using a worker pool
- **Parallel Processing**: Concurrent feed parsing and storage
- **Dynamic Configuration**: Change fetch intervals and worker counts on-the-fly
- **PostgreSQL Storage**: Persistent storage of feeds and articles
- **Docker Compose**: Easy local development and deployment

## Technologies Used

- **Language:** Go (1.23+)
- **Database:** PostgreSQL
- **Containerization:** Docker, Docker Compose
- **Migration Tool:** golang-migrate
- **Code Formatting:** gofumpt
- **Concurrency:** Goroutines, Channels, Worker Pool
- **Testing:** Go race detector for concurrency safety

## Quick Start

### Prerequisites

- Docker
- Docker Compose
- Go 1.23+ (for local development)

### Running with Docker Compose

1. **Start the Services and the Background fetcher**:
   ```bash
   make up
   ```
   This starts PostgreSQL and builds the RSSHub application.

2. **Build local app**:
   ```bash
   make build
   ```

## Usage

### Adding RSS Feeds

```bash
./rsshub add --name "tech-crunch" --url "https://techcrunch.com/feed/"
./rsshub add --name "hacker-news" --url "https://news.ycombinator.com/rss"
```

### Listing Feeds

```bash
./rsshub list                 # Show all feeds
./rsshub list --num 5        # Show 5 most recent feeds
```

### Viewing Articles

```bash
./rsshub articles --feed-name "tech-crunch"     # Show 3 latest articles
./rsshub articles --feed-name "tech-crunch" --num 5  # Show 5 latest articles
```

### Managing Background Processing

```bash
# Start background fetching
./rsshub fetch

# Change fetch interval (while fetch is running in another terminal)
./rsshub set-interval 2m

# Change number of workers
./rsshub set-workers 5

# Delete a feed
./rsshub delete --name "tech-crunch"
```

### Getting Help

```bash
./rsshub --help
```

## Makefile Commands

- `make build` - Build the application for local usage
- `make up` - Start services with Docker Compose
- `make upd` - Start services in detached mode
- `make down` - Stop services
- `make restart` - Restart services
- `make nuke` - Remove all containers, networks, and volumes
- `make migrate-up` - Run database migrations
- `make migrate-down` - Rollback database migrations
- `make migrate-version` - Check migration version
- `make fetch` - Start the RSS fetcher with environment variables

## Configuration

Default settings can be configured in `.env`:

```env
# CLI App
CLI_APP_TIMER_INTERVAL=3m
CLI_APP_WORKERS_COUNT=3

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=postgres
POSTGRES_PASSWORD=changeme
POSTGRES_DBNAME=rsshub
```

## Architecture
- **Hexagonal Architecture (Ports & Adapters)**: Separates domain logic from external systems like CLI and database
- **Worker Pool**: Concurrent processing of RSS feeds
- **Ticker-based Fetcher**: Periodic feed updates with configurable intervals
- **Graceful Shutdown**: Proper cleanup on termination
- **Race Condition Protection**: Safe concurrent operations

## Development

The project follows these Go standards:
- Code formatted with `gofumpt`
- Race condition detection enabled
- No external dependencies except for PostgreSQL driver
- Proper error handling and graceful shutdown

## Sample RSS Feeds

- TechCrunch: `https://techcrunch.com/feed/`
- Hacker News: `https://news.ycombinator.com/rss`
- BBC News: `https://feeds.bbci.co.uk/news/world/rss.xml`
- The Verge: `https://www.theverge.com/rss/index.xml`
- Ars Technica: `http://feeds.arstechnica.com/arstechnica/index`

## Important Notes

- Only one instance of the background fetcher can run at a time
- The application prevents DoS attacks by implementing rate limiting
- All goroutines are properly managed to prevent leaks
- Database connections are properly closed on shutdown

## Troubleshooting

If you encounter issues:
1. Check that PostgreSQL is running: `docker ps`
2. Verify migrations ran successfully: `make migrate-version`
3. Ensure the `.env` file is properly configured
4. Check application logs for detailed error messages

For more information, run `./rsshub --help` or check the individual command help.