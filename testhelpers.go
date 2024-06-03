package microbatch

import (
	"fmt"
	"time"
)

type FakeBatchProcessor struct {
	calls [][]int
	chans map[int]chan bool
	slow  bool
}

func NewFakeBatchProcessor(slow bool) *FakeBatchProcessor {
	return &FakeBatchProcessor{calls: make([][]int, 0), slow: slow}
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

	if !fbp.slow {
		return "some result"
	}

	for {
		// fmt.Printf("checking batch %d\n", batchId)

		select {
		case <-fbp.chans[batchId]:
			fmt.Printf("finishing job %d", batchId)
			return "some result"
		default:
			// fmt.Println("not finishing job")
		}
	}
}

type FakeResultsHandler struct {
	calls []string
}

func NewFakeResultsHandler() *FakeResultsHandler {
	return &FakeResultsHandler{calls: make([]string, 0)}
}

func (frh *FakeResultsHandler) Run(jobResult JobResult, jobId int) {
	frh.calls = append(frh.calls, jobResult.(string))
}

type fakeTicker struct{ tickerChan *chan time.Time }

func NewFakeTicker(tickerChan *chan time.Time) Ticker {
	return &fakeTicker{
		tickerChan: tickerChan,
	}
}

func (ft *fakeTicker) Stop() {}

func (ft *fakeTicker) GetChannel() <-chan time.Time {
	return *ft.tickerChan
}
