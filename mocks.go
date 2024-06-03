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

func (fbp *FakeBatchProcessor) Run(batch []int, batchId int) string {
	fbp.calls = append(fbp.calls, batch)
	return "some result"
}

type FakeResultHandler struct{ calls []string }

func NewFakeResultHandler() *FakeResultHandler {
	return &FakeResultHandler{calls: make([]string, 0)}
}

func (frh *FakeResultHandler) Run(result string, eventId int) {
	frh.calls = append(frh.calls, result)
}
