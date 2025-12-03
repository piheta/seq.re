package tests

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/piheta/seq.re/config"
	"github.com/piheta/seq.re/internal/features/ip"
)

func TestGetClientIPFromXForwardedFor(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv() // Reload config with new env var

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1, 172.16.0.1")

	result := service.GetClientIP(req)

	if result.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", result.IP)
	}
}

func TestGetClientIPFromXForwardedForSingleIP(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.45")

	result := service.GetClientIP(req)

	if result.IP != "203.0.113.45" {
		t.Errorf("expected IP 203.0.113.45, got %s", result.IP)
	}
}

func TestGetClientIPFromXForwardedForWithWhitespace(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "  192.168.1.1  , 10.0.0.1")

	result := service.GetClientIP(req)

	if result.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1 (trimmed), got %s", result.IP)
	}
}

func TestGetClientIPFromXRealIP(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Real-IP", "198.51.100.89")

	result := service.GetClientIP(req)

	if result.IP != "198.51.100.89" {
		t.Errorf("expected IP 198.51.100.89, got %s", result.IP)
	}
}

func TestGetClientIPXForwardedForTakesPreference(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	req.Header.Set("X-Real-IP", "198.51.100.89")

	result := service.GetClientIP(req)

	if result.IP != "192.168.1.1" {
		t.Errorf("expected X-Forwarded-For to take preference, got %s", result.IP)
	}
}

func TestGetClientIPFromRemoteAddr(t *testing.T) {
	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "203.0.113.100:54321"

	result := service.GetClientIP(req)

	if result.IP != "203.0.113.100" {
		t.Errorf("expected IP 203.0.113.100, got %s", result.IP)
	}
}

func TestGetClientIPRemoteAddrWithoutPort(t *testing.T) {
	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "203.0.113.100"

	result := service.GetClientIP(req)

	// net.SplitHostPort returns empty string for host when there's no port
	if result.IP != "" {
		t.Errorf("expected empty IP when RemoteAddr has no port, got %s", result.IP)
	}
}

func TestGetClientIPIPv6FromRemoteAddr(t *testing.T) {
	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "[2001:db8::1]:54321"

	result := service.GetClientIP(req)

	if result.IP != "2001:db8::1" {
		t.Errorf("expected IPv6 2001:db8::1, got %s", result.IP)
	}
}

func TestGetClientIPIPv6FromXForwardedFor(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "2001:db8::1, 192.168.1.1")

	result := service.GetClientIP(req)

	if result.IP != "2001:db8::1" {
		t.Errorf("expected IPv6 2001:db8::1, got %s", result.IP)
	}
}

func TestGetClientIPEmptyHeaders(t *testing.T) {
	_ = os.Unsetenv("BEHIND_PROXY")
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	result := service.GetClientIP(req)

	if result.IP != "127.0.0.1" {
		t.Errorf("expected IP 127.0.0.1 from RemoteAddr, got %s", result.IP)
	}
}

func TestGetClientIPIgnoresHeadersWhenNotBehindProxy(t *testing.T) {
	// Make sure BEHIND_PROXY is not set
	_ = os.Unsetenv("BEHIND_PROXY")
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.RemoteAddr = "203.0.113.5:54321"
	req.Header.Set("X-Forwarded-For", "192.168.1.1") // Should be ignored
	req.Header.Set("X-Real-IP", "10.0.0.1")          // Should be ignored

	result := service.GetClientIP(req)

	// Should use RemoteAddr, not spoofed headers
	if result.IP != "203.0.113.5" {
		t.Errorf("expected IP from RemoteAddr (203.0.113.5), got %s (headers should be ignored)", result.IP)
	}
}

func TestIPStructure(t *testing.T) {
	addr := ip.IP{IP: "192.168.1.1"}

	if addr.IP != "192.168.1.1" {
		t.Errorf("expected IP 192.168.1.1, got %s", addr.IP)
	}
}

func TestIPServiceCreation(t *testing.T) {
	service := ip.NewIPService()

	if service == nil {
		t.Error("expected NewIPService to return non-nil service")
	}
}

func TestGetClientIPConsistency(t *testing.T) {
	_ = os.Setenv("BEHIND_PROXY", "true")
	defer func() { _ = os.Unsetenv("BEHIND_PROXY") }()
	config.InitEnv()

	service := ip.NewIPService()

	req := httptest.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	result1 := service.GetClientIP(req)
	result2 := service.GetClientIP(req)

	if result1.IP != result2.IP {
		t.Errorf("expected consistent results, got %s and %s", result1.IP, result2.IP)
	}
}
