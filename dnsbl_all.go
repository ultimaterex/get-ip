package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/ultimaterex/get-ip/internal/dnsbl"
)

func writeDNSBLSection(b *strings.Builder, r *http.Request, v4 net.IP) {
	ctx := context.Background()
	if r != nil {
		ctx = r.Context()
	}
	inf := dnsbl.Lookup(ctx, v4)
	if inf == nil {
		return
	}
	fmt.Fprintf(b, "DNS blocklists (DNSBL)\n")
	fmt.Fprintf(b, "  Zones configured: %d\n", inf.ZonesChecked)
	if !inf.Eligible {
		fmt.Fprintf(b, "  Skipped: %s\n", dnsblSkippedHuman(inf.SkippedReason))
		return
	}
	fmt.Fprintf(b, "  IPv4 checked: %s\n", inf.IPv4)
	if inf.Listed {
		fmt.Fprintf(b, "  Listed on at least one zone: yes\n")
	} else {
		fmt.Fprintf(b, "  Listed on at least one zone: no\n")
	}
	for _, c := range inf.Checks {
		line := fmt.Sprintf("    - %s (%s): ", c.Source, c.Zone)
		switch {
		case c.Error != "":
			fmt.Fprintf(b, "%serror — %s (%d ms)\n", line, c.Error, c.ResponseMs)
		case c.Listed:
			fmt.Fprintf(b, "%sLISTED %v (%d ms)\n", line, c.ReturnCodes, c.ResponseMs)
		default:
			fmt.Fprintf(b, "%sok (%d ms)\n", line, c.ResponseMs)
		}
	}
}

func dnsblSkippedHuman(reason string) string {
	switch reason {
	case "no_public_ipv4":
		return "no public IPv4 (most DNSBLs are IPv4-only)"
	case "invalid_ipv4":
		return "invalid IPv4"
	default:
		return reason
	}
}
