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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ultimaterex/get-ip/internal/cleantoken"
)

const (
	defaultPerQuery        = 3 * time.Second
	defaultDeadline        = 25 * time.Second
	defaultConcurrent      = 12
	defaultCacheTTL        = 15 * time.Minute
	defaultClientMax       = 30  // fresh lookups per client key per window
	defaultClientWindow    = time.Hour
	defaultGlobalPerMinute = 120
	defaultMaxClientRLKeys = 20000
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
	Eligible      bool    `json:"eligible"`
	SkippedReason   string  `json:"skipped_reason,omitempty"`
	IPv4            string  `json:"ipv4,omitempty"`
	ZonesChecked    int     `json:"zones_checked"`
	Listed          bool    `json:"listed"`
	Checks          []Check `json:"checks,omitempty"`
	Cached          bool    `json:"cached,omitempty"`
	CacheExpiresRFC string  `json:"cache_expires,omitempty"`
	RateLimited     bool    `json:"rate_limited,omitempty"`
}

var (
	mu      sync.RWMutex
	enabled bool
	specs   []zoneSpec
	perQuery time.Duration
	deadline time.Duration
	workers  int

	cacheTTL        time.Duration
	dnsblCache      sync.Map // subject IPv4 string -> cacheEntry
	clientLimiter   *clientWindows
	globalLimiter   minuteCounter
	maxClientRLKeys int
)

type cacheEntry struct {
	info    *Info
	expires time.Time
}

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
	workers = intEnvWorkers("DNSBL_CONCURRENCY", defaultConcurrent)
	if workers < 1 {
		workers = 1
	}

	cacheTTL = durationEnvAllowZero("DNSBL_CACHE_TTL", defaultCacheTTL)
	clientMax := intEnvGE0("DNSBL_CLIENT_MAX", defaultClientMax)
	clientWin := durationEnv("DNSBL_CLIENT_WINDOW", defaultClientWindow)
	globalPerMin := intEnvGE0("DNSBL_GLOBAL_MAX_PER_MINUTE", defaultGlobalPerMinute)
	maxClientRLKeys = intEnvGE0("DNSBL_RL_MAX_CLIENT_KEYS", defaultMaxClientRLKeys)
	if maxClientRLKeys == 0 {
		maxClientRLKeys = defaultMaxClientRLKeys
	}

	clientLimiter = &clientWindows{
		window: clientWin,
		max:    clientMax,
	}
	globalLimiter.perMinute = globalPerMin

	enabled = len(specs) > 0
	if enabled {
		log.Printf("dnsbl: enabled (%d zones, per_query=%v deadline=%v concurrency=%d cache_ttl=%v client_max=%d/%v global_per_min=%d)",
			len(specs), perQuery, deadline, workers, cacheTTL, clientMax, clientWin, globalPerMin)
	}
}

func durationEnvAllowZero(key string, def time.Duration) time.Duration {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		log.Printf("dnsbl: invalid %s=%q, using %v", key, s, def)
		return def
	}
	return d
}

func durationEnv(key string, def time.Duration) time.Duration {
	d := durationEnvAllowZero(key, def)
	if d < 100*time.Millisecond {
		log.Printf("dnsbl: %s too small, using %v", key, def)
		return def
	}
	return d
}

func intEnvWorkers(key string, def int) int {
	return intEnvRange(key, def, 1, 256)
}

func intEnvGE0(key string, def int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 0 {
		log.Printf("dnsbl: invalid %s=%q, using %d", key, s, def)
		return def
	}
	return n
}

func intEnvRange(key string, def, min, max int) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < min || n > max {
		log.Printf("dnsbl: invalid %s=%q, using %d", key, s, def)
		return def
	}
	return n
}

// Lookup tests subject IPv4 against configured zones. clientKey identifies the caller for
// per-client rate limits (use the visitor’s public IP key from the HTTP request).
// Returns nil if DNSBL is not configured.
func Lookup(ctx context.Context, clientKey string, v4 net.IP) *Info {
	mu.RLock()
	defer mu.RUnlock()
	if !enabled {
		return nil
	}
	zn := len(specs)
	if clientKey == "" {
		clientKey = "unknown"
	}

	if v4 == nil || v4.To4() == nil {
		return &Info{
			Eligible:      false,
			SkippedReason: "no_public_ipv4",
			ZonesChecked:  zn,
			Listed:        false,
			Checks:        nil,
		}
	}
	ip4 := v4.To4()
	reversed := reverseIPv4(ip4)
	if reversed == "" {
		return &Info{Eligible: false, SkippedReason: "invalid_ipv4", ZonesChecked: zn}
	}
	subject := ip4.String()

	if cacheTTL > 0 {
		if inf, exp, ok := loadCache(subject); ok {
			out := *inf
			out.Cached = true
			if !exp.IsZero() {
				out.CacheExpiresRFC = exp.UTC().Format(time.RFC3339)
			}
			return &out
		}
	}

	now := time.Now()
	if !globalLimiter.allow(now) {
		log.Printf("dnsbl: global rate limit exceeded (DNSBL_GLOBAL_MAX_PER_MINUTE=%d)", globalLimiter.perMinute)
		return rateLimitedOut(zn)
	}
	if clientLimiter != nil && !clientLimiter.allow(clientKey, now) {
		log.Printf("dnsbl: client rate limit exceeded (key=%s DNSBL_CLIENT_MAX=%d per %v)", clientKey, clientLimiter.max, clientLimiter.window)
		return rateLimitedOut(zn)
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
	out := &Info{
		Eligible:     true,
		IPv4:         subject,
		ZonesChecked: zn,
		Listed:       listed,
		Checks:       res,
	}
	if cacheTTL > 0 {
		storeCache(subject, out, now.Add(cacheTTL))
	}
	return out
}

func rateLimitedOut(zn int) *Info {
	return &Info{
		Eligible:      false,
		SkippedReason: "rate_limited",
		ZonesChecked:  zn,
		RateLimited:   true,
	}
}

func loadCache(subject string) (*Info, time.Time, bool) {
	v, ok := dnsblCache.Load(subject)
	if !ok {
		return nil, time.Time{}, false
	}
	e := v.(cacheEntry)
	if time.Now().After(e.expires) {
		dnsblCache.Delete(subject)
		return nil, time.Time{}, false
	}
	return e.info, e.expires, true
}

func storeCache(subject string, inf *Info, exp time.Time) {
	cp := *inf
	cp.Cached = false
	cp.CacheExpiresRFC = ""
	cp.RateLimited = false
	dnsblCache.Store(subject, cacheEntry{info: &cp, expires: exp})
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
