# Go High-Throughput Rate Limiter (Reverse Proxy)

This project implements a lightweight reverse proxy in Go that protects an upstream API using an in-memory rate limiter (Token Bucket). It demonstrates idiomatic Go concurrency, race-safety, and a minimal Docker multi-stage build producing a small runtime image.

**Goals**
- Protect API endpoints by limiting requests per client (IP or API key).
- Showcase Go concurrency primitives (goroutines, channels, mutexes) and safe shared state.
- Provide a simple, production-minded layout using only the standard library (`net/http`, `httputil`).
- Produce a minimal container image using a multi-stage build.

**Project layout (important files)**
- `cmd/api/main.go`: application entrypoint. Parses configuration, wires limiter and proxy, and starts the HTTP server.
- `internal/limiter/`: Token Bucket implementation plus in-memory bucket store and middleware.
- `internal/proxy/`: small reverse-proxy wrapper around `httputil.ReverseProxy`.
- `internal/httpserver/`: server construction and graceful shutdown helpers.
- `internal/config/`: environment/flag parsing and defaults.
- `Dockerfile`: multi-stage build to produce a minimal `scratch` image with the statically compiled binary.

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
- The implementation keeps buckets in an in-memory `map[string]*Bucket` protected by a `sync.RWMutex` (or `sync.Map`) and provides a periodic GC to remove stale buckets.
- Alternative: channel-buffered buckets with a goroutine refill loop per-bucket (demonstrated in comments/optional code) to exhibit channel-based concurrency.

**Configuration (env/flags)**
- `BIND_ADDR` (default `:8080`) — address to bind the proxy.
- `UPSTREAM_URL` — required, the backend the proxy forwards allowed requests to.
- `RATE_LIMIT_RPS` — tokens per second (per key).
- `BURST` — maximum bucket capacity (burst size).
- `BUCKET_TTL` — TTL (in seconds) for idle buckets before GC removal.

Example environment variables:

```bash
export BIND_ADDR=":8080"
export UPSTREAM_URL="http://localhost:9000"
export RATE_LIMIT_RPS=5
export BURST=10
export BUCKET_TTL=300
```

**Run locally**

Build and run with Go (fast iteration):

```bash
go run ./cmd/api
```

Build production binary and run:

```bash
go build -o bin/proxy ./cmd/api
./bin/proxy
```

**Docker (multi-stage)**

Build image:

```bash
docker build -t go-rate-limiter:latest .
```

Run container:

```bash
docker run -e UPSTREAM_URL="http://host.docker.internal:9000" -p 8080:8080 go-rate-limiter:latest
```

Notes: the Dockerfile uses a build stage with `golang` and produces a small runtime image (e.g., `scratch` or `distroless`) containing only the statically linked binary.

**Examples (curl)**

Allow request (assuming key or IP has tokens):

```bash
curl -i -H "X-API-Key: demo" http://localhost:8080/some-path
```

When rate-limited, response will look like:

```
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 3

{"error":"rate_limited","retry_after":3}
```

**Testing and exposing race conditions**
- The repository includes unit tests for the limiter logic (table-driven) and concurrency tests that can be run with the race detector:

```bash
go test ./... -race
```

**What this demonstrates**
- Idiomatic use of Go for high-concurrency components (goroutines, channels, mutexes).
- Handling of race conditions and atomic/bounded state updates.
- A practical middleware/reverse-proxy that can be extended to persistent stores, distributed rate-limiting, or more complex policies.

**Next steps / extensions**
- Add persistence (Redis) for distributed rate-limiting.
- Support per-route or per-plan limits (different rates per API key).
- Expose metrics (Prometheus) and more detailed logs.