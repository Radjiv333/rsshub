package api

import (
	"context"
	"database/sql"
	"time"

	"RSSHub/internal/domain"
	"RSSHub/internal/domain/utils"
	"RSSHub/pkg/config"
	"RSSHub/pkg/logger"
)

type ShareVariables struct {
	repo   domain.Repository
	ticker *time.Ticker
	agg    domain.Aggregator
}

var _ domain.ShareVariables = (*ShareVariables)(nil)

func NewShareVar(repo domain.Repository, agg domain.Aggregator) *ShareVariables {
	return &ShareVariables{repo: repo, agg: agg}
}

func (share *ShareVariables) UpdateShare(dbInterval time.Duration, workersNum int, ctx context.Context) {
	share.ticker = time.NewTicker(dbInterval)

	share.repo.SetDefaultCliIntervalAndWorkersNum(config.GetEnvInterval(), workersNum)

	go func() {
		for {
			select {
			case <-share.ticker.C:
				// Getting interval value from db
				dbInterval, err := share.repo.FetchCliInterval()
				if err != sql.ErrNoRows {
					logger.Debug("Getting interval from db", "interval", dbInterval)
				}
				workersNum, err := share.repo.FetchWorkersNumber()
				if err != sql.ErrNoRows {
					logger.Debug("Getting workers number from db", "workers", workersNum)
				}

				interval, err := utils.ParseIntervalToDuration(dbInterval)
				if err != nil {
					logger.Error("error parsing interval that came from db", "error", err, "interval", interval)
					continue
				}

				// Interval Update
				if share.agg.GetCurrentInterval() != interval {
					share.agg.SetCurrentInterval(interval)
					share.agg.RestartTicker()
					logger.Debug("Current interval after update", "interval", share.agg.GetCurrentInterval())
				}

				// Worker number update
				oldWorkersNum := share.agg.GetWorkersNum()
				if oldWorkersNum != workersNum {
					share.agg.SetWorkersNum(workersNum)
					share.agg.UpdateWorkers(ctx, oldWorkersNum, workersNum)
					logger.Debug("Current workers number after update", "workers number", share.agg.GetWorkersNum())
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (share *ShareVariables) Stop() {
	share.ticker.Stop()
}
