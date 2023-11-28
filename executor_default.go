package main

import (
	"log"
	"os"
	"time"
)

// Constant definitions for default settings.
const (
	DefaultCores               = 4               // Default number of CPU cores to use.
	DefaultTimeout             = 5 * time.Minute // Default timeout for operations
	DefaultMaxRetries          = 3               // Default number of retries for an operation.
	DefaultRetryDelay          = 5 * time.Second // Default delay between retries.
	DefaultBatchSize           = 10              // Default size of each batch of operations.
	DefaultCircuitBreakerLimit = 10              // Default limit for circuit breaker.
)

func NewExecutorOptions() *ExecutorArguments {
	return &ExecutorArguments{
		cores:                   DefaultCores,
		timeout:                 DefaultTimeout,
		maxRetries:              DefaultMaxRetries,
		retryDelay:              DefaultRetryDelay,
		stopOnError:             false,
		reportBenchmarkDuration: false,
		batchSize:               DefaultBatchSize,
		circuitBreakerLimit:     DefaultCircuitBreakerLimit,
		logger:                  log.New(os.Stdout, "asyncbatch: ", log.LstdFlags),
	}
}
