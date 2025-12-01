package address

import (
	"net"
	"net/http"
	"strings"

	"github.com/piheta/seq.re/config"
)

type AddressService struct {
}

func NewAddressService() *AddressService {
	return &AddressService{}
}

func (s *AddressService) GetClientIP(r *http.Request) Address {
	behindProxy := config.Config.BehindProxy

	if behindProxy {
		// Behind reverse proxy - trust proxy headers
		// Check X-Forwarded-For first (can be comma-separated list)
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			return Address{IP: strings.TrimSpace(ips[0])} // First IP is the original client
		}

		// Check X-Real-IP
		if xri := r.Header.Get("X-Real-IP"); xri != "" {
			return Address{IP: xri}
		}
	}

	// Direct connection or proxy headers not available
	// Use RemoteAddr (cannot be spoofed)
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return Address{IP: host}
}
