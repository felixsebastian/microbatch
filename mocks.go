package microbatch

type FakeBatchProcessor struct {
	calls [][]int
}

func NewFakeBatchProcessor() *FakeBatchProcessor {
	return &FakeBatchProcessor{calls: make([][]int, 0)}
}

func (fbp *FakeBatchProcessor) Run(batch []int) string {
	fbp.calls = append(fbp.calls, batch)
	return "some result"
}

type FakeResultHandler struct{ calls []string }

func NewFakeResultHandler() *FakeResultHandler {
	return &FakeResultHandler{calls: make([]string, 0)}
}

func (frh *FakeResultHandler) Run(result string) {
	frh.calls = append(frh.calls, result)
}
