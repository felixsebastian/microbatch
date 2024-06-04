package microbatch

import "sync"

type FakeBatchProcessor struct {
	calls [][]int
	mu    sync.Mutex
}

func NewFakeBatchProcessor() *FakeBatchProcessor {
	return &FakeBatchProcessor{calls: make([][]int, 0)}
}

func (fbp *FakeBatchProcessor) Run(batch []int) string {
	fbp.mu.Lock()
	defer fbp.mu.Unlock()
	fbp.calls = append(fbp.calls, batch)
	return "some result"
}

func (fbp *FakeBatchProcessor) GetCalls() [][]int {
	fbp.mu.Lock()
	defer fbp.mu.Unlock()
	return fbp.calls
}

type FakeResultHandler struct{ calls []string }

func NewFakeResultHandler() *FakeResultHandler {
	return &FakeResultHandler{calls: make([]string, 0)}
}

func (frh *FakeResultHandler) Run(result string) {
	frh.calls = append(frh.calls, result)
}
