package microbatch

import (
	"reflect"
	"testing"
)

func waitForJobsToRun[J any, R any](mb *MicroBatcher[J, R]) {
	// We need to stop because this is the last opportunity before blocking
	// indefinitely.
	mb.Stop()

	// We need to wait to make sure everything has processed before testing.
	mb.WaitForResults()
}

func TestSimpleBatch(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor()
	fakeResultHandler := NewFakeResultHandler()

	config := Config[int, string]{
		BatchProcessor: fakeBatchProcessor,
		ResultHandler:  fakeResultHandler,
		Frequency:      100,
		MaxSize:        10,
	}

	simpleTicker := NewSimpleTicker()
	mb := StartWithTicker(config, simpleTicker)
	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)
	simpleTicker.Tick()
	waitForJobsToRun(mb)

	if !reflect.DeepEqual(fakeBatchProcessor.calls[0], []int{0, 1, 2}) {
		t.Fatalf("Batch should contain the input data")
	}

	if fakeResultHandler.calls[0] != "some result" {
		t.Fatalf("Should have called myResultHandler with result")
	}
}

func TestTimeCycles(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor()
	fakeResultHandler := NewFakeResultHandler()

	config := Config[int, string]{
		BatchProcessor: fakeBatchProcessor,
		ResultHandler:  fakeResultHandler,
		Frequency:      100,
		MaxSize:        10,
	}

	simpleTicker := NewSimpleTicker()
	mb := StartWithTicker(config, simpleTicker)
	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)
	simpleTicker.Tick()
	simpleTicker.Tick() // should not trigger an additional batch
	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)
	simpleTicker.Tick()
	waitForJobsToRun(mb)

	if len(fakeBatchProcessor.calls) != 2 {
		t.Fatalf("Should have 2 batches")
	}
}

func TestMaxSize(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor()
	fakeResultHandler := NewFakeResultHandler()

	config := Config[int, string]{
		BatchProcessor: fakeBatchProcessor,
		ResultHandler:  fakeResultHandler,
		Frequency:      100,
		MaxSize:        3,
	}

	simpleTicker := NewSimpleTicker()
	mb := StartWithTicker(config, simpleTicker)
	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)
	mb.SubmitJob(3)
	mb.SubmitJob(4)
	mb.SubmitJob(5)
	mb.SubmitJob(6)
	waitForJobsToRun(mb)

	if len(fakeBatchProcessor.calls) != 3 {
		t.Fatalf("Should have hit the maxSize limit twice and batched the remaining job")
	}
}
