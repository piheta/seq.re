package tests

import (
	"testing"

	"github.com/piheta/seq.re/internal/features/link"
	"github.com/piheta/seq.re/internal/shared"
)

func TestPrivateIPValidation(t *testing.T) {
	shared.InitValidator()

	tests := []struct {
		name      string
		url       string
		shouldErr bool
		reason    string
	}{
		// Valid public URLs
		{
			name:      "Valid public domain",
			url:       "https://example.com",
			shouldErr: false,
			reason:    "Should allow public domains",
		},
		{
			name:      "Valid HTTP URL",
			url:       "http://google.com",
			shouldErr: false,
			reason:    "Should allow HTTP scheme",
		},
		{
			name:      "Valid HTTPS URL",
			url:       "https://github.com",
			shouldErr: false,
			reason:    "Should allow HTTPS scheme",
		},

		// RFC 1918 Private Networks
		{
			name:      "10.0.0.0/8 - start",
			url:       "http://10.0.0.1",
			shouldErr: true,
			reason:    "Should block 10.0.0.0/8 private network",
		},
		{
			name:      "10.0.0.0/8 - middle",
			url:       "http://10.128.50.100",
			shouldErr: true,
			reason:    "Should block 10.0.0.0/8 private network",
		},
		{
			name:      "10.0.0.0/8 - end",
			url:       "http://10.255.255.254",
			shouldErr: true,
			reason:    "Should block 10.0.0.0/8 private network",
		},
		{
			name:      "172.16.0.0/12 - start",
			url:       "http://172.16.0.1",
			shouldErr: true,
			reason:    "Should block 172.16.0.0/12 private network",
		},
		{
			name:      "172.16.0.0/12 - middle",
			url:       "http://172.20.10.5",
			shouldErr: true,
			reason:    "Should block 172.16.0.0/12 private network",
		},
		{
			name:      "172.16.0.0/12 - end",
			url:       "http://172.31.255.254",
			shouldErr: true,
			reason:    "Should block 172.16.0.0/12 private network",
		},
		{
			name:      "192.168.0.0/16 - start",
			url:       "http://192.168.0.1",
			shouldErr: true,
			reason:    "Should block 192.168.0.0/16 private network",
		},
		{
			name:      "192.168.0.0/16 - middle",
			url:       "http://192.168.100.50",
			shouldErr: true,
			reason:    "Should block 192.168.0.0/16 private network",
		},
		{
			name:      "192.168.0.0/16 - end",
			url:       "http://192.168.255.254",
			shouldErr: true,
			reason:    "Should block 192.168.0.0/16 private network",
		},

		// Loopback addresses
		{
			name:      "Localhost 127.0.0.1",
			url:       "http://127.0.0.1",
			shouldErr: true,
			reason:    "Should block localhost IP",
		},
		{
			name:      "Localhost 127.0.0.1 with port",
			url:       "http://127.0.0.1:8080",
			shouldErr: true,
			reason:    "Should block localhost IP with port",
		},
		{
			name:      "Localhost 127.0.0.1 with path",
			url:       "http://127.0.0.1/admin",
			shouldErr: true,
			reason:    "Should block localhost IP with path",
		},
		{
			name:      "Loopback range 127.255.255.254",
			url:       "http://127.255.255.254",
			shouldErr: true,
			reason:    "Should block entire 127.0.0.0/8 range",
		},

		// Link-local addresses
		{
			name:      "Link-local 169.254.0.1",
			url:       "http://169.254.0.1",
			shouldErr: true,
			reason:    "Should block link-local addresses",
		},
		{
			name:      "AWS metadata endpoint",
			url:       "http://169.254.169.254",
			shouldErr: true,
			reason:    "Should block AWS metadata endpoint",
		},
		{
			name:      "AWS metadata endpoint with path",
			url:       "http://169.254.169.254/latest/meta-data/",
			shouldErr: true,
			reason:    "Should block AWS metadata endpoint with path",
		},

		// Multicast addresses
		{
			name:      "Multicast start 224.0.0.1",
			url:       "http://224.0.0.1",
			shouldErr: true,
			reason:    "Should block multicast addresses",
		},
		{
			name:      "Multicast 239.255.255.250",
			url:       "http://239.255.255.250",
			shouldErr: true,
			reason:    "Should block multicast addresses",
		},

		// Reserved/Special addresses
		{
			name:      "0.0.0.0",
			url:       "http://0.0.0.0",
			shouldErr: true,
			reason:    "Should block 0.0.0.0",
		},
		{
			name:      "Broadcast 255.255.255.255",
			url:       "http://255.255.255.255",
			shouldErr: true,
			reason:    "Should block broadcast address",
		},
		{
			name:      "Shared address space 100.64.0.1",
			url:       "http://100.64.0.1",
			shouldErr: true,
			reason:    "Should block RFC 6598 shared address space",
		},

		// Test networks
		{
			name:      "TEST-NET-1 192.0.2.1",
			url:       "http://192.0.2.1",
			shouldErr: true,
			reason:    "Should block TEST-NET-1",
		},
		{
			name:      "TEST-NET-2 198.51.100.1",
			url:       "http://198.51.100.1",
			shouldErr: true,
			reason:    "Should block TEST-NET-2",
		},
		{
			name:      "TEST-NET-3 203.0.113.1",
			url:       "http://203.0.113.1",
			shouldErr: true,
			reason:    "Should block TEST-NET-3",
		},

		// IPv6 addresses
		{
			name:      "IPv6 loopback",
			url:       "http://[::1]",
			shouldErr: true,
			reason:    "Should block IPv6 loopback",
		},
		{
			name:      "IPv6 link-local fe80::1",
			url:       "http://[fe80::1]",
			shouldErr: true,
			reason:    "Should block IPv6 link-local",
		},
		{
			name:      "IPv6 unique local fc00::1",
			url:       "http://[fc00::1]",
			shouldErr: true,
			reason:    "Should block IPv6 unique local",
		},
		{
			name:      "IPv6 unique local fd00::1",
			url:       "http://[fd00::1]",
			shouldErr: true,
			reason:    "Should block IPv6 unique local",
		},

		// Invalid schemes
		{
			name:      "JavaScript scheme",
			url:       "javascript:alert('xss')",
			shouldErr: true,
			reason:    "Should block javascript: scheme",
		},
		{
			name:      "File scheme",
			url:       "file:///etc/passwd",
			shouldErr: true,
			reason:    "Should block file: scheme",
		},
		{
			name:      "Data scheme",
			url:       "data:text/html,<script>alert('xss')</script>",
			shouldErr: true,
			reason:    "Should block data: scheme",
		},
		{
			name:      "FTP scheme",
			url:       "ftp://example.com",
			shouldErr: true,
			reason:    "Should block FTP scheme",
		},

		// Edge cases
		{
			name:      "URL with port",
			url:       "https://example.com:8443",
			shouldErr: false,
			reason:    "Should allow public URL with port",
		},
		{
			name:      "URL with path",
			url:       "https://example.com/path/to/resource",
			shouldErr: false,
			reason:    "Should allow public URL with path",
		},
		{
			name:      "URL with query string",
			url:       "https://example.com?param=value",
			shouldErr: false,
			reason:    "Should allow public URL with query",
		},
		{
			name:      "Empty URL",
			url:       "",
			shouldErr: true,
			reason:    "Should reject empty URL",
		},
		{
			name:      "Invalid URL format",
			url:       "not-a-url",
			shouldErr: true,
			reason:    "Should reject invalid URL format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkReq := link.LinkRequest{
				URL: tt.url,
			}

			err := shared.Validate.Struct(&linkReq)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for %s: %s", tt.url, tt.reason)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for %s: %v (%s)", tt.url, err, tt.reason)
			}
		})
	}
}

