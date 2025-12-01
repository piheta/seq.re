package shared

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"reflect"

	"github.com/go-playground/validator/v10"
)

var (
	Validate *validator.Validate
)

func InitValidator() {
	Validate = validator.New()

	if err := Validate.RegisterValidation("notprivateip", ValidateIsNotPrivateIP); err != nil {
		slog.With("error", err).Error("Failed to Init notprivateip validation")
		panic(err)
	}
}

func getStringValue(v reflect.Value) string {
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		return v.Elem().String()
	}
	return v.String()
}

func ValidateIsNotPrivateIP(fl validator.FieldLevel) bool {
	value := getStringValue(fl.Field())

	if err := validateURL(value); err == nil {
		return true
	}

	return false
}

func validateURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return errors.New("invalid URL scheme: only http and https are allowed")
	}

	hostname := parsedURL.Hostname()
	if hostname == "" {
		return errors.New("URL must have a valid hostname")
	}

	if ip := net.ParseIP(hostname); ip != nil {
		if isInternalIP(ip) {
			return fmt.Errorf("URL points to internal/private IP address: %s", ip.String())
		}
		return nil
	}

	ips, dnsErr := net.LookupIP(hostname)
	if dnsErr != nil {
		// If DNS lookup fails, we allow it (network might be down, or domain might not exist yet)
		// The important check is blocking direct IP addresses and resolved IPs
		return nil // nolint:nilerr // Intentionally allowing URLs when DNS resolution fails
	}

	for _, ip := range ips {
		if isInternalIP(ip) {
			return fmt.Errorf("URL points to internal/private IP address: %s", ip.String())
		}
	}

	return nil
}

func isInternalIP(ip net.IP) bool {
	// Check if it's an IPv4 address (To4 returns non-nil for IPv4)
	if ip.To4() != nil {
		privateIPv4Ranges := []string{
			"10.0.0.0/8",         // Private network
			"172.16.0.0/12",      // Private network
			"192.168.0.0/16",     // Private network
			"127.0.0.0/8",        // Loopback
			"169.254.0.0/16",     // Link-local
			"224.0.0.0/4",        // Multicast
			"240.0.0.0/4",        // Reserved
			"0.0.0.0/8",          // Current network
			"100.64.0.0/10",      // Shared address space (RFC 6598)
			"192.0.0.0/24",       // IETF Protocol Assignments
			"192.0.2.0/24",       // TEST-NET-1
			"198.18.0.0/15",      // Benchmarking
			"198.51.100.0/24",    // TEST-NET-2
			"203.0.113.0/24",     // TEST-NET-3
			"255.255.255.255/32", // Broadcast
		}

		for _, cidr := range privateIPv4Ranges {
			_, subnet, err := net.ParseCIDR(cidr)
			if err != nil {
				continue
			}
			if subnet.Contains(ip) {
				return true
			}
		}
		return false
	}

	privateIPv6Ranges := []string{
		"::1/128",       // Loopback
		"fe80::/10",     // Link-local
		"fc00::/7",      // Unique local address
		"::/128",        // Unspecified
		"::ffff:0:0/96", // IPv4-mapped IPv6
		"100::/64",      // Discard prefix
		"2001::/32",     // TEREDO
		"2001:10::/28",  // Deprecated
		"2001:db8::/32", // Documentation
		"ff00::/8",      // Multicast
	}

	for _, cidr := range privateIPv6Ranges {
		_, subnet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if subnet.Contains(ip) {
			return true
		}
	}

	return false
}
