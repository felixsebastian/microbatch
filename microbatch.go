// Package microbatch is a simple library for micro-batching events.
package microbatch

import (
	"sync"
)

type eventQueue[E any] struct {
	slice []E
	mu    sync.Mutex
}

// MicroBatcher is the object that manages microbatching of an event stream.
type MicroBatcher[E any, R any] struct {
	config     Config[E, R]
	eventQueue eventQueue[E]
	stopChan   chan bool
	resultChan chan R
	batchWg    sync.WaitGroup
}

// Start will create a MicroBatcher and start listening for events.
func Start[E any, R any](config Config[E, R]) *MicroBatcher[E, R] {
	return StartWithTicker(config, nil)
}

// StartWithTicker allows the caller to provide a Ticker for controlling time.
// Useful for testing.
func StartWithTicker[E any, R any](config Config[E, R], ticker Ticker) *MicroBatcher[E, R] {
	mb := &MicroBatcher[E, R]{
		config:     config,
		eventQueue: eventQueue[E]{},
		stopChan:   make(chan bool),
		resultChan: make(chan R),
	}

	if ticker == nil {
		ticker = NewRealTimeTicker()
	}

	mb.listen(ticker)
	return mb
}

// RecordEvent will submit a new event to be batched.
func (mb *MicroBatcher[E, R]) RecordEvent(event E) {
	mb.eventQueue.mu.Lock()
	defer mb.eventQueue.mu.Unlock()
	mb.eventQueue.slice = append(mb.eventQueue.slice, event)
	maxReached := len(mb.eventQueue.slice) >= mb.config.MaxSize

	if maxReached {
		mb.send(false)
	}
}

// WaitForResults should be called to send results to the ResultHandler.
// This will block until Stop() is called.
func (mb *MicroBatcher[E, R]) WaitForResults() {
	for result := range mb.resultChan {
		mb.config.ResultHandler.Run(result)
	}
}

// Stop will trigger the stop sequence and stop listening for new events. Once
// this is complete, WaitForResults() will unblock.
func (mb *MicroBatcher[E, R]) Stop() {
	mb.stopChan <- true // stop the timer
}

func (mb *MicroBatcher[E, R]) listen(ticker Ticker) {
	tickerChannel := ticker.Start(mb.config.Frequency)

	go func() {
		for {
			select {
			case <-tickerChannel:
				mb.send(true)
			case <-mb.stopChan:
				ticker.Stop()        // stop the timer
				mb.send(true)        // send remaining events
				mb.batchWg.Wait()    // wait before closing the results channel
				close(mb.resultChan) // this will unblock WaitForResults()
				close(mb.stopChan)   // no longer needed
				return
			}
		}
	}()
}

// will send all events currently in the queue to the batch processor.
func (mb *MicroBatcher[E, R]) send(lock bool) {
	if lock {
		mb.eventQueue.mu.Lock()
		defer mb.eventQueue.mu.Unlock()
	}

	if len(mb.eventQueue.slice) == 0 {
		return
	}

	batch := mb.eventQueue.slice
	mb.eventQueue.slice = make([]E, 0)
	mb.batchWg.Add(1)

	go func() {
		defer mb.batchWg.Done()
		mb.resultChan <- mb.config.BatchProcessor.Run(batch)
	}()
}
