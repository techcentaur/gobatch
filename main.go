package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func MockOperation(ctx context.Context, data interface{}) error {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	//fmt.Println("data is ", data)
	return nil
}

func main() {
	logger := log.New(os.Stdout, "executor: ", log.LstdFlags)
	rateLimiter := NewAsyncLimiter(100000, 1)

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
		WithReportBenchmarkSequentialRun(true),
	}

	// Create a batch of data to process
	dataBatch := make([]interface{}, 100) // Example data
	for i := range dataBatch {
		dataBatch[i] = fmt.Sprintf("data-%d", i)
	}

	// Execute batch operation
	err := ExecuteBatchAsync(MockOperation, dataBatch, opts)
	if err != nil {
		logger.Printf("Batch execution error: %v\n", err)
	}

	return
}
