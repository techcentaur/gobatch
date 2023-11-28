package main

import (
	"context"
	"log"
	"time"
)

// IOOperation represents a function type for I/O-based operations.
// It takes a context and an interface{} as input, and returns an interface{} and an error.
type IOOperation func(context.Context, interface{}) (interface{}, error)

// BatchResult represents the outcome of a batch operation.
// It stores the data result and any error that occurred during the operation.
type BatchResult struct {
	Data interface{}
	Err  error
}

// ExecutorArguments defines configuration options for the Executor.
// It includes settings for rate limiting, concurrency, retries, hooks, and more.
type ExecutorArguments struct {
	rateLimiter *AsyncLimiter // Controls the rate of operations to prevent overloading.

	cores      int           // Number of cores to use for parallel processing.
	batchSize  int           // Size of each batch of operations to process.
	timeout    time.Duration // Maximum duration to wait for an operation to complete.
	maxRetries int           // Maximum number of retries for a failed operation.
	retryDelay time.Duration // Duration to wait before retrying a failed operation.

	stopOnError         bool      // Whether to stop processing on the first error encountered.
	progressReportFunc  func(int) // Function to report progress of operations.
	circuitBreakerLimit int       // Threshold for tripping the circuit breaker to stop operations.

	logger              *log.Logger // Logger for logging messages.
	beforeStartHook     func()      // Hook function to be called before starting operations.
	afterCompletionHook func()      // Hook function to be called after completing all operations.
	beforeRetryHook     func(error) // Hook function to be called before retrying a failed operation.

	customSchedulerFunc     func([]interface{}) []interface{} // Custom function to schedule operations.
	reportBenchmarkDuration bool                              // Flag to enable/disable reporting of benchmark durations.
}

// ExecutorOptions is a function type used for applying configuration options to ExecutorArguments.
type ExecutorOptions func(*ExecutorArguments)
