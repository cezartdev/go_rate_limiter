package limiter

import (
	"net"
	"net/http"
	"strings"
)

func ExtractKey(r *http.Request) string {
	// Prefer explicit API key
	if k := strings.TrimSpace(r.Header.Get("X-API-Key")); k != "" {
		return k
	}

	// X-Forwarded-For may contain a comma-separated list; take the first
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if host, _, err := net.SplitHostPort(ip); err == nil {
				return host
			}
			return ip
		}
	}

	// X-Real-IP is commonly set by reverse proxies
	if xr := strings.TrimSpace(r.Header.Get("X-Real-IP")); xr != "" {
		if host, _, err := net.SplitHostPort(xr); err == nil {
			return host
		}
		return xr
	}

	// Fallback to RemoteAddr (may include port)
	if ra := strings.TrimSpace(r.RemoteAddr); ra != "" {
		if host, _, err := net.SplitHostPort(ra); err == nil {
			return host
		}
		return ra
	}

	return "unknown"
}
