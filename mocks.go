package microbatch

type FakeBatchProcessor struct {
	calls [][]int
	chans map[int]chan bool
}

func NewFakeBatchProcessor() *FakeBatchProcessor {
	return &FakeBatchProcessor{calls: make([][]int, 0)}
}

func (fbp *FakeBatchProcessor) TellJobToFinish(batchId int) {
	fbp.chans[batchId] <- true
}

func (fbp *FakeBatchProcessor) Run(jobs []Job, batchId int) JobResult {
	batch := make([]int, 0)

	for _, j := range jobs {
		batch = append(batch, j.(int))
	}

	fbp.calls = append(fbp.calls, batch)
	return "some result"
}

type FakeResultsHandler struct{ calls []string }

func NewFakeResultsHandler() *FakeResultsHandler {
	return &FakeResultsHandler{calls: make([]string, 0)}
}

func (frh *FakeResultsHandler) Run(jobResult JobResult, jobId int) {
	frh.calls = append(frh.calls, jobResult.(string))
}
