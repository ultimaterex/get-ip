// Package blocklist loads optional IP deny/block lists (CIDR text feeds) and answers membership queries.
// Data sources and licenses are operator responsibilities — see project README.
package blocklist

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ultimaterex/get-ip/internal/cleantoken"
	"github.com/yl2chen/cidranger"
)

const (
	maxDownloadBytes = 64 << 20 // 64 MiB per URL
	httpTimeout      = 2 * time.Minute
	defaultRefresh   = 24 * time.Hour
	userAgent        = "get-ip/blocklist (+https://github.com/ultimaterex/get-ip)"
)

// MatchDetail is one longest-prefix hit from a configured feed for a given client IP.
type MatchDetail struct {
	Source string // feed tag (from URL or |tag)
	IP     string // client IP that matched
	Prefix string // containing CIDR from the feed (most specific when overlaps exist)
	Family string // "ipv4" or "ipv6"
}

// Info is a snapshot suitable for JSON after checking visitor IPs against configured lists.
type Info struct {
	Listed        bool
	Matched       []string      // unique source tags with any hit
	Details       []MatchDetail // per IP × source when listed (longest-prefix row)
	SourcesLoaded int
	LastRefresh   *time.Time
}

var (
	mu       sync.RWMutex
	snapshot *snapshotData
	enabled  bool
	specs    []sourceSpec
	refresh  time.Duration
)

type snapshotData struct {
	sources []*loadedSource
	at      time.Time
}

type loadedSource struct {
	tag string
	r   cidranger.Ranger
}

type sourceSpec struct {
	rawURL string
	tag    string
}

// InitFromEnv starts optional background refresh. If BLOCKLIST_URLS is empty, lookups are disabled.
func InitFromEnv(ctx context.Context) {
	specs = parseSpecs(strings.TrimSpace(os.Getenv("BLOCKLIST_URLS")))
	if len(specs) == 0 {
		return
	}
	enabled = true
	refresh = refreshInterval()

	firstCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	if err := refreshOnce(firstCtx); err != nil {
		log.Printf("blocklist: initial refresh: %v", err)
	}
	cancel()

	go loop(ctx)
}

func refreshInterval() time.Duration {
	s := strings.TrimSpace(os.Getenv("BLOCKLIST_REFRESH"))
	if s == "" {
		return defaultRefresh
	}
	d, err := time.ParseDuration(s)
	if err != nil || d < time.Minute {
		log.Printf("blocklist: invalid BLOCKLIST_REFRESH %q, using %v", s, defaultRefresh)
		return defaultRefresh
	}
	return d
}

func loop(ctx context.Context) {
	t := time.NewTicker(refresh)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c, cancel := context.WithTimeout(context.Background(), httpTimeout)
			if err := refreshOnce(c); err != nil {
				log.Printf("blocklist: refresh: %v", err)
			}
			cancel()
		}
	}
}

func refreshOnce(ctx context.Context) error {
	if len(specs) == 0 {
		return nil
	}
	client := &http.Client{Timeout: httpTimeout}

	var loaded []*loadedSource
	for _, sp := range specs {
		r := cidranger.NewPCTrieRanger()
		n, err := fetchAndInsert(ctx, client, sp.rawURL, r)
		if err != nil {
			return fmt.Errorf("%s: %w", sp.tag, err)
		}
		if n == 0 {
			log.Printf("blocklist: %s: no CIDR rows parsed from %s", sp.tag, sp.rawURL)
		} else {
			log.Printf("blocklist: %s: loaded %d prefixes from %s", sp.tag, n, sp.rawURL)
		}
		loaded = append(loaded, &loadedSource{tag: sp.tag, r: r})
	}

	mu.Lock()
	snapshot = &snapshotData{sources: loaded, at: time.Now().UTC()}
	mu.Unlock()
	return nil
}

func fetchAndInsert(ctx context.Context, client *http.Client, rawURL string, r cidranger.Ranger) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return 0, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	lr := io.LimitReader(resp.Body, maxDownloadBytes)
	sc := bufio.NewScanner(lr)
	const maxLine = 512 << 10
	buf := make([]byte, 0, 64*1024)
	sc.Buffer(buf, maxLine)

	n := 0
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		ipnet, ok := parseLine(line)
		if !ok {
			continue
		}
		if err := r.Insert(cidranger.NewBasicRangerEntry(*ipnet)); err != nil {
			return n, err
		}
		n++
	}
	if err := sc.Err(); err != nil && !errors.Is(err, io.EOF) {
		return n, err
	}
	return n, nil
}

