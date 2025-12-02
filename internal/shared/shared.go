package shared

import (
	"crypto/rand"
	"net"
	"net/http"
	"strings"

	"github.com/piheta/seq.re/config"
)

// GetIP extracts the client IP address from the HTTP request.
// If behind a reverse proxy (BEHIND_PROXY=true), it trusts X-Forwarded-For or X-Real-IP headers.
// Otherwise, it uses RemoteAddr which cannot be spoofed.
func GetIP(r *http.Request) string {
	behindProxy := config.Config.BehindProxy

	if behindProxy {
		// Behind reverse proxy - trust proxy headers
		// Check X-Forwarded-For first (can be comma-separated list)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			return strings.TrimSpace(ips[0]) // First IP is the original client
		}

		// Check X-Real-IP
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return xri
		}
	}

	// Direct connection or proxy headers not available
	// Use RemoteAddr (cannot be spoofed)
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func CreateShort() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_"
	b := make([]byte, 6)
	rand.Read(b) // nolint
	for i := range b {
		b[i] = chars[b[i]%byte(len(chars))]
	}
	return string(b)
}
