package domain

import (
	"context"
	"time"
)

type Aggregator interface {
	Start(ctx context.Context) error
	Stop() error
	Worker(ctx context.Context, id int)
	Resize(workers int) error
	GetInterval() time.Duration
}
