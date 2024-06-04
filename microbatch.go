// Package microbatch is a simple library for micro-batching jobs.
package microbatch

import (
	"sync"
)

type jobQueue[J any] struct {
	slice []J
	mu    sync.Mutex
}

// MicroBatcher is the object that manages microbatching of an job stream.
type MicroBatcher[J any, R any] struct {
	config     Config[J, R]
	jobQueue   jobQueue[J]
	stopChan   chan bool
	resultChan chan R
	batchWg    sync.WaitGroup
}

// Start will create a MicroBatcher and start listening for jobs.
func Start[J any, R any](config Config[J, R]) *MicroBatcher[J, R] {
	return StartWithTicker(config, nil)
}

// StartWithTicker allows the caller to provide a Ticker for controlling time.
// Useful for testing.
func StartWithTicker[J any, R any](config Config[J, R], ticker Ticker) *MicroBatcher[J, R] {
	mb := &MicroBatcher[J, R]{
		config:     config,
		jobQueue:   jobQueue[J]{},
		stopChan:   make(chan bool),
		resultChan: make(chan R),
	}

	if ticker == nil {
		ticker = NewRealTimeTicker()
	}

	mb.listen(ticker)
	return mb
}

// SubmitJob will submit a new job to be batched.
func (mb *MicroBatcher[J, R]) SubmitJob(job J) {
	mb.jobQueue.mu.Lock()
	defer mb.jobQueue.mu.Unlock()
	mb.jobQueue.slice = append(mb.jobQueue.slice, job)
	maxReached := len(mb.jobQueue.slice) >= mb.config.MaxSize

	if maxReached {
		mb.send(false)
	}
}

// WaitForResults should be called to send results to the ResultHandler.
// This will block until Stop() is called.
func (mb *MicroBatcher[J, R]) WaitForResults() {
	for result := range mb.resultChan {
		mb.config.ResultHandler.Run(result)
	}
}

// Stop will trigger the stop sequence and stop listening for new jobs.
// Once this is complete, WaitForResults() will unblock.
func (mb *MicroBatcher[J, R]) Stop() {
	mb.stopChan <- true // stop the timer
}

func (mb *MicroBatcher[J, R]) listen(ticker Ticker) {
	tickerChannel := ticker.Start(mb.config.Frequency)

	go func() {
		for {
			select {
			case <-tickerChannel:
				mb.send(true)
			case <-mb.stopChan:
				ticker.Stop()        // stop the timer
				mb.send(true)        // send remaining jobs
				mb.batchWg.Wait()    // wait before closing the results channel
				close(mb.resultChan) // this will unblock WaitForResults()
				close(mb.stopChan)   // no longer needed
				return
			}
		}
	}()
}

// Sends all jobs currently in the queue to the batch processor.
func (mb *MicroBatcher[J, R]) send(lock bool) {
	if lock {
		mb.jobQueue.mu.Lock()
		defer mb.jobQueue.mu.Unlock()
	}

	if len(mb.jobQueue.slice) == 0 {
		return
	}

	batch := mb.jobQueue.slice
	mb.jobQueue.slice = make([]J, 0)
	mb.batchWg.Add(1)

	go func() {
		defer mb.batchWg.Done()
		mb.resultChan <- mb.config.BatchProcessor.Run(batch)
	}()
}
