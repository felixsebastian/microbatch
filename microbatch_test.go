package microbatch

import (
	"reflect"
	"testing"
	"time"
)

func TestSimpleBatch(t *testing.T) {
	fakeBatchProcessor := NewFakeBatchProcessor(false)
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          10,
	}

	fakeTickerChan := make(chan time.Time)
	mb := StartWithTickerFactory(config, func(_ int) Ticker { return NewFakeTicker(&fakeTickerChan) })

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	fakeTickerChan <- time.Time{}

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
	fakeBatchProcessor := NewFakeBatchProcessor(false)
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          10,
	}

	fakeTickerChan := make(chan time.Time)
	mb := StartWithTickerFactory(config, func(_ int) Ticker { return NewFakeTicker(&fakeTickerChan) })

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	fakeTickerChan <- time.Time{}

	// should not call the processor
	fakeTickerChan <- time.Time{}

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	fakeTickerChan <- time.Time{}

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
	fakeBatchProcessor := NewFakeBatchProcessor(false)
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          3,
	}

	fakeTickerChan := make(chan time.Time)
	mb := StartWithTickerFactory(config, func(_ int) Ticker { return NewFakeTicker(&fakeTickerChan) })

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

func TestConcurrentBatches(t *testing.T) {
	return
	fakeBatchProcessor := NewFakeBatchProcessor(true)
	fakeResultsHandler := NewFakeResultsHandler()

	config := Config{
		BatchProcessor:   fakeBatchProcessor,
		JobResultHandler: fakeResultsHandler,
		BatchFrequency:   100,
		MaxSize:          10,
	}

	fakeTickerChan := make(chan time.Time)
	mb := StartWithTickerFactory(config, func(_ int) Ticker { return NewFakeTicker(&fakeTickerChan) })

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	fakeTickerChan <- time.Time{}

	mb.SubmitJob(0)
	mb.SubmitJob(1)
	mb.SubmitJob(2)

	fakeTickerChan <- time.Time{}

	// make sure it is possible for job 2 to finish before job 1
	fakeBatchProcessor.TellJobToFinish(2)
	mb.WaitForJob(2)

	fakeBatchProcessor.TellJobToFinish(1)
	mb.WaitForJob(1)

	mb.Stop()
	mb.WaitForResults()

	calls := fakeBatchProcessor.calls
	lastCall := calls[0]

	if len(lastCall) != 3 {
		t.Fatalf("Batch should have all jobs")
	}

	if !reflect.DeepEqual(lastCall, []int{0, 1, 2}) {
		t.Fatalf("Batch should contain the input data")
	}

	if fakeResultsHandler.calls[len(fakeResultsHandler.calls)-1] != "some result" {
		t.Fatalf("Should have called myJobResultHandler with result")
	}
}
