package domain

import (
	"context"
	"time"
)

type Aggregator interface {
	Start(ctx context.Context) error
	Stop()
	Worker(ctx context.Context, id int)
	Resize(workers int) error
	GetCurrentInterval() time.Duration
	SetCurrentInterval(interval time.Duration)
	RestartTicker()
}
