package share

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"RSSHub/internal/domain"
	"RSSHub/pkg/config"
	"RSSHub/pkg/logger"
)

type ShareVariables struct {
	repo   domain.Repository
	ticker *time.Ticker
	agg    domain.Aggregator
}

func NewShareVar(repo domain.Repository, agg domain.Aggregator) *ShareVariables {
	return &ShareVariables{repo: repo, agg: agg}
}

func (share *ShareVariables) UpdateShare(dbInterval time.Duration, ctx context.Context) {
	share.ticker = time.NewTicker(dbInterval)
	share.repo.SetDefaultCLIInterval(config.GetEnvInterval())
	go func() {
		for {
			select {
			case <-share.ticker.C:
				// Getting interval value from db
				dbInterval, err := share.repo.FetchInterval()
				if err != sql.ErrNoRows {
					logger.Debug("Getting interval from db", "interval", dbInterval)
					continue
				}

				interval, err := ParseInterval(dbInterval)
				if err != nil {
					logger.Error("error parsing interval that came from db", "error", err, "interval", interval)
					continue
				}
				if share.agg.GetCurrentInterval() != interval {
					share.agg.UpdateCurrentInterval(interval)
					logger.Debug("Current interval after update", "interval", share.agg.GetCurrentInterval())
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func ParseInterval(intervalStr string) (time.Duration, error) {
	if len(intervalStr) < 2 {
		return 0, fmt.Errorf("env value for db_interval is invalid!")
	}

	unit := intervalStr[len(intervalStr)-1]
	value := intervalStr[:len(intervalStr)-1]

	interval, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid interval value %q: %w", value, err)
	}

	switch unit {
	case 's':
		return time.Duration(interval) * time.Second, nil
	case 'm':
		return time.Duration(interval) * time.Minute, nil
	case 'h':
		return time.Duration(interval) * time.Hour, nil
	case 'd':
		return time.Duration(interval) * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("unsupported unit: %c", unit)
	}
}

func (share *ShareVariables) Stop() {
	share.ticker.Stop()
}
