package main

import (
	"fmt"
)

type TotalResultHandler struct{}

func (TotalResultHandler) Run(result TotalResult, batchId int) {
	if result.err != nil {
		fmt.Printf("Batch %d failed with error: %s\n", batchId, result.err)
	} else {
		fmt.Printf("Batch %d total was %d\n", batchId, result.total)
	}
}
