package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBucketAllowBurst(t *testing.T) {
	b := &Bucket{
		tokens:     5,
		capacity:   5,
		rate:       1,
		lastRefill: time.Now(),
	}

	for i := 0; i < 5; i++ {
		if !b.Allow() {
			t.Fatalf("expected allow at iteration %d", i)
		}
	}

	if b.Allow() {
		t.Fatalf("expected deny after consuming burst")
	}
}

func TestBucketRefill(t *testing.T) {
	now := time.Now()
	b := &Bucket{
		tokens:     0,
		capacity:   5,
		rate:       2,                                 // 2 tokens/sec
		lastRefill: now.Add(-1500 * time.Millisecond), // 1.5s -> ~3 tokens
	}

	if !b.Allow() {
		t.Fatalf("expected allow after refill")
	}
}

func TestRetryAfterCalculation(t *testing.T) {
	b := &Bucket{
		tokens:     0.3,
		capacity:   5,
		rate:       2, // tokens/sec
		lastRefill: time.Now(),
	}

	ra := b.RetryAfter()
	if ra != 1 {
		t.Fatalf("expected retry-after 1 second, got %d", ra)
	}
}

func TestBucketConcurrency(t *testing.T) {
	b := &Bucket{
		tokens:     1,
		capacity:   1,
		rate:       0,
		lastRefill: time.Now(),
	}

	var wg sync.WaitGroup
	var success int32
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b.Allow() {
				atomic.AddInt32(&success, 1)
			}
		}()
	}
	wg.Wait()

	if atomic.LoadInt32(&success) != 1 {
		t.Fatalf("expected exactly 1 successful Allow, got %d", success)
	}
}
