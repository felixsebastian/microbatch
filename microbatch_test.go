package microbatch

import (
	"reflect"
	"testing"
)

func TestSimpleBatch(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor()
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          10,
	}

	simpleTicker := NewSimpleTicker()
	mb := StartWithTickerFactory(config, simpleTicker)

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	simpleTicker.Tick()

	// We need to stop because this is the last opportunity before blocking
	// indefinitely.
	mb.Stop()

	// We need to wait to make sure everything has processed before testing.
	mb.WaitForResults()

	if !reflect.DeepEqual(fakeBatchProcessor.calls[0], []int{0, 1, 2}) {
		t.Fatalf("Batch should contain the input data")
	}

	if fakeResultsHandler.calls[0] != "some result" {
		t.Fatalf("Should have called myJobResultHandler with result")
	}
}

func TestTimeCycles(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor()
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          10,
	}

	simpleTicker := NewSimpleTicker()
	mb := StartWithTickerFactory(config, simpleTicker)

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	simpleTicker.Tick()

	// should not trigger an empty batch
	simpleTicker.Tick()

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	simpleTicker.Tick()

	// We need to stop because this is the last opportunity before blocking
	// indefinitely.
	mb.Stop()

	// We need to wait to make sure everything has processed before testing.
	mb.WaitForResults()

	if len(fakeBatchProcessor.calls) != 2 {
		t.Fatalf("Should have 2 batches")
	}
}

func TestMaxSize(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor()
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          3,
	}

	simpleTicker := NewSimpleTicker()
	mb := StartWithTickerFactory(config, simpleTicker)

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	mb.SubmitJob(3)
	mb.SubmitJob(4)
	mb.SubmitJob(5)

	mb.SubmitJob(6)

	// We need to stop because this is the last opportunity before blocking
	// indefinitely.
	mb.Stop()

	// We need to wait to make sure everything has processed before testing.
	mb.WaitForResults()

	if len(fakeBatchProcessor.calls) != 3 {
		t.Fatalf("Should have hit the maxSize limit twice and batched the remaining job")
	}
}
