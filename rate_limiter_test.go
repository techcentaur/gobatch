package main

import (
	"sync"
	"testing"
	"time"
)

func TestAsyncLimiter(t *testing.T) {
	// Create a new AsyncLimiter with a rate of 2 operations per second
	limiter := NewAsyncLimiter(2, 1)

	// Number of goroutines to simulate
	const numGoroutines = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			// Each goroutine tries to acquire capacity from the limiter
			t.Logf("Goroutine %d attempting to acquire capacity...\n", id)
			err := limiter.Acquire(1)
			if err != nil {
				t.Logf("Goroutine %d failed to acquire capacity: %v\n", id, err)
				return
			}
			t.Logf("Goroutine %d acquired capacity. Performing operation...\n", id)

			// Simulate some work
			time.Sleep(100 * time.Millisecond)

			t.Logf("Goroutine %d completed operation.\n", id)
		}(i)
	}

	wg.Wait()
	t.Logf("All goroutines completed.")
}
