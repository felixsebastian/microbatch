package main

import (
	"fmt"
	"time"

	"github.com/felixsebastian/microbatch"
)

type LettersBatchProcessor struct{}

func (LettersBatchProcessor) Run(batch []microbatch.Job, _ int) microbatch.JobResult {
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
