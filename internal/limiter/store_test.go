package limiter

import (
	"sync"
	"testing"
	"time"
)

func TestStoreGetOrCreateConcurrency(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup
	results := make([]*Bucket, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			b := s.GetOrCreate("concurrent-key", 1, 5)
			results[i] = b
		}(i)
	}
	wg.Wait()

	first := results[0]
	for i := 1; i < len(results); i++ {
		if results[i] != first {
			t.Fatalf("expected same bucket pointer for all goroutines")
		}
	}
}

func TestStoreCleanupRemovesStale(t *testing.T) {
	s := NewStore()

	stale := &Bucket{
		tokens:     1,
		capacity:   1,
		rate:       0,
		lastRefill: time.Now().Add(-10 * time.Minute),
	}

	s.mu.Lock()
	s.buckets["old"] = stale
	s.mu.Unlock()

	s.Cleanup(5 * time.Minute)

	s.mu.RLock()
	_, exists := s.buckets["old"]
	s.mu.RUnlock()

	if exists {
		t.Fatalf("expected stale bucket to be removed by Cleanup")
	}
}
