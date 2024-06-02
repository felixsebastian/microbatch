# Microbatch

A simple go library for micro-batching event streams.

## What is microbatching anyway?

Microbatching is a simple technique used in **stream processing** pipelines to group events into small batches.

A nice property of stream processing is that events are processed in real-time. The problem is, it can become too expensive to process every event individually. Grouping events into small batches maintains the real-time properties of stream processing, but reduces some of the overhead.

There are some examples of microbatching in the GraphQL ecosystem in which clients might batch together requests to reduce network round-trips. Similarly, the server can batch together database calls to reduce the number of queries.

## API

The API for this library is simple. First start the batcher with a config object. Then, send events to `mb` with `mb.SubmitJob(someEvent)`. Events are domain specific. Finally, call `mb.WaitForResults()` to wait for results to come back. `JobResultHandler` will be called with the results of each batch.

```
config := microbatcher.Config{
  BatchProcessor:   someFunc,    // some function to process each batch
  JobResultHandler: someFunc,    // some function to process batch result
  BatchFrequency:   time.Second, // we'll send events to BatchProcessor at this interval
  MaxSize:          30,          // if this limit is reached, we'll send early
}

mb := microbatcher.Start(config)

// would normally happen in response to a real event
mb.SubmitJob(someEvent)

mb.WaitForResults()
```

`BatchProcessor` will be run in its own thread. The `JobResultHandler` is run from whatever thread `mb.WaitForResults()` is called from, so should be a safe place to pass data back to your application.
