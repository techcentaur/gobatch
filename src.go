package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

// IOOperation represents an I/O based operation function type
type IOOperation func(context.Context, interface{}) (interface{}, error)

// BatchResult represents the result of a batch operation.
type BatchResult struct {
	Data interface{}
	Err  error
}

type executorOptions struct {
	cores               int
	maxGoroutines       int
	timeout             time.Duration
	maxRetries          int
	retryDelay          time.Duration
	concurrencyLimit    map[string]int // key: category, value: limit
	stopOnError         bool
	maxResources        int
	batchSize           int
	progressReportFunc  func(int) // int: number of tasks completed
	circuitBreakerLimit int
	logger              *log.Logger
	beforeStartHook     func()
	afterCompletionHook func()
	beforeRetryHook     func(error)
	customSchedulerFunc func([]interface{}) []interface{}
}

type Option func(*executorOptions)

func WithCores(cores int) Option {
	return func(opt *executorOptions) {
		opt.cores = cores
	}
}

func WithMaxGoroutines(maxGoroutines int) Option {
	return func(opt *executorOptions) {
		opt.maxGoroutines = maxGoroutines
	}
}

// ExecuteBatchAsync asynchronously performs the given I/O operations on the provided data.
// It returns a channel of BatchResult.
func ExecuteBatchAsync(ctx context.Context, operationFunc IOOperation, dataBatch []interface{}, opts ...Option) (<-chan BatchResult, error) {
	// Default configurations
	conf := &executorOptions{
		cores:         runtime.NumCPU(),
		maxGoroutines: 100,
		timeout:       time.Minute * 5,
		maxRetries:    3,
		retryDelay:    time.Second * 5,
		stopOnError:   false,
		maxResources:  50,
		batchSize:     10,
		logger:        log.New(os.Stdout, "asyncbatch: ", log.LstdFlags),
		// Add defaults for other configurations
	}

	// Apply provided options
	for _, o := range opts {
		o(conf)
	}

	// Validate input
	if conf.cores <= 0 || conf.maxGoroutines <= 0 {
		return nil, errors.New("number of cores and goroutines must be greater than 0")

	}

	// Apply custom scheduler if provided
	if conf.customSchedulerFunc != nil {
		dataBatch = conf.customSchedulerFunc(dataBatch)
	}

	// If there's a before start hook, execute it
	if conf.beforeStartHook != nil {
		conf.beforeStartHook()
	}

	// Set the number of CPU cores to use
	runtime.GOMAXPROCS(conf.cores)

	semaphore := make(chan struct{}, conf.maxGoroutines)
	resultsChannel := make(chan BatchResult, len(dataBatch))
	var wg sync.WaitGroup
	var errorsCount int

	for _, data := range dataBatch {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(d interface{}) {
			defer wg.Done()
			defer func() { <-semaphore }()

			var result interface{}
			var err error

			opCtx, cancel := context.WithTimeout(ctx, conf.timeout)
			defer cancel()

			for i := 0; i <= conf.maxRetries; i++ {
				result, err = operationFunc(opCtx, d)
				if err == nil || conf.retryDelay <= 0 {
					break
				}

				// If there's a before retry hook, execute it
				if conf.beforeRetryHook != nil {
					conf.beforeRetryHook(err)
				}

				time.Sleep(conf.retryDelay)
			}

			resultsChannel <- BatchResult{
				Data: result,
				Err:  err,
			}

			if err != nil {
				errorsCount++
				if conf.stopOnError {
					return
				}
				if errorsCount > conf.circuitBreakerLimit && conf.circuitBreakerLimit > 0 {
					cancel()
				}
			}

			if conf.progressReportFunc != nil {
				conf.progressReportFunc(1)
			}
		}(data)
	}

	go func() {
		wg.Wait()
		close(resultsChannel)

		// If there's an after completion hook, execute it
		if conf.afterCompletionHook != nil {
			conf.afterCompletionHook()
		}
	}()

	return resultsChannel, nil
}

func ReadFileOperation(filename interface{}) (interface{}, error) {
	// Simulate reading a file (just returning its name for this example).
	return filename, nil
}

func main() {
	data := []interface{}{"file1.txt", "file2.txt", "file3.txt"}
	results, err := ExecuteBatchAsync(ReadFileOperation, data, WithCores(4), WithMaxGoroutines(50))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	for res := range results {
		if res.Err != nil {
			fmt.Println("Error:", res.Err)
		} else {
			fmt.Println("Data:", res.Data)
		}
	}
}
