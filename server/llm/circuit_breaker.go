package llm

import (
	"sync"
	"time"
)

type BreakerState int

const (
	BreakerClosed   BreakerState = iota
	BreakerOpen
	BreakerHalfOpen
)

type CircuitBreaker struct {
	mu               sync.Mutex
	state            BreakerState
	failureCount     int
	failureThreshold int
	retryAfter       time.Duration
	lastFailure      time.Time
}

func NewCircuitBreaker(threshold int, retryAfter time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            BreakerClosed,
		failureThreshold: threshold,
		retryAfter:       retryAfter,
	}
}

func (cb *CircuitBreaker) State() BreakerState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state == BreakerOpen && time.Since(cb.lastFailure) >= cb.retryAfter {
		cb.state = BreakerHalfOpen
	}
	return cb.state
}

func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = BreakerClosed
}

func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount++
	cb.lastFailure = time.Now()
	if cb.failureCount >= cb.failureThreshold {
		cb.state = BreakerOpen
	}
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failureCount = 0
	cb.state = BreakerClosed
}
