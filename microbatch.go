// Package microbatch is a simple library for micro-batching events.
package microbatch

import (
	"sync"
)

type jobQueue[J any] struct {
	jobs []J
	mu   sync.Mutex
}

type batchResult[JR any] struct {
	batchId   int
	jobResult JR
}

// MicroBatcher is the object that manages microbatching of an event stream.
type MicroBatcher[J any, JR any] struct {
	config      Config[J, JR]
	jobQueue    jobQueue[J]
	stopChan    chan bool
	resultsChan chan batchResult[JR]
	lastBatchId int
	sendMu      sync.Mutex
	wg          sync.WaitGroup
}

// Start will create a MicroBatcher and start listening for events.
func Start[J any, JR any](config Config[J, JR]) *MicroBatcher[J, JR] {
	return StartWithTicker(config, nil)
}

// StartWithTicker allows the caller to provide a Ticker for controlling time.
// Useful for testing.
func StartWithTicker[J any, JR any](config Config[J, JR], ticker Ticker) *MicroBatcher[J, JR] {
	mb := &MicroBatcher[J, JR]{
		config:      config,
		jobQueue:    jobQueue[J]{},
		stopChan:    make(chan bool),
		resultsChan: make(chan batchResult[JR]),
	}

	if ticker == nil {
		ticker = NewRealTimeTicker()
	}

	mb.listen(ticker)
	return mb
}

// SubmitJob should be called to submit new Job objects to be batched.
func (mb *MicroBatcher[J, JR]) SubmitJob(job J) error {
	mb.jobQueue.mu.Lock()
	defer mb.jobQueue.mu.Unlock()
	mb.jobQueue.jobs = append(mb.jobQueue.jobs, job)

	if len(mb.jobQueue.jobs) >= mb.config.MaxSize {
		mb.send()
	}

	return nil
}

// WaitForResults should be called to send results to the JobResultHandler.
// This will block until Stop() is called.
func (mb *MicroBatcher[J, JR]) WaitForResults() {
	for batchResult := range mb.resultsChan {
		mb.config.JobResultHandler.Run(batchResult.jobResult, batchResult.batchId)
	}
}

// Stop can be called if we want to stop listening for events.
func (mb *MicroBatcher[J, JR]) Stop() {
	mb.stopChan <- true // stop the timer

}

func (mb *MicroBatcher[J, JR]) listen(ticker Ticker) {
	tickerChannel := ticker.Start(mb.config.BatchFrequency)

	go func() {
		for {
			select {
			case <-tickerChannel:
				mb.send()
			case <-mb.stopChan:
				ticker.Stop()         // stop the timer
				mb.send()             // send remaining events
				mb.wg.Wait()          // wait before closing the results channel
				close(mb.resultsChan) // this will unblock WaitForResults()
				close(mb.stopChan)    // no longer needed
				return
			}
		}
	}()
}

// will send all jobs currently in the queue to the batch processor.
func (mb *MicroBatcher[J, JR]) send() {
	mb.sendMu.Lock()
	defer mb.sendMu.Unlock()

	if len(mb.jobQueue.jobs) == 0 {
		return
	}

	batch := mb.jobQueue.jobs
	mb.jobQueue.jobs = make([]J, 0)
	newBatchId := mb.lastBatchId + 1
	mb.lastBatchId = newBatchId
	mb.wg.Add(1)

	go func() {
		defer mb.wg.Done()
		jobResult := mb.config.BatchProcessor.Run(batch, newBatchId)
		mb.resultsChan <- batchResult[JR]{jobResult: jobResult, batchId: newBatchId}
	}()
}
