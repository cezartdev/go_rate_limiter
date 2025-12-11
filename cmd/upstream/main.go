package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// Dummy upstream server for testing the rate limiter proxy.
// This simulates a backend service that the proxy protects.

func main() {
	addr := os.Getenv("UPSTREAM_BIND")
	if addr == "" {
		addr = ":9000"
	}

	mux := http.NewServeMux()

	// Echo endpoint - returns request info
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":         "Hello from upstream!",
			"path":            r.URL.Path,
			"method":          r.Method,
			"x_forwarded_for": r.Header.Get("X-Forwarded-For"),
			"x_real_ip":       r.Header.Get("X-Real-IP"),
			"timestamp":       time.Now().Format(time.RFC3339),
		})
	})

	// Slow endpoint - simulates slow response
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "slow response completed",
		})
	})

	// Error endpoint - simulates 500 error
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "simulated upstream error",
		})
	})

	log.Printf("upstream dummy server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