func TestBoundaryIPAddresses(t *testing.T) {
	shared.InitValidator()

	tests := []struct {
		name      string
		url       string
		shouldErr bool
		reason    string
	}{
		// Just outside private ranges (should be allowed)
		{
			name:      "9.255.255.255 (just before 10.0.0.0/8)",
			url:       "http://9.255.255.255",
			shouldErr: false,
			reason:    "Should allow IPs just outside private range",
		},
		{
			name:      "11.0.0.0 (just after 10.0.0.0/8)",
			url:       "http://11.0.0.0",
			shouldErr: false,
			reason:    "Should allow IPs just outside private range",
		},
		{
			name:      "172.15.255.255 (just before 172.16.0.0/12)",
			url:       "http://172.15.255.255",
			shouldErr: false,
			reason:    "Should allow IPs just outside private range",
		},
		{
			name:      "172.32.0.0 (just after 172.16.0.0/12)",
			url:       "http://172.32.0.0",
			shouldErr: false,
			reason:    "Should allow IPs just outside private range",
		},
		{
			name:      "192.167.255.255 (just before 192.168.0.0/16)",
			url:       "http://192.167.255.255",
			shouldErr: false,
			reason:    "Should allow IPs just outside private range",
		},
		{
			name:      "192.169.0.0 (just after 192.168.0.0/16)",
			url:       "http://192.169.0.0",
			shouldErr: false,
			reason:    "Should allow IPs just outside private range",
		},

		// Edge of private ranges (should be blocked)
		{
			name:      "10.0.0.0 (start of 10.0.0.0/8)",
			url:       "http://10.0.0.0",
			shouldErr: true,
			reason:    "Should block start of private range",
		},
		{
			name:      "10.255.255.255 (end of 10.0.0.0/8)",
			url:       "http://10.255.255.255",
			shouldErr: true,
			reason:    "Should block end of private range",
		},
		{
			name:      "172.16.0.0 (start of 172.16.0.0/12)",
			url:       "http://172.16.0.0",
			shouldErr: true,
			reason:    "Should block start of private range",
		},
		{
			name:      "172.31.255.255 (end of 172.16.0.0/12)",
			url:       "http://172.31.255.255",
			shouldErr: true,
			reason:    "Should block end of private range",
		},
		{
			name:      "192.168.0.0 (start of 192.168.0.0/16)",
			url:       "http://192.168.0.0",
			shouldErr: true,
			reason:    "Should block start of private range",
		},
		{
			name:      "192.168.255.255 (end of 192.168.0.0/16)",
			url:       "http://192.168.255.255",
			shouldErr: true,
			reason:    "Should block end of private range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkReq := link.LinkRequest{
				URL: tt.url,
			}

			err := shared.Validate.Struct(&linkReq)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for %s: %s", tt.url, tt.reason)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for %s: %v (%s)", tt.url, err, tt.reason)
			}
		})
	}
}

