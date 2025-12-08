package limiter

import (
	"sync"
	"time"
)

type Bucket struct {
	Tokens     float64
	Capacity   float64
	Rate       float64
	LastRefill time.Time
	Mu         sync.Mutex
}

func (b *Bucket) Allow() bool {

	b.Mu.Lock()
	defer b.Mu.Unlock()

	now := time.Now()

	elapsed := now.Sub(b.LastRefill).Seconds()

	tokensToAdd := elapsed * b.Rate

	b.Tokens = min(b.Capacity, b.Tokens+tokensToAdd)

	b.LastRefill = now

	if b.Tokens >= 1.0 {
		b.Tokens -= 1.0
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
