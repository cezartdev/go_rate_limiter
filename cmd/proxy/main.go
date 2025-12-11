package main

import (
	"log"
	"net/http"

	"github.com/cezartdev/go_rate_limiter/internal/config"
	"github.com/cezartdev/go_rate_limiter/internal/httpserver"
	"github.com/cezartdev/go_rate_limiter/internal/limiter"
	"github.com/cezartdev/go_rate_limiter/internal/proxy"
)

func main() {
	// 1. Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	log.Printf("environment: %s", cfg.Env)
	log.Printf("rate limit: %.1f req/s, burst: %.0f", cfg.RateLimitRPS, cfg.Burst)

	// 2. Create limiter store and start cleanup goroutine
	store := limiter.NewStore()
	go store.StartCleanup(cfg.BucketTTL, cfg.CleanupInterval)

	// 3. Create reverse proxy to upstream
	reverseProxy, err := proxy.NewReverseProxy(cfg.UpstreamURL)
	if err != nil {
		log.Fatalf("failed to create proxy: %v", err)
	}

	// 4. Wrap proxy with limiter middleware
	limitedProxy := limiter.Middleware(store, cfg.RateLimitRPS, cfg.Burst)(reverseProxy)

	// 5. Setup routes
	mux := http.NewServeMux()

	// Health endpoint (bypasses rate limiter)
	mux.HandleFunc("/health", proxy.HealthHandler())

	// All other requests go through limiter + proxy
	mux.Handle("/", limitedProxy)

	// 6. Create and run server with graceful shutdown
	srv := httpserver.New(cfg.BindAddr, mux)

	log.Printf("starting proxy server on %s -> upstream %s", cfg.BindAddr, cfg.UpstreamURL)

	if err := httpserver.Run(srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
