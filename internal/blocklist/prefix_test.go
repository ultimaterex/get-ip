package blocklist

import (
	"net"
	"testing"

	"github.com/yl2chen/cidranger"
)

func TestLongestPrefixCIDR(t *testing.T) {
	t.Parallel()
	r := cidranger.NewPCTrieRanger()
	mustCIDR := func(s string) net.IPNet {
		t.Helper()
		_, ipnet, err := net.ParseCIDR(s)
		if err != nil {
			t.Fatal(err)
		}
		return *ipnet
	}
	_ = r.Insert(cidranger.NewBasicRangerEntry(mustCIDR("203.0.113.0/24")))
	_ = r.Insert(cidranger.NewBasicRangerEntry(mustCIDR("203.0.113.0/28")))

	ip := net.ParseIP("203.0.113.7")
	entries, err := r.ContainingNetworks(ip)
	if err != nil || len(entries) != 2 {
		t.Fatalf("entries=%d err=%v", len(entries), err)
	}
	if got := longestPrefixCIDR(entries); got != "203.0.113.0/28" {
		t.Fatalf("got %q", got)
	}
}
