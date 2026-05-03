package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleRoot)
	mux.HandleFunc("/all", handleAll)

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
		http.Error(w, "could not determine client IP", http.StatusServiceUnavailable)
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

	candidates := collectCandidates(r)
	var v4, v6 net.IP
	for _, ip := range candidates {
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

	remoteHost, remotePort, _ := net.SplitHostPort(r.RemoteAddr)
	remoteIP := net.ParseIP(remoteHost)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	var b strings.Builder
	fmt.Fprintf(&b, "IPv4: %s\n", formatIP(v4))
	fmt.Fprintf(&b, "IPv6: %s\n", formatIP(v6))
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "Direct connection\n")
	if remoteIP != nil && isPublicVisitorIP(remoteIP) {
		fmt.Fprintf(&b, "  RemoteAddr: %s\n", r.RemoteAddr)
		fmt.Fprintf(&b, "  Parsed IP: %s\n", remoteIP.String())
		if remotePort != "" {
			fmt.Fprintf(&b, "  Port: %s\n", remotePort)
		}
	} else {
		fmt.Fprintf(&b, "  (upstream peer address not shown)\n")
	}
	fmt.Fprintf(&b, "\n")
	writePublicIPHeader(&b, r, "CF-Connecting-IP")
	writePublicIPHeader(&b, r, "True-Client-IP")
	writePublicIPHeader(&b, r, "X-Real-IP")
	writePublicXForwardedFor(&b, r)
	fmt.Fprintf(&b, "\n")
	fmt.Fprintf(&b, "Request\n")
	fmt.Fprintf(&b, "  Method: %s\n", r.Method)
	fmt.Fprintf(&b, "  Host: %s\n", r.Host)
	fmt.Fprintf(&b, "  Protocol: %s\n", r.Proto)
	if ua := r.Header.Get("User-Agent"); ua != "" {
		fmt.Fprintf(&b, "  User-Agent: %s\n", ua)
	}

	fmt.Fprint(w, b.String())
}

// isPublicVisitorIP reports IPs safe to show visitors: globally routable unicast (not RFC1918,
// ULA, loopback, link-local, CGNAT, etc.). Uses net.IP.IsGlobalUnicast.
func isPublicVisitorIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast()
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
	v := r.Header.Get("X-Forwarded-For")
	if v == "" {
		return
	}
	var parts []string
	for _, p := range strings.Split(v, ",") {
		p = strings.TrimSpace(p)
		ip := net.ParseIP(p)
		if ip != nil && isPublicVisitorIP(ip) {
			parts = append(parts, ip.String())
		}
	}
	if len(parts) == 0 {
		return
	}
	fmt.Fprintf(b, "  X-Forwarded-For: %s\n", strings.Join(parts, ", "))
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
