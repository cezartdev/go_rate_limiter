package limiter

import (
	"sync"
	"time"
)

type Bucket struct {
	tokens     float64
	capacity   float64
	rate       float64
	lastRefill time.Time
	mu         sync.Mutex
}

func (b *Bucket) Allow() bool {

	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()

	elapsed := now.Sub(b.lastRefill).Seconds()

	tokensToAdd := elapsed * b.rate

	b.tokens = min(b.capacity, b.tokens+tokensToAdd)

	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}
	return false

}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
