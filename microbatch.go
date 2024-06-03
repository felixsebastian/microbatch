// Package microbatch is a simple library for micro-batching events.
package microbatch

import (
	"sync"
	"time"
)

type Job interface{}
type JobResult interface{}

type BatchProcessor interface {
	Run(batch []Job, batchId int) JobResult
}

type JobResultHandler interface {
	Run(jobResult JobResult, batchId int)
}

// Config is necessary to start microbatching.
type Config struct {
	BatchProcessor   BatchProcessor
	JobResultHandler JobResultHandler
	BatchFrequency   int // usually millseconds
	MaxSize          int
}

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

// Ticker is used to control timing of batches, defaults to time.NewTicker()
type Ticker interface {
	Stop()
	GetChannel() <-chan time.Time
}

type TickerFactory func(frequency int) Ticker

// Start will create a MicroBatcher and start listening for events.
func Start(config Config) *MicroBatcher {
	return StartWithTickerFactory(config, nil)
}

// StartWithTickerFactory allows the caller to provide a tickerFactory for controlling time
func StartWithTickerFactory(config Config, tickerFactory TickerFactory) *MicroBatcher {
	mb := &MicroBatcher{
		config:      config,
		jobQueue:    jobQueue{},
		stopChan:    make(chan bool),
		resultsChan: make(chan batchResult),
	}

	if tickerFactory == nil {
		tickerFactory = NewRealTimeTicker
	}

	mb.listen(tickerFactory)
	return mb
}

// SubmitJob should be called to submit new jobs to be batched.
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
func (mb *MicroBatcher) Stop() { mb.stopChan <- true }

func (mb *MicroBatcher) listen(tickerFactory TickerFactory) {
	ticker := tickerFactory(mb.config.BatchFrequency)
	tickerChannel := ticker.GetChannel()

	go func() {
		for {
			select {
			case <-tickerChannel:
				mb.send()
			case <-mb.stopChan:
				mb.send()             // send remaining events
				mb.wg.Wait()          // wait for events to finish processing
				close(mb.stopChan)    // no longer needed
				close(mb.resultsChan) // this will unblock WaitForResults()
				return
			}
		}
	}()

	ticker.Stop()
}

// sends all jobs currently in the queue to the batch processor.
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

func (mb *MicroBatcher) WaitForJob(batchId int) {
	finishedJobs := map[int]bool{}
	var finished bool

	for !finished {
		result := <-mb.resultsChan
		finishedJobs[result.batchId] = true
		_, finished = finishedJobs[batchId]
	}
}
