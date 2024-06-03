package microbatch

import "time"

// Ticker is used to control the timing of batches.
type Ticker interface {
	Start(frequency int) <-chan time.Time
	Stop()
}

type realTimeTicker struct{ systemTicker time.Ticker }

func NewRealTimeTicker() Ticker { return &realTimeTicker{} }

func (rtt *realTimeTicker) Start(frequency int) <-chan time.Time {
	rtt.systemTicker = *time.NewTicker(time.Duration(frequency) * time.Millisecond)
	return rtt.systemTicker.C
}

func (rtt *realTimeTicker) Stop() { rtt.systemTicker.Stop() }

type simpleTicker struct{ tickerChan chan time.Time }

func NewSimpleTicker() *simpleTicker {
	return &simpleTicker{tickerChan: make(chan time.Time)}
}

func (ft *simpleTicker) Start(_ int) <-chan time.Time {
	return ft.tickerChan
}

func (ft *simpleTicker) Stop() {}
func (ft *simpleTicker) Tick() { ft.tickerChan <- time.Time{} }
