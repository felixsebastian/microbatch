package main

import (
	"fmt"
	"time"
)

type LettersBatchProcessor struct{}

func (LettersBatchProcessor) Run(batch []LetterEvent) TotalResult {
	mappings := map[string]int{"a": 1, "b": 2, "d": 4}
	var sum int

	for _, letterEvent := range batch {
		number, ok := mappings[letterEvent.letter]

		if !ok {
			return TotalResult{err: fmt.Errorf("could not map letter '%s'", letterEvent.letter)}
		}

		sum = sum + number
	}

	// In the real world we might need to wait for networks, disks etc.
	time.Sleep(500)

	return TotalResult{total: sum}
}
