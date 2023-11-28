package main

import (
	"errors"
	"sync"
	"time"
)

type AsyncLimiter struct {
	maxRate    float64       // Maximum rate of operations
	timePeriod time.Duration // Time period for the rate limit
	ratePerSec float64       // Calculated rate per second
	level      float64       // Current level of the bucket
	lastCheck  time.Time     // Last time the bucket was checked/leaked
	mu         sync.Mutex    // Mutex to protect concurrent access
	waiters    chan bool     // Channel to signal waiting goroutines
}

func NewAsyncLimiter(maxRate float64, timePeriod time.Duration) *AsyncLimiter {
	return &AsyncLimiter{
		maxRate:    maxRate,
		timePeriod: timePeriod,
		ratePerSec: maxRate / timePeriod.Seconds(),
		level:      0.0,
		lastCheck:  time.Now(),
		waiters:    make(chan bool, 1),
	}
}

func (l *AsyncLimiter) leak() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.level > 0 {
		elapsed := time.Since(l.lastCheck).Seconds()
		decrement := elapsed * l.ratePerSec
		l.level = max(0, l.level-decrement)
	}
	l.lastCheck = time.Now()
}

func (l *AsyncLimiter) hasCapacity(amount float64) bool {
	l.leak()
	requested := l.level + amount
	if requested <= l.maxRate {
		select {
		case l.waiters <- true:
		default:
		}
		return true
	}
	return false
}

func (l *AsyncLimiter) Acquire(amount float64) error {
	if amount > l.maxRate {
		return errors.New("cannot acquire more than the maximum capacity")
	}

	for !l.hasCapacity(amount) {
		select {
		case <-l.waiters:
		case <-time.After(time.Duration(1/l.ratePerSec*amount) * time.Second):
		}
	}

	l.mu.Lock()
	l.level += amount
	l.mu.Unlock()

	return nil
}
