// Package microbatch is a simple library for micro-batching events.
package microbatch

import (
	"sync"
)

type eventQueue[E any] struct {
	events []E
	mu     sync.Mutex
}

type batchCompletedMessage[R any] struct {
	batchId int
	result  R
}

// MicroBatcher is the object that manages microbatching of an event stream.
type MicroBatcher[E any, R any] struct {
	config             Config[E, R]
	eventQueue         eventQueue[E]
	stopChan           chan bool
	batchCompletedChan chan batchCompletedMessage[R]
	lastBatchId        int
	sendMu             sync.Mutex
	batchWg            sync.WaitGroup
}

// Start will create a MicroBatcher and start listening for events.
func Start[E any, R any](config Config[E, R]) *MicroBatcher[E, R] {
	return StartWithTicker(config, nil)
}

// StartWithTicker allows the caller to provide a Ticker for controlling time.
// Useful for testing.
func StartWithTicker[E any, R any](config Config[E, R], ticker Ticker) *MicroBatcher[E, R] {
	mb := &MicroBatcher[E, R]{
		config:             config,
		eventQueue:         eventQueue[E]{},
		stopChan:           make(chan bool),
		batchCompletedChan: make(chan batchCompletedMessage[R]),
	}

	if ticker == nil {
		ticker = NewRealTimeTicker()
	}

	mb.listen(ticker)
	return mb
}

// SubmitJob should be called to submit new Job objects to be batched.
func (mb *MicroBatcher[E, R]) SubmitJob(event E) error {
	mb.eventQueue.mu.Lock()
	defer mb.eventQueue.mu.Unlock()
	mb.eventQueue.events = append(mb.eventQueue.events, event)

	if len(mb.eventQueue.events) >= mb.config.MaxSize {
		mb.send()
	}

	return nil
}

// WaitForResults should be called to send results to the ResultHandler.
// This will block until Stop() is called.
func (mb *MicroBatcher[E, R]) WaitForResults() {
	for result := range mb.batchCompletedChan {
		mb.config.ResultHandler.Run(result.result, result.batchId)
	}
}

// Stop can be called if we want to stop listening for events.
func (mb *MicroBatcher[E, R]) Stop() {
	mb.stopChan <- true // stop the timer

}

func (mb *MicroBatcher[E, R]) listen(ticker Ticker) {
	tickerChannel := ticker.Start(mb.config.BatchFrequency)

	go func() {
		for {
			select {
			case <-tickerChannel:
				mb.send()
			case <-mb.stopChan:
				ticker.Stop()                // stop the timer
				mb.send()                    // send remaining events
				mb.batchWg.Wait()            // wait before closing the results channel
				close(mb.batchCompletedChan) // this will unblock WaitForResults()
				close(mb.stopChan)           // no longer needed
				return
			}
		}
	}()
}

// will send all events currently in the queue to the batch processor.
func (mb *MicroBatcher[E, R]) send() {
	mb.sendMu.Lock()
	defer mb.sendMu.Unlock()

	if len(mb.eventQueue.events) == 0 {
		return
	}

	batch := mb.eventQueue.events
	mb.eventQueue.events = make([]E, 0)
	newBatchId := mb.lastBatchId + 1
	mb.lastBatchId = newBatchId
	mb.batchWg.Add(1)

	go func() {
		defer mb.batchWg.Done()

		mb.batchCompletedChan <- batchCompletedMessage[R]{
			result:  mb.config.BatchProcessor.Run(batch, newBatchId),
			batchId: newBatchId,
		}
	}()
}
