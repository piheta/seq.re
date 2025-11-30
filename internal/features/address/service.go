package address

import (
	"net"
	"net/http"
	"strings"
)

type AddressService struct {
}

func NewAddressService() *AddressService {
	return &AddressService{}
}

func (s *AddressService) GetClientIP(r *http.Request) Address {
	address := Address{""}

	// Check X-Forwarded-For first (can be comma-separated list)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		address.IP = strings.TrimSpace(ips[0]) // First IP is the original client
		return address
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		address.IP = xri
		return address
	}

	// Fall back to RemoteAddr
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	address.IP = host

	return address
}