func TestURLValidationEdgeCases(t *testing.T) {
	shared.InitValidator()

	tests := []struct {
		name      string
		url       string
		shouldErr bool
		reason    string
	}{
		{
			name:      "URL without scheme",
			url:       "example.com",
			shouldErr: true,
			reason:    "Should reject URL without scheme",
		},
		{
			name:      "URL with username",
			url:       "https://user@example.com",
			shouldErr: false,
			reason:    "Should allow URL with username",
		},
		{
			name:      "URL with username and password",
			url:       "https://user:pass@example.com",
			shouldErr: false,
			reason:    "Should allow URL with credentials",
		},
		{
			name:      "URL with fragment",
			url:       "https://example.com#section",
			shouldErr: false,
			reason:    "Should allow URL with fragment",
		},
		{
			name:      "Private IP with credentials",
			url:       "http://user:pass@192.168.1.1",
			shouldErr: true,
			reason:    "Should block private IP even with credentials",
		},
		{
			name:      "Private IP with path and query",
			url:       "http://10.0.0.1:8080/admin?debug=true",
			shouldErr: true,
			reason:    "Should block private IP with complex URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linkReq := link.LinkRequest{
				URL: tt.url,
			}

			err := shared.Validate.Struct(&linkReq)

			if tt.shouldErr && err == nil {
				t.Errorf("Expected error for %s: %s", tt.url, tt.reason)
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Unexpected error for %s: %v (%s)", tt.url, err, tt.reason)
			}
		})
	}
}
