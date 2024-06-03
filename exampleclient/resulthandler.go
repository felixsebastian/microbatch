package main

import (
	"fmt"

	"github.com/felixsebastian/microbatch"
)

type TotalResultHandler struct{}

func (TotalResultHandler) Run(result microbatch.JobResult, batchId int) {
	totalResult := result.(TotalResult)

	if totalResult.err != nil {
		fmt.Printf("Batch %d failed with error: %s\n", batchId, totalResult.err)
	} else {
		fmt.Printf("Batch %d total was %d\n", batchId, totalResult.total)
	}
}
