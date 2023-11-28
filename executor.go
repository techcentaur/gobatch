package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

func ExecuteBatchAsync(opts *ExecutorOptions, operationFunc func(interface{}) error, dataBatch []interface{}) error {
	var wg sync.WaitGroup
	batchSize := opts.batchSize
	for i := 0; i < len(dataBatch); i += batchSize {
		end := i + batchSize
		if end > len(dataBatch) {
			end = len(dataBatch)
		}
		batch := dataBatch[i:end]

		wg.Add(1)
		go func(b []interface{}) {
			defer wg.Done()
			for _, data := range b {
				// Acquire capacity from the rate limiter
				if opts.rateLimiter != nil {
					err := opts.rateLimiter.Acquire(1) // Acquire capacity for one operation
					if err != nil {
						opts.logger.Printf("Rate limit error: %v\n", err)
						return
					}
				}

				// Execute the operation
				err := operationFunc(data)
				if err != nil {
					// handle error, implement retry logic, etc.
				}

				// Implement other features like progress report, hooks, etc.
			}
		}(batch)
	}
	wg.Wait()
	return nil
}

// attemptOperationWithRetries tries to execute the operation with retries.
func attemptOperationWithRetries(ctx context.Context, operationFunc IOOperation, data interface{}, conf *ExecutorOptions) (interface{}, error) {
	var result interface{}
	var err error

	for i := 0; i <= conf.maxRetries; i++ {
		result, err = operationFunc(ctx, data)
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
func handleErrors(errorsCount *int, err error, cancel context.CancelFunc, conf *ExecutorOptions) {
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

func WithCores(cores int) Option {
	return func(opt *ExecutorOptions) {
		opt.cores = cores
	}
}

func WithMaxGoroutines(maxBatches int) Option {
	return func(opt *ExecutorOptions) {
		opt.maxBatches = maxBatches
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(opt *ExecutorOptions) {
		opt.timeout = timeout
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(opt *ExecutorOptions) {
		opt.maxRetries = maxRetries
	}
}

func WithRetryDelay(retryDelay time.Duration) Option {
	return func(opt *ExecutorOptions) {
		opt.retryDelay = retryDelay
	}
}

func WithConcurrencyLimit(concurrencyLimit map[string]int) Option {
	return func(opt *ExecutorOptions) {
		opt.concurrencyLimit = concurrencyLimit
	}
}

func WithStopOnError(stopOnError bool) Option {
	return func(opt *ExecutorOptions) {
		opt.stopOnError = stopOnError
	}
}

func WithMaxResources(maxResources int) Option {
	return func(opt *ExecutorOptions) {
		opt.maxResources = maxResources
	}
}

func WithBatchSize(batchSize int) Option {
	return func(opt *ExecutorOptions) {
		opt.batchSize = batchSize
	}
}

func WithProgressReportFunc(progressReportFunc func(int)) Option {
	return func(opt *ExecutorOptions) {
		opt.progressReportFunc = progressReportFunc
	}
}

func WithCircuitBreakerLimit(circuitBreakerLimit int) Option {
	return func(opt *ExecutorOptions) {
		opt.circuitBreakerLimit = circuitBreakerLimit
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(opt *ExecutorOptions) {
		opt.logger = logger
	}
}

func WithBeforeStartHook(hook func()) Option {
	return func(opt *ExecutorOptions) {
		opt.beforeStartHook = hook
	}
}

func WithAfterCompletionHook(hook func()) Option {
	return func(opt *ExecutorOptions) {
		opt.afterCompletionHook = hook
	}
}

func WithBeforeRetryHook(hook func(error)) Option {
	return func(opt *ExecutorOptions) {
		opt.beforeRetryHook = hook
	}
}

func WithCustomSchedulerFunc(customSchedulerFunc func([]interface{}) []interface{}) Option {
	return func(opt *ExecutorOptions) {
		opt.customSchedulerFunc = customSchedulerFunc
	}
}

func WithRateLimiter(rateLimiter *AsyncLimiter) Option {
	return func(opt *ExecutorOptions) {
		opt.rateLimiter = rateLimiter
	}
}

// Validate checks the provided configuration for validity.
func (e *ExecutorOptions) Validate() error {
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
	if e.maxResources < 0 {
		return errors.New("maximum resources cannot be negative")
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
