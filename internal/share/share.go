package share

import (
	"log"
	"time"

	"RSSHub/internal/domain"
)

type ShareVariables struct {
	repo   domain.Repository
	ticker *time.Ticker
	agg    domain.Aggregator
}

func NewShareVar(repo domain.Repository, agg domain.Aggregator) *ShareVariables {
	return &ShareVariables{repo: repo, agg: agg}
}

func (share *ShareVariables) UpdateShare(tickerDuration time.Duration) {
	share.ticker = time.NewTicker(tickerDuration)
	go func() {
		for {
			<-share.ticker.C
			currentInterval, err := share.repo.FetchInterval()
			if err != nil {
				log.Fatalf("Caught an error when tried to fetch the 'fetch interval': %v", err)
			}
			
			if share.agg.GetInterval() != currentInterval {
			}

			// select {
			// case :
			// 	share.FetchInterval()
			// }
		}
	}()
}
