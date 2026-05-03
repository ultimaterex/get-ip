package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/ultimaterex/get-ip/internal/envload"
	"github.com/ultimaterex/get-ip/internal/geolookup"
)

func main() {
	log.SetOutput(os.Stdout)

	envload.DotEnv()
	cleanupLog := configureLogOutput()
	defer cleanupLog()
	initGeoLite(context.Background())

	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/all", handleAll)
	mux.HandleFunc("/json", handleJSON)

	addr := ":" + port
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// handleRoot returns a single IPv4 address (plain text), or the best available public/client IP as IPv4 when possible.
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ip := preferredIPv4(r)
	if ip == nil {
		ip = preferredIPv6(r)
	}
	if ip == nil {
		log.Printf("%s %s -> no public address", r.Method, r.URL.Path)
		http.Error(w, "could not determine client IP", http.StatusServiceUnavailable)
		return
	}

	log.Printf("%s %s -> %s", r.Method, r.URL.Path, ip.String())

	if prefersHTML(r) {
		writeRootHTML(w, r, ip.String())
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, ip.String())
}

func handleAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	v4, v6 := publicIPv4IPv6(r)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	var b strings.Builder
	fmt.Fprintf(&b, "IPv4: %s\n", formatIP(v4))
	fmt.Fprintf(&b, "IPv6: %s\n", formatIP(v6))
	fmt.Fprintf(&b, "\n")
	writeForwardedSection(&b, r)
	fmt.Fprintf(&b, "\n")
	writeGeoSection(&b, r)
	fmt.Fprintf(&b, "\n")
	writeASNSection(&b, r)
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "Request\n")
	fmt.Fprintf(&b, "  Method: %s\n", r.Method)
	fmt.Fprintf(&b, "  Host: %s\n", r.Host)
	fmt.Fprintf(&b, "  Protocol: %s\n", r.Proto)
	if ua := r.Header.Get("User-Agent"); ua != "" {
		fmt.Fprintf(&b, "  User-Agent: %s\n", ua)
	}

	log.Printf("%s %s -> v4=%s v6=%s", r.Method, r.URL.Path, formatIP(v4), formatIP(v6))

	fmt.Fprint(w, b.String())
}

func handleJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/json" {
		http.NotFound(w, r)
		return
	}

	v4, v6 := publicIPv4IPv6(r)
	resp := jsonResponse{
		Request: jsonRequestMeta{
			Method:    r.Method,
			Host:      r.Host,
			Protocol:  r.Proto,
			UserAgent: r.Header.Get("User-Agent"),
		},
		Forwarded: buildJSONForwarded(r),
		Geo:       lookupVisitorGeo(r),
		ASN:       lookupVisitorASN(r),
	}
	if v4 != nil {
		s := v4.String()
		resp.IPv4 = &s
	}
	if v6 != nil {
		s := v6.String()
		resp.IPv6 = &s
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Printf("json marshal: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	log.Printf("%s %s -> v4=%s v6=%s", r.Method, r.URL.Path, formatIP(v4), formatIP(v6))

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(out)
	w.Write([]byte("\n"))
}

type jsonResponse struct {
	IPv4      *string         `json:"ipv4"`
	IPv6      *string         `json:"ipv6"`
	Forwarded *jsonForwarded  `json:"forwarded,omitempty"`
	Geo       *geolookup.Geo `json:"geo,omitempty"`
	ASN       *geolookup.ASN `json:"asn,omitempty"`
	Request   jsonRequestMeta `json:"request"`
}

type jsonForwarded struct {
	CFConnectingIP *string  `json:"cf_connecting_ip,omitempty"`
	TrueClientIP   *string  `json:"true_client_ip,omitempty"`
	XRealIP        *string  `json:"x_real_ip,omitempty"`
	XForwardedFor  []string `json:"x_forwarded_for,omitempty"`
}

type jsonRequestMeta struct {
	Method    string `json:"method"`
	Host      string `json:"host"`
	Protocol  string `json:"protocol"`
	UserAgent string `json:"user_agent,omitempty"`
}

func publicIPv4IPv6(r *http.Request) (v4, v6 net.IP) {
	for _, ip := range collectCandidates(r) {
		if !isPublicVisitorIP(ip) {
			continue
		}
		if ip4 := ip.To4(); ip4 != nil {
			if v4 == nil {
				v4 = ip4
			}
		} else if ip.To16() != nil && ip.To4() == nil {
			if v6 == nil {
				v6 = ip
			}
		}
	}
	return v4, v6
}

func optionalPublicIPString(r *http.Request, header string) *string {
	v := strings.TrimSpace(r.Header.Get(header))
	if v == "" {
		return nil
	}
	ip := net.ParseIP(v)
	if ip == nil || !isPublicVisitorIP(ip) {
		return nil
	}
	s := ip.String()
	return &s
}

func buildJSONForwarded(r *http.Request) *jsonForwarded {
	f := jsonForwarded{
		CFConnectingIP: optionalPublicIPString(r, "CF-Connecting-IP"),
		TrueClientIP:   optionalPublicIPString(r, "True-Client-IP"),
		XRealIP:        optionalPublicIPString(r, "X-Real-IP"),
	}
	if x := publicXFFList(r); len(x) > 0 {
		f.XForwardedFor = x
	}
	if f.CFConnectingIP == nil && f.TrueClientIP == nil && f.XRealIP == nil && len(f.XForwardedFor) == 0 {
		return nil
	}
	return &f
}

// isPublicVisitorIP reports IPs safe to show visitors: routable on the public Internet.
// net.IP.IsGlobalUnicast is NOT sufficient — for IPv4 it returns true even for RFC1918
// private space (e.g. 172.22.0.1). We exclude private, loopback, link-local, and multicast.
func isPublicVisitorIP(ip net.IP) bool {
	if ip == nil || ip.IsUnspecified() {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsMulticast() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		// RFC 6598 CGNAT — treat as non-public for consistent behavior across Go versions.
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
	}
	return ip.IsGlobalUnicast()
}

// writeForwardedSection prints proxy-supplied headers that contain public addresses only.
// If none are present after filtering, the section is omitted.
func writeForwardedSection(b *strings.Builder, r *http.Request) {
	var inner strings.Builder
	writePublicIPHeader(&inner, r, "CF-Connecting-IP")
	writePublicIPHeader(&inner, r, "True-Client-IP")
	writePublicIPHeader(&inner, r, "X-Real-IP")
	writePublicXForwardedFor(&inner, r)
	if inner.Len() == 0 {
		return
	}
	fmt.Fprintf(b, "Forwarded headers (public addresses only)\n")
	b.WriteString(inner.String())
}

func writePublicIPHeader(b *strings.Builder, r *http.Request, name string) {
	v := strings.TrimSpace(r.Header.Get(name))
	if v == "" {
		return
	}
	ip := net.ParseIP(v)
	if ip == nil || !isPublicVisitorIP(ip) {
		return
	}
	fmt.Fprintf(b, "  %s: %s\n", name, ip.String())
}

func writePublicXForwardedFor(b *strings.Builder, r *http.Request) {
	parts := publicXFFList(r)
	if len(parts) == 0 {
		return
	}
	fmt.Fprintf(b, "  X-Forwarded-For: %s\n", strings.Join(parts, ", "))
}

func publicXFFList(r *http.Request) []string {
	v := r.Header.Get("X-Forwarded-For")
	if v == "" {
		return nil
	}
	var parts []string
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		ip := net.ParseIP(p)
		if ip != nil && isPublicVisitorIP(ip) {
			parts = append(parts, ip.String())
		}
	}
	return parts
}

func formatIP(ip net.IP) string {
	if ip == nil {
		return "(none)"
	}
	return ip.String()
}

func preferredIPv4(r *http.Request) net.IP {
	for _, ip := range collectCandidates(r) {
		if !isPublicVisitorIP(ip) {
			continue
		}
		if v4 := ip.To4(); v4 != nil {
			return v4
		}
	}
	return nil
}

func preferredIPv6(r *http.Request) net.IP {
	for _, ip := range collectCandidates(r) {
		if !isPublicVisitorIP(ip) {
			continue
		}
		if ip.To4() == nil && ip.To16() != nil {
			return ip
		}
	}
	return nil
}

// collectCandidates returns IPs in trust order for typical CDN → reverse-proxy → app stacks.
// Cloudflare and similar set CF-Connecting-IP / True-Client-IP to the end user; X-Forwarded-For
// may only list the CDN edge (e.g. 172.68.x.x), so those headers must come before XFF.
func collectCandidates(r *http.Request) []net.IP {
	var out []net.IP
	seen := map[string]struct{}{}

	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		ip := net.ParseIP(s)
		if ip == nil {
			return
		}
		key := ip.String()
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		out = append(out, ip)
	}

	add(r.Header.Get("CF-Connecting-IP"))
	add(r.Header.Get("True-Client-IP"))
	add(r.Header.Get("X-Real-IP"))
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		for _, part := range strings.Split(xff, ",") {
			add(strings.TrimSpace(part))
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		add(r.RemoteAddr)
	} else {
		add(host)
	}
	return out
}
