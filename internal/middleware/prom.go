package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	prometheus.MustRegister(HTTPRequestsTotal)
	prometheus.MustRegister(HTTPRequestDuration)
}

func NewPrometheusMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(recorder, r)

			if isValidRoutePrefix(r.URL.Path) {
				path := normalizePath(r.URL.Path)
				duration := time.Since(start).Seconds()
				status := strconv.Itoa(recorder.statusCode)
				HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
				HTTPRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
			}
		})
	}
}

func isValidRoutePrefix(path string) bool {
	switch {
	case strings.HasPrefix(path, "/static"):
		return false
	case path == "/":
		return true
	case strings.HasPrefix(path, "/tab/"):
		return true
	case strings.HasPrefix(path, "/web/"):
		return true
	case strings.HasPrefix(path, "/api/"):
		return true
	case strings.HasPrefix(path, "/i/"):
		return true
	case strings.HasPrefix(path, "/s/"):
		return true
	case strings.HasPrefix(path, "/p/"):
		return true
	case len(path) == 7 && path[0] == '/':
		return true
	default:
		return false
	}
}

func normalizePath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if len(part) == 6 {
			parts[i] = "{short}"
		}
	}
	return strings.Join(parts, "/")
}
