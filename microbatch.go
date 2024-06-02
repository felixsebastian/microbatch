// Package microbatch is a simple library for micro-batching events.
package microbatch

import (
	"sync"
	"time"
)

type Job interface{}
type JobResult interface{}
type batchProcessor func(job []Job) JobResult
type jobResultHandler func(jobResult JobResult)

// Config is necessary to start microbatching.
type Config struct {
	BatchProcessor   batchProcessor
	JobResultHandler jobResultHandler
	BatchFrequency   time.Duration
	MaxSize          int
}

type jobQueue struct {
	jobs []Job
	mu   sync.Mutex
}

// MicroBatcher is the object that manages microbatching of an event stream.
type MicroBatcher struct {
	config      Config
	jobQueue    jobQueue
	stopChan    chan bool
	resultsChan chan JobResult
	wg          sync.WaitGroup
}

// Start will create a MicroBatcher and start listening for events.
func Start(config Config) *MicroBatcher {
	mb := &MicroBatcher{
		config:      config,
		jobQueue:    jobQueue{},
		stopChan:    make(chan bool),
		resultsChan: make(chan JobResult),
	}

	go mb.listen()
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
	for result := range mb.resultsChan {
		mb.config.JobResultHandler(result)
	}
}

// Stop can be called if we want to stop listening for events.
func (mb *MicroBatcher) Stop() { mb.stopChan <- true }

func (mb *MicroBatcher) listen() {
	ticker := time.NewTicker(mb.config.BatchFrequency)

	for {
		select {
		case <-ticker.C:
			mb.send()
		case <-mb.stopChan:
			mb.send()             // send remaining events
			mb.wg.Wait()          // wait for events to finish processing
			close(mb.stopChan)    // no longer needed
			close(mb.resultsChan) // this will unblock WaitForResults()
			return
		}
	}
}

// sends all jobs currently in the queue to the batch processor.
func (mb *MicroBatcher) send() {
	if len(mb.jobQueue.jobs) == 0 {
		return
	}

	batch := mb.jobQueue.jobs
	mb.jobQueue.jobs = make([]Job, 0)
	mb.wg.Add(1)

	go func() {
		defer mb.wg.Done()
		mb.resultsChan <- mb.config.BatchProcessor(batch)
	}()
}
