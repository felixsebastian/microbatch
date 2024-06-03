package main

import (
	"fmt"
	"time"

	"github.com/felixsebastian/microbatch"
)

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
