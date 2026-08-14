package ratelimit

import "time"

// Limiter is a simple token-bucket rate limiter.
// It allows up to burst requests per second, then blocks until tokens replenish.
type Limiter struct {
	tokens chan struct{}
}

// New creates a new rate limiter that allows up to burst requests per second.
func New(burst int) *Limiter {
	l := &Limiter{
		tokens: make(chan struct{}, burst),
	}
	// Fill the bucket initially
	for i := 0; i < burst; i++ {
		l.tokens <- struct{}{}
	}
	// Replenish tokens at the specified rate
	go l.refill(burst)
	return l
}

// Acquire blocks until a token is available.
func (l *Limiter) Acquire() {
	<-l.tokens
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired.
func (l *Limiter) TryAcquire() bool {
	select {
	case <-l.tokens:
		return true
	default:
		return false
	}
}

func (l *Limiter) refill(burst int) {
	// Replenish one token every (1000/burst) milliseconds
	interval := time.Duration(1000/burst) * time.Millisecond
	if interval < 1*time.Millisecond {
		interval = 1 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		select {
		case l.tokens <- struct{}{}:
		default:
			// bucket full, skip
		}
	}
}
