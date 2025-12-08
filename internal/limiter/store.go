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
