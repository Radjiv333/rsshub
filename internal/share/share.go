package share

import "time"

type ShareVariables struct {
	ticker   *time.Ticker
	interval time.Duration
}

func NewShareVar(interval time.Duration) *ShareVariables {
	return &ShareVariables{interval: interval}
}

func (share *ShareVariables) UpdateShare(tickerDuration time.Duration) {
	share.ticker = time.NewTicker(tickerDuration)
	
}

func (share *ShareVariables) FetchInterval() {
		
}
