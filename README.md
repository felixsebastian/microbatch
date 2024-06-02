# Microbatch

A simple go library for micro-batching event streams.

## What is microbatching anyway?

Microbatching is a simple technique used in **stream processing** pipelines to group events into small batches.

A nice property of stream processing is that events are processed in real-time. The problem is, sometimes if we have a lot of events it becomes too expensive to process every event individually. Grouping events into small batches maintains the real-time properties of stream processing, but reduces some of the overhead.

There are some good examples of microbatching in the GraphQL ecosystem in which clients often batch together requests to reduce network round-trips. Similarly, the server can batch together database calls to reduce the number of queries.

## API

The API for this library is simple. We can start the batcher with a config object like this:

```
config := microbatcher.Config{
  BatchProcessor:   someFunc,    // some function to process each batch
  JobResultHandler: someFunc,    // some function to process batch result
  BatchFrequency:   time.Second, // we'll send events at this interval
  MaxSize:          2,           // if this is reached, we'll send early
}

mb := microbatcher.Start(config)
```

We can then send events to `mb` with `mb.SubmitJob(someEvent)`. Events are domain specific.

Finally we should call `mb.WaitForResults()` to wait for results to come back. `JobResultHandler` will be called with the results of each batch.

BatchProcessor will be run in its own thread. The JobResultHandler is run from whatever thread `mb.WaitForResults()` is called from, so should be a safe place to pass data back to your application.
