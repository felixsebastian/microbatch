package microbatch

// BatchProcessor needs to be implemented, this will be called with each batch of jobs.
// Each batch will be called from it's own goroutine, so be careful when sharing state with BatchProcessor.
type BatchProcessor[J any, R any] interface {
	Run(batch []J) R
}

// ResultHandler will be called after each batch is done processing.
// This is called from the same goroutine as WaitForResults().
// Use this if you don't want to think too much about parallelism.
type ResultHandler[R any] interface {
	Run(result R)
}

// Config is a required param of Start(), which starts the MicroBatcher.
type Config[J any, R any] struct {
	BatchProcessor BatchProcessor[J, R]
	ResultHandler  ResultHandler[R]
	Frequency      int // usually millseconds
	MaxSize        int
}
