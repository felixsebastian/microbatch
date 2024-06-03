package microbatch

// BatchProcessor needs to be implemented, this will be called with each batch
// of events. Each batch will be called from it's own goroutine, so be careful
// when sharing state with BatchProcessor.
type BatchProcessor[E any, R any] interface {
	Run(batch []E) R
}

// ResultHandler will be called after each batch is done processing. This is
// called from the same goroutine as WaitForResults(). Use this if you don't
// want to think too much about parallelism.
type ResultHandler[R any] interface {
	Run(result R)
}

// Config is required to call Start(), to start the MicroBatcher.
type Config[E any, R any] struct {
	BatchProcessor BatchProcessor[E, R]
	ResultHandler  ResultHandler[R]
	BatchFrequency int // usually millseconds
	MaxSize        int
}
