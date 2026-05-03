package dnsbl

import (
	"net"
	"testing"
)

func TestReverseIPv4(t *testing.T) {
	t.Parallel()
	ip := net.ParseIP("190.98.108.105")
	got := reverseIPv4(ip)
	want := "105.108.98.190"
	if got != want {
		t.Fatalf("reverseIPv4: got %q want %q", got, want)
	}
	if reverseIPv4(nil) != "" {
		t.Fatal("nil ip")
	}
}

func TestInterpretReturnAddrs(t *testing.T) {
	t.Parallel()
	ok, codes := interpretReturnAddrs([]string{"127.0.0.9", "127.0.0.4"})
	if !ok || len(codes) != 2 {
		t.Fatalf("listed=%v codes=%v", ok, codes)
	}
	ok2, _ := interpretReturnAddrs([]string{"192.0.2.1"})
	if ok2 {
		t.Fatal("non-127 should not list")
	}
}

func TestParseZones(t *testing.T) {
	t.Parallel()
	s := " zen.spamhaus.org|zen ; b.barracudacentral.org "
	got := parseZones(s)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].zone != "zen.spamhaus.org" || got[0].tag != "zen" {
		t.Errorf("first %+v", got[0])
	}
	if got[1].tag != "b-barracudacentral-org" {
		t.Errorf("tag %+v", got[1])
	}
}

func TestItoaOctet(t *testing.T) {
	t.Parallel()
	if itoa(0) != "0" || itoa(255) != "255" || itoa(16) != "16" {
		t.Fatalf("itoa broken: %q %q %q", itoa(0), itoa(255), itoa(16))
	}
}
