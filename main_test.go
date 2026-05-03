package main

import (
	"net"
	"testing"
)

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
