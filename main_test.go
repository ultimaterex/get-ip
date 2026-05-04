package main

import (
	"net"
	"net/http/httptest"
	"testing"
)

func TestCollectIPCandidates_invalidXFFToken(t *testing.T) {
	t.Parallel()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "not-an-ip, 192.0.2.1")
	r.RemoteAddr = "198.51.100.1:4444"
	got := collectIPCandidates(r)
	if len(got) < 2 {
		t.Fatalf("want at least invalid + valid, got %d", len(got))
	}
	if got[0].IP != nil || got[0].Raw != "not-an-ip" {
		t.Fatalf("first hop: %+v", got[0])
	}
	var saw192 bool
	for _, c := range got {
		if c.IP != nil && c.IP.String() == "192.0.2.1" {
			saw192 = true
		}
	}
	if !saw192 {
		t.Fatalf("missing 192.0.2.1 in %+v", got)
	}
}

func TestDebugHTTPEnabled(t *testing.T) {
	t.Setenv("GET_IP_DEBUG_HTTP", "")
	if debugHTTPEnabled() {
		t.Fatal("empty should be off")
	}
	t.Setenv("GET_IP_DEBUG_HTTP", "1")
	if !debugHTTPEnabled() {
		t.Fatal("1 should enable")
	}
}

func TestIsPublicVisitorIP(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"172.22.0.1", false},
		{"10.0.0.1", false},
		{"192.168.1.1", false},
		{"100.64.0.1", false},
		{"127.0.0.1", false},
		{"::1", false},
		{"fe80::1", false},
		{"86.38.216.249", true},
		{"2001:db8::1", true},
	}
	for _, c := range cases {
		ip := net.ParseIP(c.s)
		if got := isPublicVisitorIP(ip); got != c.want {
			t.Errorf("isPublicVisitorIP(%q) = %v, want %v (IsGlobalUnicast=%v IsPrivate=%v)", c.s, got, c.want, ip.IsGlobalUnicast(), ip.IsPrivate())
		}
	}
}
