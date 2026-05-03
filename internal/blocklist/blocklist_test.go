package blocklist

import (
	"net"
	"testing"
)

func TestParseLine(t *testing.T) {
	t.Parallel()
	tests := []struct {
		line string
		ok   bool
	}{
		{"", false},
		{"# comment", false},
		{"; comment", false},
		{"192.0.2.0/24", true},
		{"192.0.2.7", true},
		{"2001:db8::/32", true},
		{"not-an-ip", false},
	}
	for _, tt := range tests {
		_, ok := parseLine(tt.line)
		if ok != tt.ok {
			t.Errorf("parseLine(%q) ok=%v want %v", tt.line, ok, tt.ok)
		}
	}
	ipnet, ok := parseLine("192.0.2.1")
	if !ok || ipnet == nil {
		t.Fatal("expected /32")
	}
	if ones, bits := ipnet.Mask.Size(); ones != 32 || bits != 32 {
		t.Fatalf("mask %v/%v", ones, bits)
	}
}

func TestLookupDisabled(t *testing.T) {
	t.Parallel()
	oldE, oldSnap := enabled, snapshot
	t.Cleanup(func() {
		enabled = oldE
		snapshot = oldSnap
	})
	enabled = false
	snapshot = nil
	if Lookup(net.ParseIP("192.0.2.1"), nil) != nil {
		t.Fatal("expected nil when disabled")
	}
}
