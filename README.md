# Microbatch

A simple go library for micro-batching event streams.

## What is microbatching anyway?

Microbatching is a simple technique used in **stream processing** pipelines to group events into small batches.

A nice property of stream processing is that events are processed in real-time. The problem is, it can become too expensive to process every event individually. Grouping events into small batches maintains the real-time properties of stream processing, but reduces some of the overhead.

This library is designed for use cases where batches of events such as network calls, writes to disk, user input etc; are processed which can be slow. A batch may or may not return a result later that needs to be handled.

## Using this library

To start microbatching, first start the batcher with the necessary config object. This includes a `BatchProcessor` and `ResultHandler` which need to be defined.

```
type jobsProcessor struct{}
type jobResultHandler struct{}

func (jobsProcessor) Run(batch []Job) JobResult {
  // domain specific logic
}

func (jobResultHandler) Run(result JobResult) {
  // domain specific logic
}

config := microbatcher.Config[Job, JobResult]{
  BatchProcessor: jobsProcessor{},    // for processing each batch
  ResultHandler:  jobResultHandler{}, // for processing each batch result
  Frequency:      1000,               // number of milliseconds between each batch
  MaxSize:        30,                 // if this limit is reached, we'll batch early
}

mb := microbatcher.Start(config)
```

Then, start recording events with `mb.RecordEvent(job)`. Events are domain specific, called `Job` in this example but can be of any type.

Finally, call `mb.WaitForResults()` to wait for results to come back. `ResultHandler` will be called with the results of each batch, in this case `JobResult`.

`BatchProcessor` will be run in a new goroutine for each batch. The `ResultHandler` is run from whatever thread `mb.WaitForResults()` is called from, so should be a safe place to pass data back to your application.

## How does it work?

At the end of each cycle (according to `Frequency`) any new events will be sent to the `BatchProcessor`. Additionally, if the `MaxSize` limit is reached, events will be sent early. Note that this doesn't affect the time until the next cycle. Each batch is processed in it's own goroutine so that long running batches don't block each other. Here is a simple visualisation:

![335857194-2fb682e5-baea-40ba-869d-4769fb987138](https://github.com/felixsebastian/microbatch/assets/30063980/455f2e0c-bd82-4040-aba0-98ac529f2903)

## Example client

To run the example, run `cd ./exampleclient; go run .`. [See here](https://github.com/felixsebastian/microbatch/tree/main/exampleclient) for more information about the example application.
