package microbatch

// BatchProcessor needs to be implemented, this will be called with each batch
// of events. Each batch will be called from it's own goroutine, so be careful
// when sharing state with BatchProcessor.
type BatchProcessor[J any, JR any] interface {
	Run(batch []J, batchId int) JR
}

// JobResultHandler will be called after each batch is done processing. This is
// called from the same goroutine as WaitForResults(). Use this if you don't
// want to think too much about parallelism.
type JobResultHandler[JR any] interface {
	Run(jobResult JR, batchId int)
}

// Config is required to call Start(), to start the MicroBatcher.
type Config[J any, JR any] struct {
	BatchProcessor   BatchProcessor[J, JR]
	JobResultHandler JobResultHandler[JR]
	BatchFrequency   int // usually millseconds
	MaxSize          int
}
