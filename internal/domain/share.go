package domain

import (
	"context"
	"time"
)

type ShareVariables interface {
	UpdateShare(dbInterval time.Duration, workersNum int, ctx context.Context)
	Stop()
}
