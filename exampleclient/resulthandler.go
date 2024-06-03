package main

import (
	"fmt"
)

type TotalResultHandler struct{}

func (TotalResultHandler) Run(totalResult TotalResult, batchId int) {
	if totalResult.err != nil {
		fmt.Printf("Batch %d failed with error: %s\n", batchId, totalResult.err)
	} else {
		fmt.Printf("Batch %d total was %d\n", batchId, totalResult.total)
	}
}
