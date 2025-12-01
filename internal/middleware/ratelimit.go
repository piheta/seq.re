package middleware

import (
	"net/http"
	"sync"

	"github.com/piheta/seq.re/internal/shared"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter *rate.Limiter
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.RWMutex
)

// RateLimit creates a rate limiting middleware that limits requests per IP address.
// It uses the token bucket algorithm from golang.org/x/time/rate.
// r is the rate (requests per second), b is the burst size.
func RateLimit(r rate.Limit, b int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip := shared.GetIP(req)

			mu.Lock()
			v, exists := visitors[ip]
			if !exists {
				limiter := rate.NewLimiter(r, b)
				visitors[ip] = &visitor{limiter: limiter}
				v = visitors[ip]
			}
			mu.Unlock()

			if !v.limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(429)
				_, _ = w.Write([]byte(`{"status":429,"type":"rate_limit","msg":"Too many requests"}`))
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
