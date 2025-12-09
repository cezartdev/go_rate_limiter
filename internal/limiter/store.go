package limiter

import (
	"sync"
	"time"
)

type Store struct {
	buckets map[string]*Bucket
	mu      sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		buckets: make(map[string]*Bucket),
	}
}

func (s *Store) GetOrCreate(key string, rate, capacity float64) *Bucket {
	s.mu.RLock()
	bucket, exists := s.buckets[key]
	s.mu.RUnlock()

	if exists {
		return bucket
	}

	s.mu.Lock()

	if bucket, exists := s.buckets[key]; exists {
		s.mu.Unlock()
		return bucket
	}

	bucket = &Bucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}

	s.buckets[key] = bucket
	s.mu.Unlock()

	return bucket
}

func (s *Store) Cleanup(ttl time.Duration) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    now := time.Now()
    for key, bucket := range s.buckets {
        bucket.mu.Lock()
        if now.Sub(bucket.lastRefill) > ttl {
            delete(s.buckets, key)
        }
        bucket.mu.Unlock()
    }
}

func (s *Store) StartCleanup(ttl, interval time.Duration) {
    ticker := time.NewTicker(interval)
    for range ticker.C {
        s.Cleanup(ttl)
    }
}