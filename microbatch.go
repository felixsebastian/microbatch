// Package microbatch is a simple library for micro-batching events.
package microbatch

import (
	"sync"
)

type jobQueue struct {
	jobs []Job
	mu   sync.Mutex
}

type batchResult struct {
	batchId   int
	jobResult JobResult
}

// MicroBatcher is the object that manages microbatching of an event stream.
type MicroBatcher struct {
	config      Config
	jobQueue    jobQueue
	stopChan    chan bool
	resultsChan chan batchResult
	lastBatchId int
	sendMu      sync.Mutex
	wg          sync.WaitGroup
}

// Start will create a MicroBatcher and start listening for events.
func Start(config Config) *MicroBatcher {
	return StartWithTicker(config, nil)
}

// StartWithTicker allows the caller to provide a Ticker for controlling time.
// Useful for testing.
func StartWithTicker(config Config, ticker Ticker) *MicroBatcher {
	mb := &MicroBatcher{
		config:      config,
		jobQueue:    jobQueue{},
		stopChan:    make(chan bool),
		resultsChan: make(chan batchResult),
	}

	if ticker == nil {
		ticker = NewRealTimeTicker()
	}

	mb.listen(ticker)
	return mb
}

// SubmitJob should be called to submit new Job objects to be batched.
func (mb *MicroBatcher) SubmitJob(job Job) error {
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
func (mb *MicroBatcher) WaitForResults() {
	for batchResult := range mb.resultsChan {
		mb.config.JobResultHandler.Run(batchResult.jobResult, batchResult.batchId)
	}
}

// Stop can be called if we want to stop listening for events.
func (mb *MicroBatcher) Stop() {
	mb.stopChan <- true // stop the timer

}

func (mb *MicroBatcher) listen(ticker Ticker) {
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
func (mb *MicroBatcher) send() {
	mb.sendMu.Lock()
	defer mb.sendMu.Unlock()

	if len(mb.jobQueue.jobs) == 0 {
		return
	}

	batch := mb.jobQueue.jobs
	mb.jobQueue.jobs = make([]Job, 0)
	newBatchId := mb.lastBatchId + 1
	mb.lastBatchId = newBatchId
	mb.wg.Add(1)

	go func() {
		defer mb.wg.Done()
		jobResult := mb.config.BatchProcessor.Run(batch, newBatchId)
		mb.resultsChan <- batchResult{jobResult: jobResult, batchId: newBatchId}
	}()
}
