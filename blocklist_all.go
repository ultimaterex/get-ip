package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ultimaterex/get-ip/internal/blocklist"
)

func writeBlocklistSection(b *strings.Builder, v4, v6 net.IP) {
	inf := blocklist.Lookup(v4, v6)
	if inf == nil {
		return
	}
	fmt.Fprintf(b, "Blocklists\n")
	fmt.Fprintf(b, "  Sources loaded: %d\n", inf.SourcesLoaded)
	if inf.LastRefresh != nil {
		fmt.Fprintf(b, "  Last refresh: %s\n", inf.LastRefresh.UTC().Format(time.RFC3339))
	}
	if inf.Listed && len(inf.Matched) > 0 {
		fmt.Fprintf(b, "  Listed: yes\n")
		fmt.Fprintf(b, "  Matched: %s\n", strings.Join(inf.Matched, ", "))
		for _, d := range inf.Details {
			fmt.Fprintf(b, "    - %s · %s → %s (%s)\n", d.Source, d.IP, d.Prefix, d.Family)
		}
	} else {
		fmt.Fprintf(b, "  Listed: no\n")
	}
}
