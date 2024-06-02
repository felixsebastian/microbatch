package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/felixsebastian/microbatch"
)

type LetterEvent struct {
	letter string
}

type TotalResult struct {
	total int
	err   error
}

func main() {
	config := microbatch.Config{
		BatchProcessor:   letterBatchProcessor,
		JobResultHandler: totalResultHandler,
		BatchFrequency:   time.Second,
		MaxSize:          2,
	}

	mb := microbatch.Start(config)
	fmt.Println("Listening for events...")

	// schedule some random events
	scheduleLetterEvent(mb, "a", 1100*time.Millisecond)
	scheduleLetterEvent(mb, "b", 1200*time.Millisecond)
	scheduleLetterEvent(mb, "d", 1300*time.Millisecond)
	scheduleLetterEvent(mb, "a", 2100*time.Millisecond)
	scheduleLetterEvent(mb, "b", 2200*time.Millisecond)
	scheduleLetterEvent(mb, "c", 2300*time.Millisecond)
	scheduleLetterEvent(mb, "d", 2400*time.Millisecond)
	scheduleLetterEvent(mb, "a", 4900*time.Millisecond)

	// stop batching after 5 seconds
	scheduleStopEvent(mb, 5000*time.Millisecond)

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		signal.Notify(signalChan, syscall.SIGTERM)
		<-signalChan
		mb.Stop()
	}()

	mb.WaitForResults()
}

func letterBatchProcessor(batch []microbatch.Job) microbatch.JobResult {
	mappings := map[string]int{"a": 1, "b": 2, "d": 4}
	var sum int

	for _, job := range batch {
		letter := job.(LetterEvent).letter
		number, ok := mappings[letter]

		if !ok {
			return TotalResult{err: fmt.Errorf("could not map letter '%s'", letter)}
		}

		sum = sum + number
	}

	// in the real world we might need to wait for networks, disks etc
	time.Sleep(500)

	return TotalResult{total: sum}
}

func totalResultHandler(result microbatch.JobResult) {
	totalResult := result.(TotalResult)

	if totalResult.err != nil {
		fmt.Printf("Batch failed with error: %s\n", totalResult.err)
	} else {
		fmt.Printf("The total for this batch was %d\n", totalResult.total)
	}
}

func scheduleLetterEvent(mb *microbatch.MicroBatcher, letter string, wait time.Duration) {
	timer := time.NewTimer(wait)

	go func() {
		<-timer.C
		mb.SubmitJob(LetterEvent{letter: letter})
		fmt.Printf("Letter event emmited '%s'\n", letter)
	}()
}

func scheduleStopEvent(mb *microbatch.MicroBatcher, wait time.Duration) {
	timer := time.NewTimer(wait)

	go func() {
		<-timer.C
		mb.Stop()
	}()
}
