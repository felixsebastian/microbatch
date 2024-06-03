package main

import (
	"fmt"
)

type TotalResultHandler struct{}

func (TotalResultHandler) Run(totalResult TotalResult) {
	if totalResult.err != nil {
		fmt.Printf("Batch failed with error: %s\n", totalResult.err)
	} else {
		fmt.Printf("Batch total was %d\n", totalResult.total)
	}
}