func parseLine(line string) (*net.IPNet, bool) {
	if line == "" {
		return nil, false
	}
	if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		return nil, false
	}
	// Spamhaus-style metadata lines
	if strings.HasPrefix(strings.ToLower(line), "your match") {
		return nil, false
	}

	fields := strings.Fields(line)
	if len(fields) == 0 {
		return nil, false
	}
	content := fields[0]

	if strings.Contains(content, "/") {
		_, ipnet, err := net.ParseCIDR(content)
		if err != nil {
			return nil, false
		}
		return ipnet, true
	}
	ip := net.ParseIP(content)
	if ip == nil {
		return nil, false
	}
	if v4 := ip.To4(); v4 != nil {
		return &net.IPNet{IP: v4, Mask: net.CIDRMask(32, 32)}, true
	}
	return &net.IPNet{IP: ip.To16(), Mask: net.CIDRMask(128, 128)}, true
}

// Lookup returns membership info for public visitor IPs (typically v4 and/or v6).
func Lookup(v4, v6 net.IP) *Info {
	if !enabled {
		return nil
	}
	mu.RLock()
	snap := snapshot
	mu.RUnlock()
	if snap == nil || len(snap.sources) == 0 {
		return &Info{
			Listed:        false,
			Matched:       nil,
			Details:       nil,
			SourcesLoaded: 0,
			LastRefresh:   nil,
		}
	}

	tagSeen := make(map[string]struct{})
	var details []MatchDetail

	for _, ip := range []net.IP{v4, v6} {
		if ip == nil {
			continue
		}
		ipStr := ip.String()
		fam := "ipv6"
		if ip.To4() != nil {
			fam = "ipv4"
		}
		for _, src := range snap.sources {
			entries, err := src.r.ContainingNetworks(ip)
			if err != nil || len(entries) == 0 {
				continue
			}
			prefix := longestPrefixCIDR(entries)
			if prefix == "" {
				continue
			}
			details = append(details, MatchDetail{
				Source: src.tag,
				IP:     ipStr,
				Prefix: prefix,
				Family: fam,
			})
			tagSeen[src.tag] = struct{}{}
		}
	}

	matched := make([]string, 0, len(tagSeen))
	for t := range tagSeen {
		matched = append(matched, t)
	}
	sort.Strings(matched)

	sort.Slice(details, func(i, j int) bool {
		if details[i].Source != details[j].Source {
			return details[i].Source < details[j].Source
		}
		if details[i].IP != details[j].IP {
			return details[i].IP < details[j].IP
		}
		return details[i].Prefix < details[j].Prefix
	})

	at := snap.at
	return &Info{
		Listed:        len(details) > 0,
		Matched:       matched,
		Details:       details,
		SourcesLoaded: len(snap.sources),
		LastRefresh:   &at,
	}
}

func longestPrefixCIDR(entries []cidranger.RangerEntry) string {
	var best *net.IPNet
	bestOnes := -1
	for _, e := range entries {
		n := e.Network()
		ones, bits := n.Mask.Size()
		if bits == 0 {
			continue
		}
		if ones > bestOnes {
			bestOnes = ones
			nn := n
			best = &nn
		}
	}
	if best == nil {
		return ""
	}
	return best.String()
}

func parseSpecs(block string) []sourceSpec {
	var out []sourceSpec
	for _, part := range strings.Split(block, ";") {
		part = cleantoken.SemicolonPart(part)
		if part == "" {
			continue
		}
		rawURL, tag := part, ""
		if bits := strings.SplitN(part, "|", 2); len(bits) == 2 {
			rawURL = cleantoken.SemicolonPart(bits[0])
			tag = cleantoken.SemicolonPart(bits[1])
		}
		if rawURL == "" {
			continue
		}
		if _, err := url.ParseRequestURI(rawURL); err != nil {
			log.Printf("blocklist: skip invalid URL %q: %v", rawURL, err)
			continue
		}
		if tag == "" {
			u, err := url.Parse(rawURL)
			if err != nil {
				continue
			}
			tag = tagFromURL(u)
		}
		out = append(out, sourceSpec{rawURL: rawURL, tag: tag})
	}
	return out
}

func tagFromURL(u *url.URL) string {
	base := path.Base(u.Path)
	if base == "" || base == "." || base == "/" {
		host := strings.ReplaceAll(u.Hostname(), ".", "-")
		return sanitizeTag(host)
	}
	base = strings.TrimSuffix(base, ".txt")
	base = strings.TrimSuffix(base, ".netset")
	base = strings.TrimSuffix(base, ".ipset")
	return sanitizeTag(base)
}

func sanitizeTag(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	t := strings.Trim(b.String(), "-")
	if t == "" {
		return "list"
	}
	return t
}
