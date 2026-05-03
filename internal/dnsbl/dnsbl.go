// Package dnsbl performs optional DNS-based blocklist (DNSBL / RBL) lookups for IPv4.
// Each publisher sets acceptable-use rules — operators must comply.
package dnsbl

import (
	"context"
	"errors"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ultimaterex/get-ip/internal/cleantoken"
)

const (
	defaultPerQuery   = 3 * time.Second
	defaultDeadline   = 25 * time.Second
	defaultConcurrent = 12
)

// Check is one zone result against the visitor IPv4 address.
type Check struct {
	Source      string   `json:"source"`
	Zone        string   `json:"zone"`
	Listed      bool     `json:"listed"`
	ReturnCodes []string `json:"return_codes,omitempty"`
	Error       string   `json:"error,omitempty"`
	ResponseMs  int64    `json:"response_ms"`
}

// Info is returned when DNSBL is configured (even if the client has no IPv4 to check).
type Info struct {
	Eligible      bool   `json:"eligible"`
	SkippedReason string `json:"skipped_reason,omitempty"`
	IPv4          string `json:"ipv4,omitempty"`
	ZonesChecked  int    `json:"zones_checked"`
	Listed        bool   `json:"listed"`
	Checks        []Check `json:"checks,omitempty"`
}

var (
	mu       sync.RWMutex
	enabled  bool
	specs    []zoneSpec
	perQuery time.Duration
	deadline time.Duration
	workers  int
)

type zoneSpec struct {
	zone string
	tag  string
}

// InitFromEnv enables lookups when DNSBL_ZONES is non-empty.
func InitFromEnv() {
	mu.Lock()
	defer mu.Unlock()
	specs = parseZones(strings.TrimSpace(os.Getenv("DNSBL_ZONES")))
	perQuery = durationEnv("DNSBL_PER_QUERY", defaultPerQuery)
	deadline = durationEnv("DNSBL_DEADLINE", defaultDeadline)
	workers = intEnv("DNSBL_CONCURRENCY", defaultConcurrent)
	if workers < 1 {
		workers = 1
	}
	enabled = len(specs) > 0
	if enabled {
		log.Printf("dnsbl: enabled (%d zones, per_query=%v deadline=%v concurrency=%d)", len(specs), perQuery, deadline, workers)
	}
}

func durationEnv(key string, def time.Duration) time.Duration {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil || d < 100*time.Millisecond {
		log.Printf("dnsbl: invalid %s=%q, using %v", key, s, def)
		return def
	}
	return d
}

func intEnv(key string, def int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	var n int
	for _, r := range s {
		if r < '0' || r > '9' {
			log.Printf("dnsbl: invalid %s=%q, using %d", key, s, def)
			return def
		}
		n = n*10 + int(r-'0')
	}
	if n < 1 || n > 256 {
		log.Printf("dnsbl: %s=%d out of range, using %d", key, n, def)
		return def
	}
	return n
}

// Lookup tests visitor IPv4 against configured zones. Returns nil if DNSBL is not configured.
func Lookup(ctx context.Context, v4 net.IP) *Info {
	mu.RLock()
	defer mu.RUnlock()
	if !enabled {
		return nil
	}
	zn := len(specs)
	if v4 == nil || v4.To4() == nil {
		return &Info{
			Eligible:      false,
			SkippedReason:   "no_public_ipv4",
			ZonesChecked:    zn,
			Listed:        false,
			Checks:        nil,
		}
	}
	ip4 := v4.To4()
	reversed := reverseIPv4(ip4)
	if reversed == "" {
		return &Info{Eligible: false, SkippedReason: "invalid_ipv4", ZonesChecked: zn}
	}

	ctx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()

	var res []Check
	var wg sync.WaitGroup
	ch := make(chan Check, zn)
	sem := make(chan struct{}, workers)

	for _, sp := range specs {
		wg.Add(1)
		go func(sp zoneSpec) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			fqdn := reversed + "." + strings.TrimSuffix(sp.zone, ".")
			zctx, zcancel := context.WithTimeout(ctx, perQuery)
			defer zcancel()

			start := time.Now()
			addrs, err := net.DefaultResolver.LookupHost(zctx, fqdn)
			ms := time.Since(start).Milliseconds()

			chk := Check{
				Source:     sp.tag,
				Zone:       sp.zone,
				ResponseMs: ms,
			}
			if err != nil {
				var dnsErr *net.DNSError
				if errors.As(err, &dnsErr) && dnsErr.IsNotFound {
					chk.Listed = false
				} else {
					chk.Error = trimErr(err)
				}
				ch <- chk
				return
			}
			listed, codes := interpretReturnAddrs(addrs)
			chk.Listed = listed
			chk.ReturnCodes = codes
			ch <- chk
		}(sp)
	}
	wg.Wait()
	close(ch)
	for c := range ch {
		res = append(res, c)
	}
	sort.Slice(res, func(i, j int) bool {
		if res[i].Source != res[j].Source {
			return res[i].Source < res[j].Source
		}
		return res[i].Zone < res[j].Zone
	})
	listed := false
	for _, c := range res {
		if c.Listed {
			listed = true
			break
		}
	}
	return &Info{
		Eligible:     true,
		IPv4:         ip4.String(),
		ZonesChecked: zn,
		Listed:       listed,
		Checks:       res,
	}
}

func trimErr(err error) string {
	s := err.Error()
	if len(s) > 180 {
		return s[:180] + "…"
	}
	return s
}

func interpretReturnAddrs(addrs []string) (listed bool, codes []string) {
	for _, a := range addrs {
		ip := net.ParseIP(a)
		if ip == nil {
			continue
		}
		if v4 := ip.To4(); v4 != nil && v4[0] == 127 {
			codes = append(codes, v4.String())
		}
	}
	return len(codes) > 0, codes
}

func reverseIPv4(ip net.IP) string {
	v4 := ip.To4()
	if v4 == nil {
		return ""
	}
	return strings.Join([]string{
		itoa(int(v4[3])),
		itoa(int(v4[2])),
		itoa(int(v4[1])),
		itoa(int(v4[0])),
	}, ".")
}

func itoa(n int) string {
	// single-octet values only
	if n < 0 || n > 255 {
		return "0"
	}
	if n == 0 {
		return "0"
	}
	var buf [3]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func parseZones(block string) []zoneSpec {
	var out []zoneSpec
	for _, part := range strings.Split(block, ";") {
		part = cleantoken.SemicolonPart(part)
		if part == "" {
			continue
		}
		zone, tag := part, ""
		if bits := strings.SplitN(part, "|", 2); len(bits) == 2 {
			zone = cleantoken.SemicolonPart(bits[0])
			tag = cleantoken.SemicolonPart(bits[1])
		}
		zone = strings.TrimSpace(zone)
		if zone == "" {
			continue
		}
		zone = strings.TrimSuffix(zone, ".")
		if tag == "" {
			tag = tagFromZone(zone)
		}
		out = append(out, zoneSpec{zone: zone, tag: tag})
	}
	return out
}

func tagFromZone(zone string) string {
	t := strings.ReplaceAll(strings.ToLower(zone), ".", "-")
	var b strings.Builder
	for _, r := range t {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	s := strings.Trim(b.String(), "-")
	if s == "" {
		return "zone"
	}
	return s
}
