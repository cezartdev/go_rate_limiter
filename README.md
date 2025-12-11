# Go High-Throughput Rate Limiter (Reverse Proxy)

This project implements a lightweight reverse proxy in Go that protects an upstream API using an in-memory rate limiter (Token Bucket). It demonstrates idiomatic Go concurrency, race-safety, and a minimal Docker multi-stage build producing a small runtime image.

**Goals**
- Protect API endpoints by limiting requests per client (IP or API key).
- Showcase Go concurrency primitives (goroutines, channels, mutexes) and safe shared state.
- Provide a simple, production-minded layout using only the standard library (`net/http`, `httputil`).
- Produce a minimal container image using a multi-stage build.

**Project layout**
```
cmd/
  proxy/main.go        # Rate limiter proxy entrypoint
  upstream/main.go     # Dummy upstream server for testing
internal/
  config/env.go        # Environment/config loading
  limiter/             # Token Bucket implementation + middleware
    bucket.go          # Bucket struct and Allow/RetryAfter methods
    store.go           # In-memory store with cleanup
    key_extractor.go   # Extract client key (API key or IP)
    middleware.go      # HTTP middleware
    *_test.go          # Unit tests
  proxy/proxy.go       # Reverse proxy wrapper
  httpserver/          # Server with graceful shutdown
Dockerfile.proxy       # Multi-stage build for proxy
Dockerfile.upstream    # Multi-stage build for upstream
docker-compose.yml     # Run both services together
Makefile               # Common commands (dev, build, test, etc.)
.env.development       # Development environment config
.env.production        # Production environment config
```

**High level behavior**
1. The proxy receives an incoming HTTP request.
2. A key is extracted (priority: `X-API-Key` header, fallback: client IP).
3. The limiter middleware checks the token bucket for that key.
   - If a token is available, the request is forwarded to the configured upstream.
   - If not, the proxy responds with `429 Too Many Requests` and a `Retry-After` header.

**Rate limiter design (Token Bucket)**
- Each client key has an associated bucket state: `tokens`, `capacity` (burst), `rate` (tokens/sec), and `lastRefill` timestamp.
- Refill is calculated on-demand: when a request arrives the handler computes how many tokens have been regenerated since `lastRefill` and updates the bucket atomically.
- The handler decrements one token per allowed request; if no tokens remain, it replies `429` with a `Retry-After` hint computed as `ceil((1 - tokens) / rate)` seconds.
- The implementation keeps buckets in an in-memory `map[string]*Bucket` protected by a `sync.RWMutex` and provides a periodic GC to remove stale buckets.

**Configuration (environment variables)**
| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ENV` | `development` | Environment name (loads `.env.<APP_ENV>`) |
| `BIND_ADDR` | `:8080` | Address to bind the proxy |
| `UPSTREAM_URL` | (required) | Backend URL to forward requests to |
| `RATE_LIMIT_RPS` | `5` | Tokens per second (per client key) |
| `BURST` | `10` | Maximum bucket capacity (burst size) |
| `BUCKET_TTL` | `300` | Seconds before idle buckets are removed |
| `CLEANUP_INTERVAL` | `60` | Seconds between cleanup cycles |

Example `.env.development`:
```bash
APP_ENV=development
BIND_ADDR=:8080
UPSTREAM_URL=http://localhost:9000
RATE_LIMIT_RPS=5
BURST=10
BUCKET_TTL=300
CLEANUP_INTERVAL=60
```

**Run locally**

Using Make (with hot reload via Air):
```bash
make dev
```

Or directly with Go:
```bash
# Terminal 1: Start dummy upstream
go run ./cmd/upstream

# Terminal 2: Start proxy
go run ./cmd/proxy
```

Build production binary:
```bash
make build
# Binary output in ./build/
```

**Docker**

Build images:
```bash
docker build -f Dockerfile.proxy -t go-rate-limiter:latest .
docker build -f Dockerfile.upstream -t go-rate-limiter-upstream:latest .
```

**Docker Compose (recommended)**

Start both proxy and upstream together:
```bash
docker-compose up --build
```

This starts:
- `proxy` on port 8080 (rate limiter)
- `upstream` on port 9000 (dummy backend)

Stop:
```bash
docker-compose down
```

**Testing**

```bash
# Health check (bypasses rate limiter)
curl http://localhost:8080/health

# Request through limiter
curl http://localhost:8080/

# With API key
curl -H "X-API-Key: demo" http://localhost:8080/

# Flood test (to trigger rate limiting)
for i in {1..15}; do curl -s http://localhost:8080/; echo; done
```

When rate-limited, response will look like:
```
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 1

{"error":"rate_limited","retry_after":1}
```

**Running tests**

```bash
# Run all tests
make test

# Run with race detector
make race
```

**Makefile targets**
| Target | Description |
|--------|-------------|
| `make dev` | Run with hot reload (Air) |
| `make build` | Build production binary |
| `make test` | Run unit tests |
| `make race` | Run tests with race detector |
| `make fmt` | Format code |
| `make vet` | Run go vet |

**Next steps / extensions**
- Add persistence (Redis) for distributed rate-limiting.
- Support per-route or per-plan limits (different rates per API key).
- Expose metrics (Prometheus) and more detailed logs.