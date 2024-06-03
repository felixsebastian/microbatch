package main

import (
	"fmt"
	"time"

	"github.com/felixsebastian/microbatch"
)

func scheduleLetterEvent(mb *microbatch.MicroBatcher[LetterEvent, TotalResult], letter string, ms int) {
	timer := time.NewTimer(time.Duration(ms) * time.Millisecond)

	go func() {
		<-timer.C
		mb.RecordEvent(LetterEvent{letter: letter})
		fmt.Printf("Letter event emmited '%s'\n", letter)
	}()
}

func scheduleStop(mb *microbatch.MicroBatcher[LetterEvent, TotalResult], ms int) {
	timer := time.NewTimer(time.Duration(ms) * time.Millisecond)

	go func() {
		<-timer.C
		mb.Stop()
	}()
}
