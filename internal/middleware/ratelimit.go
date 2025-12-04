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

// RateLimit wraps a handler with rate limiting based on IP address.
// requestsPerSecond is the rate limit, burst is the maximum burst size.
func RateLimit(requestsPerSecond int, burst int, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := shared.GetIP(r)

		mu.Lock()
		v, exists := visitors[ip]
		if !exists {
			limiter := rate.NewLimiter(rate.Limit(requestsPerSecond), burst)
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

		handler.ServeHTTP(w, r)
	})
}
