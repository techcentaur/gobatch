
# gobatch

`gobatch` is a versatile and robust GOlang package designed for efficient and robust concurrent asynchronous batch processing. Optimized for I/O-intensive tasks, such as HTTP requests, but can be applied to CPU bounded operations also.

## Key Features

- Concurrent Processing
- Customizable Batch Size
- Sophisticated Error Handling and Retries
- Configurable Timeouts and Delays
- Resource Management
- Progress Reporting
- Extensibility and Custom Scheduling

## Installation

Get started with `gobatch` by running:

```shell
go get -u github.com/techcentaur/gobatch
```


## Executor Options

```go
	opts := []Option{
		WithCores(8),
		WithRateLimiter(rateLimiter),
		WithBatchSize(5),
		WithStopOnError(false),
		WithTimeout(5 * time.Minute),
		WithMaxRetries(5),
		WithBeforeStartHook(func() {
			logger.Println("Starting batch operation...")
		}),
		WithAfterCompletionHook(func() {
			logger.Println("Batch operation completed.")
		}),
		WithBeforeRetryHook(func(err error) {
			logger.Printf("Retrying operation due to error: %v\n", err)
		}),
		WithProgressReportFunc(
			func(numProcessed int) {}),
		WithLogger(logger),
		WithCustomSchedulerFunc(func(data []interface{}) []interface{} {
			// Custom scheduler logic
			return data
		}),
		WithRetryDelay(5 * time.Second),
		WithReportBenchmarkDuration(true),
	}
```

## Async Limiter Options

```go
NewAsyncLimiter(maxRate, timePeriodInSeconds)
// APIs that are 100 credits per minute give (100, 60)

```


## Usage Example

```go
package main

import (
    "context"
    "fmt"
    "github.com/techcentaur/gobatch"
    "log"
    "os"
    "time"
)

func main() {
    logger := log.New(os.Stdout, "executor: ", log.LstdFlags)
    rateLimiter := gobatch.NewAsyncLimiter(100, 1)
    opts := []gobatch.Option{
        gobatch.WithCores(8),
        gobatch.WithRateLimiter(rateLimiter),
        // ... other options ...
    }

    dataBatch := make([]interface{}, 100) // Example data
    for i := range dataBatch {
        // your data variables logic here
		dataBatch[i] = fmt.Sprintf("data-%d", i)
    }

    err := gobatch.ExecuteBatchAsync(
        context.Background(), 
        myOperationFunc, 
        dataBatch, 
        opts...
    )
    if err != nil {
        logger.Printf("Batch execution error: %v\n", err)
    }
}

func myOperationFunc(ctx context.Context, data interface{}) error {
    // Your operation logic here
}
```


## Contributing

We welcome contributions to `gobatch`! Whether it's bug reports, feature requests, or code contributions, your input is valuable.

1. Fork the repository.
2. Create a new branch for your feature or fix.
3. Write and test your code.
4. Submit a pull request.

## License

`gobatch` is released under the [MIT License](LICENSE).

---
