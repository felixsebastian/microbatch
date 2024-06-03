package microbatch

import "time"

type realTimeTicker struct {
	systemTicker time.Ticker
}

func NewRealTimeTicker(frequency int) Ticker {
	return &realTimeTicker{
		systemTicker: *time.NewTicker(time.Duration(frequency) * time.Millisecond),
	}
}

func (rtt *realTimeTicker) Stop() {
	rtt.systemTicker.Stop()
}

func (rtt *realTimeTicker) GetChannel() <-chan time.Time {
	return rtt.systemTicker.C
}
