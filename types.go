package main

import (
	"context"
	"log"
	"time"
)

// IOOperation represents an I/O based operation function type
type IOOperation func(context.Context, interface{}) (interface{}, error)

// BatchResult represents the result of a batch operation.
type BatchResult struct {
	Data interface{}
	Err  error
}

type ExecutorOptions struct {
	rateLimiter *AsyncLimiter

	cores      int
	batchSize  int
	timeout    time.Duration
	maxRetries int
	retryDelay time.Duration

	stopOnError         bool
	progressReportFunc  func(int)
	circuitBreakerLimit int

	logger              *log.Logger
	beforeStartHook     func()
	afterCompletionHook func()
	beforeRetryHook     func(error)

	customSchedulerFunc          func([]interface{}) []interface{}
	reportBenchmarkDuration      bool
	reportBenchmarkSequentialRun bool // If true, the benchmark will be run sequentially. please don't do unless you are testing.
}

type Option func(*ExecutorOptions)
