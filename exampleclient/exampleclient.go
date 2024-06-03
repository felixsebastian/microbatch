package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
		BatchProcessor:   LettersBatchProcessor{},
		JobResultHandler: TotalResultHandler{},
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
