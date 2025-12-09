package limiter

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func Middleware(store *Store, rate, capacity float64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := ExtractKey(r) // IP o API key
			bucket := store.GetOrCreate(key, rate, capacity)

			if !bucket.Allow() {
				// Rate limited
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", strconv.Itoa(bucket.RetryAfter()))
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"error":       "rate_limited",
					"retry_after": bucket.RetryAfter(),
				})
				return
			}

			// Permitido → pasar al siguiente handler (proxy)
			next.ServeHTTP(w, r)
		})
	}
}
