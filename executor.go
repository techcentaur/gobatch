package main

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
	"time"
)

func ExecuteBatchAsync(operationFunc func(context.Context, interface{}) error, dataBatch []interface{}, opts []ExecutorOptions) error {
	cfg := NewExecutorOptions()

	// Apply provided options to override defaults
	for _, o := range opts {
		o(cfg)
	}

	// Validate configuration settings
	if err := cfg.Validate(); err != nil {
		return err
	}

	// Custom scheduler adjustment (if provided)
	if cfg.customSchedulerFunc != nil {
		dataBatch = cfg.customSchedulerFunc(dataBatch)
	}

	// Execute any 'before start' hook
	if cfg.beforeStartHook != nil {
		cfg.beforeStartHook()
	}

	// Setting maximum CPU cores
	runtime.GOMAXPROCS(cfg.cores)

	var wg sync.WaitGroup
	var errorsCount int

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	size := cfg.batchSize
	batches := (len(dataBatch) + size - 1) / size

	startTime := time.Now()

	for i := 0; i < batches; i++ {
		batch := dataBatch[i*size : min((i+1)*size, len(dataBatch))]

		wg.Add(1)
		go func(b []interface{}) {
			defer wg.Done()
			for _, j := range b {
				// Acquire capacity from the rate limiter
				if cfg.rateLimiter != nil {
					err := cfg.rateLimiter.Acquire(1) // Acquire capacity for one operation
					if err != nil {
						cfg.logger.Printf("Rate limit error: %v\n", err)
						return
					}
				}

				_, err := attemptOperationWithRetries(ctx, operationFunc, j, cfg)
				if err != nil {
					handleErrors(&errorsCount, err, cancel, cfg)
					if cfg.stopOnError {
						return
					}
				}

				if cfg.progressReportFunc != nil {
					cfg.progressReportFunc(1) // Reporting progress after each operation
				}
			}
		}(batch)
	}
	wg.Wait()

	if cfg.reportBenchmarkDuration {
		duration := time.Now().Sub(startTime)
		cfg.logger.Printf("Time benchmark to execute: %v\n", duration)
	}

	// Execute any 'after completion' hook
	if cfg.afterCompletionHook != nil {
		cfg.afterCompletionHook()
	}

	return nil
}

// attemptOperationWithRetries tries to execute the operation with retries.
func attemptOperationWithRetries(ctx context.Context, operationFunc func(context.Context, interface{}) error, data interface{}, conf *ExecutorArguments) (interface{}, error) {
	var result interface{}
	var err error

	for i := 0; i <= conf.maxRetries; i++ {
		err = operationFunc(ctx, data)
		if err == nil || conf.retryDelay <= 0 {
			break
		}
		if conf.beforeRetryHook != nil {
			conf.beforeRetryHook(err)
		}
		time.Sleep(conf.retryDelay)
	}

	return result, err
}

// handleErrors manages error counting and circuit breaker logic.
func handleErrors(errorsCount *int, err error, cancel context.CancelFunc, conf *ExecutorArguments) {
	if err != nil {
		*errorsCount++
		if conf.stopOnError {
			cancel()
		}
		if *errorsCount > conf.circuitBreakerLimit && conf.circuitBreakerLimit > 0 {
			cancel()
		}
	}
	if conf.progressReportFunc != nil {
		conf.progressReportFunc(1)
	}
}

func WithCores(cores int) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.cores = cores
	}
}

func WithTimeout(timeout time.Duration) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.timeout = timeout
	}
}

func WithMaxRetries(maxRetries int) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.maxRetries = maxRetries
	}
}

func WithRetryDelay(retryDelay time.Duration) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.retryDelay = retryDelay
	}
}

func WithStopOnError(stopOnError bool) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.stopOnError = stopOnError
	}
}

func WithBatchSize(batchSize int) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.batchSize = batchSize
	}
}

func WithProgressReportFunc(progressReportFunc func(int)) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.progressReportFunc = progressReportFunc
	}
}

func WithLogger(logger *log.Logger) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.logger = logger
	}
}

func WithBeforeStartHook(hook func()) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.beforeStartHook = hook
	}
}

func WithAfterCompletionHook(hook func()) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.afterCompletionHook = hook
	}
}

func WithBeforeRetryHook(hook func(error)) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.beforeRetryHook = hook
	}
}

func WithCustomSchedulerFunc(customSchedulerFunc func([]interface{}) []interface{}) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.customSchedulerFunc = customSchedulerFunc
	}
}

func WithRateLimiter(rateLimiter *AsyncLimiter) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.rateLimiter = rateLimiter
	}
}

func WithReportBenchmarkDuration(reportBenchmark bool) ExecutorOptions {
	return func(opt *ExecutorArguments) {
		opt.reportBenchmarkDuration = reportBenchmark
	}
}

// Validate checks the provided configuration for validity.
func (e *ExecutorArguments) Validate() error {
	if e.cores <= 0 {
		return errors.New("number of cores must be greater than 0")
	}
	if e.timeout <= 0 {
		return errors.New("timeout must be greater than 0")
	}
	if e.maxRetries < 0 {
		return errors.New("maximum retries cannot be negative")
	}
	if e.retryDelay < 0 {
		return errors.New("retry delay cannot be negative")
	}
	if e.batchSize <= 0 {
		return errors.New("batch size must be greater than 0")
	}
	if e.circuitBreakerLimit < 0 {
		return errors.New("circuit breaker limit cannot be negative")
	}
	if e.logger == nil {
		return errors.New("logger cannot be nil")
	}
	// No validation required for boolean fields, hooks, and customSchedulerFunc as they are optional and can be nil.
	return nil
}
