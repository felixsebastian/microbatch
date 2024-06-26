# Microbatch

A simple go library for micro-batching.

## What is microbatching anyway?

Microbatching is a technique commonly used in stream processing pipelines of grouping events into small batches to reduce the overhead of processing events individually, and for fault tolerance. It can also be used to batch together tasks like network calls or writes to disk. There are good examples of microbatching in the GraphQL ecosystem with tools like dataloader.

## Using this library

To start microbatching, call `Start()` with the necessary config object. This includes a `BatchProcessor` and `ResultHandler` which need to be defined.

```
type batchProcessor struct{}
type resultHandler struct{}

func (batchProcessor) Run(batch []Job) JobResult {
  // domain specific logic for handling a batch
}

func (resultHandler) Run(result JobResult) {
  // domain specific logic for handling any results
}

config := microbatcher.Config[Job, JobResult]{
  BatchProcessor: batchProcessor{},   // for processing each batch
  ResultHandler:  resultHandler{},    // for processing each batch result
  Frequency:      1000,               // number of milliseconds between each batch
  MaxSize:        30,                 // if this limit is reached, we'll batch early
}

mb := microbatcher.Start(config)
```

Then submit jobs with `mb.SubmitJob(job)`. Jobs are domain specific, called `Job` in this example but can be of any type.

Finally, call `mb.WaitForResults()` to wait for results to come back. `ResultHandler` will be called with the results of each batch.

`BatchProcessor` will be run in a new goroutine for each batch. The `ResultHandler` is called by `WaitForResults()`, so should be a safe place to pass data back to your application.

## How does it work?

At the end of each cycle (according to `Frequency`) any new jobs will be sent to the `BatchProcessor`. Additionally, if the `MaxSize` limit is reached, jobs will be sent early. Note that this doesn't affect the time until the next cycle. Each batch is processed in it's own goroutine so that long running batches don't block each other. Here is a simple visualisation:

![335857194-2fb682e5-baea-40ba-869d-4769fb987138](https://github.com/felixsebastian/microbatch/assets/30063980/455f2e0c-bd82-4040-aba0-98ac529f2903)

## Example client

To run the example, `cd` into `./exampleclient` and run `go run .`. [See here](https://github.com/felixsebastian/microbatch/tree/main/exampleclient) for more information about the example application.
