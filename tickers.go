package microbatch

import "time"

// Ticker is used to control the timing of batches.
type Ticker interface {
	Start(frequency int) <-chan time.Time
	Stop()
}

// RealTimeTicker uses time.Ticker, and will tick every `frequency` milliseconds.
// This is usually what we want!
type RealTimeTicker struct{ systemTicker time.Ticker }

func NewRealTimeTicker() Ticker { return &RealTimeTicker{} }

func (rtt *RealTimeTicker) Start(frequency int) <-chan time.Time {
	rtt.systemTicker = *time.NewTicker(time.Duration(frequency) * time.Millisecond)
	return rtt.systemTicker.C
}

func (rtt *RealTimeTicker) Stop() { rtt.systemTicker.Stop() }

// SimpleTicker only ticks when we call the Tick() method.
// For testing purposes.
type SimpleTicker struct{ tickerChan chan time.Time }

func NewSimpleTicker() *SimpleTicker {
	return &SimpleTicker{tickerChan: make(chan time.Time)}
}

func (ft *SimpleTicker) Start(_ int) <-chan time.Time { return ft.tickerChan }
func (ft SimpleTicker) Stop()                         {}

func (ft SimpleTicker) Tick() {
	ft.tickerChan <- time.Time{}
}
