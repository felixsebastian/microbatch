# Microbatch

A simple go library for micro-batching event streams.

## What is microbatching anyway?

Microbatching is a simple technique used in **stream processing** pipelines to group events into small batches.

A nice property of stream processing is that events are processed in real-time. The problem is, it can become too expensive to process every event individually. Grouping events into small batches maintains the real-time properties of stream processing, but reduces some of the overhead.

There are some examples of microbatching in the GraphQL ecosystem in which clients might batch together requests to reduce network round-trips. Similarly, the server can batch together database calls to reduce the number of queries. Other examples of events include writes to disk, or user input events.

This library is designed for such use cases where batches of events are processed, and may or may not return a result later that needs to be handled.

## Getting started

To start microbatching, first start the batcher with the necessary config object. This includes a BatchProcessor and ResultHandler which you need to define.

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
  BatchProcessor: jobsProcessor,    // some function to process each batch
  ResultHandler:  jobResultHandler, // some function to process batch result
  BatchFrequency: time.Second,      // batches of jobs will be sent to BatchProcessor at this interval
  MaxSize:        30,               // if this limit is reached, we'll send early
}

mb := microbatcher.Start(config)
```

Then, start recording events with `RecordEvent()`. Events are domain specific, called `Job` in this example but can be any type.

```
// would normally happen in response to a real event
mb.RecordEvent(job)
```

Finally, call `mb.WaitForResults()` to wait for results to come back. `ResultHandler` will be called with the results of each batch, in this case `JobResult`.

```
// will block indefinitely until mb.Stop() is called
mb.WaitForResults()
```

`BatchProcessor` will be run in its own thread. The `ResultHandler` is run from whatever thread `mb.WaitForResults()` is called from, so should be a safe place to pass data back to your application.

## How does it work?

At the end of each cycle (according to `BatchFrequency`) any new events will be sent to the `BatchProcessor`. Additionally, if the `MaxSize` limit is reached, events will be sent early. Note that this doesn't affect the time until the next cycle. Each batch is processed in it's own thread so that long running batches don't block each other. Here is a simple vizualisation:

<img width="644" alt="image" src="https://github.com/felixsebastian/microbatch/assets/30063980/2fb682e5-baea-40ba-869d-4769fb987138">
