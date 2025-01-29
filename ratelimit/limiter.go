package ratelimit

import (
	"sync"
	"time"
)

type attempt struct {
	count     int
	startTime time.Time
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	attempts map[string]attempt
	mu       sync.RWMutex
	// Configuration
	maxAttempts int
	window      time.Duration
}

// New creates a new rate limiter
func New(maxAttempts int, window time.Duration) *RateLimiter {
	limiter := &RateLimiter{
		attempts:    make(map[string]attempt),
		maxAttempts: maxAttempts,
		window:      window,
	}

	// Start cleanup goroutine
	go limiter.cleanup()

	return limiter
}

// Allow checks if the key is allowed to make another attempt
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	att, exists := rl.attempts[key]

	// If the key doesn't exist or the window has expired, reset the counter
	if !exists || now.Sub(att.startTime) > rl.window {
		rl.attempts[key] = attempt{
			count:     1,
			startTime: now,
		}
		return true
	}

	// If within window and under limit, increment counter
	if att.count < rl.maxAttempts {
		rl.attempts[key] = attempt{
			count:     att.count + 1,
			startTime: att.startTime,
		}
		return true
	}

	return false
}

// GetRemainingAttempts returns the number of remaining attempts and time until reset
func (rl *RateLimiter) GetRemainingAttempts(key string) (int, time.Duration) {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	now := time.Now()
	att, exists := rl.attempts[key]

	if !exists {
		return rl.maxAttempts, 0
	}

	timeElapsed := now.Sub(att.startTime)
	if timeElapsed > rl.window {
		return rl.maxAttempts, 0
	}

	remaining := rl.maxAttempts - att.count
	if remaining < 0 {
		remaining = 0
	}

	timeRemaining := rl.window - timeElapsed
	return remaining, timeRemaining
}

// cleanup periodically removes expired entries
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, att := range rl.attempts {
			if now.Sub(att.startTime) > rl.window {
				delete(rl.attempts, key)
			}
		}
		rl.mu.Unlock()
	}
}