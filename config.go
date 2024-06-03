package microbatch

// Job represents an event submitted with SubmitJob(), use your own domain
// specific event type for this and assert this type in BatchProcessor.
type Job interface{}

// JobResult represents the result of a **batch** of Job events (confusingly).
// It will contain whatever value is returned from BatchProcessor. Similarly,
// assert your own domain specific result type in JobResultHandler.
type JobResult interface{}

// BatchProcessor needs to be implemented, this will be called with each batch
// of events. Each batch will be called from it's own goroutine, so be careful
// when sharing state with BatchProcessor.
type BatchProcessor interface {
	Run(batch []Job, batchId int) JobResult
}

// JobResultHandler will be called after each batch is done processing. This is
// called from the same goroutine as WaitForResults(). Use this if you don't
// want to think too much about parallelism.
type JobResultHandler interface {
	Run(jobResult JobResult, batchId int)
}

// Config is required to call Start(), to start the MicroBatcher.
type Config struct {
	BatchProcessor   BatchProcessor
	JobResultHandler JobResultHandler
	BatchFrequency   int // usually millseconds
	MaxSize          int
}
