package share

import (
	"fmt"
	"strconv"
	"time"

	"RSSHub/internal/domain"
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

func (share *ShareVariables) UpdateShare(tickerDuration time.Duration) error {
	share.ticker = time.NewTicker(tickerDuration)
	go func() {
		for range share.ticker.C {
			// Getting interval value from db
			dbInterval, err := share.repo.FetchInterval()
			if err != nil {
				logger.Debug("Getting interval from db", "interval", dbInterval)
				continue
			}

			interval, err := ParseInterval(dbInterval)
			if err != nil {
				logger.Error("error parsing interval that came from db", "error", err)
				continue
			}

			if share.agg.GetCurrentInterval() != interval {
				share.agg.UpdateCurrentInterval(interval)
				logger.Debug("Current interval after update", "interval", share.agg.GetCurrentInterval())
			}
		}
	}()
	return nil
}

func ParseInterval(intervalStr string) (time.Duration, error) {
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
