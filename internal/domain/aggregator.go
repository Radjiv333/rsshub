package domain

import (
	"context"
	"time"
)

type Aggregator interface {
	Start(ctx context.Context) error
	Stop()
	Worker(ctx context.Context, id int)
	GetCurrentInterval() time.Duration
	SetCurrentInterval(interval time.Duration)
	RestartTicker()
	SetInterval(d time.Duration)
	GetWorkersNum() int
	SetWorkersNum(workersNum int)
	UpdateWorkers(ctx context.Context, oldWorkersNum int, workersNum int)
}
