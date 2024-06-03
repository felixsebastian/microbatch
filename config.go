package microbatch

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
