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
		BatchProcessor:   lettersBatchProcessor{},
		JobResultHandler: totalResultHandler{},
		BatchFrequency:   1000,
		MaxSize:          2,
	}

	mb := microbatch.Start(config)
	fmt.Println("Listening for events...")

	// schedule some random events
	scheduleLetterEvent(mb, "a", 1100)
	scheduleLetterEvent(mb, "b", 1200)
	scheduleLetterEvent(mb, "d", 1300)
	scheduleLetterEvent(mb, "a", 2100)
	scheduleLetterEvent(mb, "b", 2200)
	scheduleLetterEvent(mb, "c", 2300)
	scheduleLetterEvent(mb, "d", 2400)
	scheduleLetterEvent(mb, "a", 4900)

	// stop batching after 5 seconds
	scheduleStop(mb, 5000)

	go func() {
		signalChan := make(chan os.Signal, 1)
		signal.Notify(signalChan, os.Interrupt)
		signal.Notify(signalChan, syscall.SIGTERM)
		<-signalChan
		mb.Stop()
	}()

	mb.WaitForResults()
}

type lettersBatchProcessor struct{}

func (lettersBatchProcessor) Run(batch []microbatch.Job, _ int) microbatch.JobResult {
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

type totalResultHandler struct{}

func (totalResultHandler) Run(result microbatch.JobResult, _ int) {
	totalResult := result.(TotalResult)

	if totalResult.err != nil {
		fmt.Printf("Batch failed with error: %s\n", totalResult.err)
	} else {
		fmt.Printf("The total for this batch was %d\n", totalResult.total)
	}
}

func scheduleLetterEvent(mb *microbatch.MicroBatcher, letter string, wait int) {
	timer := time.NewTimer(time.Duration(wait) * time.Millisecond)

	go func() {
		<-timer.C
		mb.SubmitJob(LetterEvent{letter: letter})
		fmt.Printf("Letter event emmited '%s'\n", letter)
	}()
}

func scheduleStop(mb *microbatch.MicroBatcher, wait int) {
	timer := time.NewTimer(time.Duration(wait) * time.Millisecond)

	go func() {
		<-timer.C
		mb.Stop()
	}()
}
