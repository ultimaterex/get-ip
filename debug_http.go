package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"os"
	"sort"
	"strings"
	"time"
)

// debugHTTPEnabled reports whether GET_IP_DEBUG_HTTP is set (gated /debug JSON for troubleshooting client IP).
func debugHTTPEnabled() bool {
	s := strings.TrimSpace(strings.ToLower(os.Getenv("GET_IP_DEBUG_HTTP")))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

// handleDebug returns JSON describing how the server derived (or failed to derive) the client IP.
// Only registered when GET_IP_DEBUG_HTTP is enabled — do not expose in production without understanding the data leak risk.
func handleDebug(w http.ResponseWriter, r *http.Request) {
	if !debugHTTPEnabled() {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	switch r.URL.Path {
	case "/debug", "/debug/json":
	default:
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}

	resp := buildDebugResponse(r)
	logDebugHTTP(r, resp)

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(out)
	_, _ = w.Write([]byte("\n"))
}

type debugRemote struct {
	Raw        string `json:"raw"`
	Host       string `json:"host,omitempty"`
	Port       string `json:"port,omitempty"`
	SplitError string `json:"split_error,omitempty"`
}

type debugCandidateOut struct {
	Source       string  `json:"source"`
	Raw          string  `json:"raw,omitempty"`
	Parsed       *string `json:"parsed,omitempty"`
	IsPublic     bool    `json:"is_public"`
	RejectReason string  `json:"reject_reason,omitempty"`
}

type debugTLS struct {
	Version            string `json:"version,omitempty"`
	CipherSuite        string `json:"cipher_suite,omitempty"`
	NegotiatedProtocol string `json:"negotiated_protocol,omitempty"`
	ServerName         string `json:"server_name,omitempty"`
	PeerCertificates   int    `json:"peer_certificates,omitempty"`
}

type debugHTTPResponse struct {
	Now       string `json:"now"`
	Endpoint  string `json:"endpoint"`
	Remote    debugRemote         `json:"remote_addr"`
	Candidates []debugCandidateOut `json:"ip_candidates"`
	Diagnostics struct {
		PublicIPv4      *string `json:"public_ipv4,omitempty"`
		PublicIPv6      *string `json:"public_ipv6,omitempty"`
		PreferredIPv4   *string `json:"preferred_ipv4,omitempty"`
		PreferredIPv6   *string `json:"preferred_ipv6,omitempty"`
		RootWouldFail   bool    `json:"root_would_fail"`
		RootFailReason  string  `json:"root_fail_reason,omitempty"`
		DNSBLClientKey  string  `json:"dnsbl_client_key,omitempty"`
	} `json:"diagnostics"`
	Request struct {
		Method       string `json:"method"`
		RequestURI   string `json:"request_uri"`
		Proto        string `json:"proto"`
		Host         string `json:"host"`
		ContentLength int64 `json:"content_length"`
		RemoteAddr   string `json:"remote_addr_raw"`
	} `json:"request"`
	Headers         map[string][]string `json:"headers"`
	ForwardedParsed *jsonForwarded      `json:"forwarded_parsed,omitempty"`
	TLS             *debugTLS           `json:"tls,omitempty"`
	Hints           []string            `json:"hints"`
}

func buildDebugResponse(r *http.Request) debugHTTPResponse {
	var out debugHTTPResponse
	out.Now = time.Now().UTC().Format(time.RFC3339Nano)
	out.Endpoint = "GET_IP_DEBUG_HTTP"
	out.Hints = []string{
		"If every candidate is private/CGNAT (100.64.0.0/10) or RFC1918, the reverse proxy may be stripping X-Forwarded-For or the visitor is behind carrier-grade NAT.",
		"If RemoteAddr is a private IP, the app is behind a reverse proxy — configure the proxy to set X-Forwarded-For or X-Real-IP with the client public address.",
		"Cloudflare and similar must send CF-Connecting-IP (or equivalent); X-Forwarded-For alone may only contain the CDN edge.",
		"Disable GET_IP_DEBUG_HTTP when finished troubleshooting — this response can include sensitive headers (Authorization/Cookie are redacted).",
	}

	out.Request.Method = r.Method
	out.Request.RequestURI = r.RequestURI
	out.Request.Proto = r.Proto
	out.Request.Host = r.Host
	out.Request.ContentLength = r.ContentLength
	out.Request.RemoteAddr = r.RemoteAddr

	out.Remote.Raw = r.RemoteAddr
	host, port, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		out.Remote.SplitError = err.Error()
	} else {
		out.Remote.Host = host
		out.Remote.Port = port
	}

	for _, c := range collectIPCandidates(r) {
		row := debugCandidateOut{Source: c.Source, Raw: c.Raw}
		if c.IP == nil {
			row.RejectReason = "not_a_valid_ip"
			out.Candidates = append(out.Candidates, row)
			continue
		}
		s := c.IP.String()
		row.Parsed = &s
		pub := isPublicVisitorIP(c.IP)
		row.IsPublic = pub
		if !pub {
			row.RejectReason = visitorIPRejectReason(c.IP)
		}
		out.Candidates = append(out.Candidates, row)
	}

	v4, v6 := publicIPv4IPv6(r)
	if v4 != nil {
		s := v4.String()
		out.Diagnostics.PublicIPv4 = &s
	}
	if v6 != nil {
		s := v6.String()
		out.Diagnostics.PublicIPv6 = &s
	}
	if p4 := preferredIPv4(r); p4 != nil {
		s := p4.String()
		out.Diagnostics.PreferredIPv4 = &s
	}
	if p6 := preferredIPv6(r); p6 != nil {
		s := p6.String()
		out.Diagnostics.PreferredIPv6 = &s
	}
	out.Diagnostics.DNSBLClientKey = dnsblClientKey(r)

	if out.Diagnostics.PreferredIPv4 == nil && out.Diagnostics.PreferredIPv6 == nil {
		out.Diagnostics.RootWouldFail = true
		out.Diagnostics.RootFailReason = "no_public_ipv4_or_ipv6_after_headers_and_remote_addr"
	}

	out.Headers = redactHeadersForDebug(r)
	out.ForwardedParsed = buildJSONForwarded(r)

	if r.TLS != nil {
		out.TLS = &debugTLS{
			Version:            tlsVersionString(r.TLS.Version),
			CipherSuite:        fmt.Sprintf("0x%04x", r.TLS.CipherSuite),
			NegotiatedProtocol: r.TLS.NegotiatedProtocol,
			ServerName:         r.TLS.ServerName,
			PeerCertificates:   len(r.TLS.PeerCertificates),
		}
	}

	return out
}

func tlsVersionString(v uint16) string {
	switch v {
	case tls.VersionTLS13:
		return "TLS 1.3"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS10:
		return "TLS 1.0"
	case 0x0300:
		return "SSL 3.0"
	default:
		return fmt.Sprintf("0x%04x", v)
	}
}

// logDebugHTTP writes a multi-line snapshot to the process log (stdout by default).
// Sensitive header values use the same redaction rules as the JSON body.
func logDebugHTTP(r *http.Request, resp debugHTTPResponse) {
	log.Printf("debug: --- begin snapshot at=%s ---", resp.Now)
	log.Printf("debug: request method=%q request_uri=%q proto=%q host=%q remote_addr=%q content_length=%d",
		resp.Request.Method, resp.Request.RequestURI, resp.Request.Proto, resp.Request.Host, resp.Request.RemoteAddr, resp.Request.ContentLength)
	if r.URL != nil {
		log.Printf("debug: url path=%q raw_query=%q", r.URL.Path, r.URL.RawQuery)
	}
	log.Printf("debug: remote parsed raw=%q host=%q port=%q split_error=%q",
		resp.Remote.Raw, resp.Remote.Host, resp.Remote.Port, resp.Remote.SplitError)

	// Raw Header.Get for common client-IP carriers (duplicate of map; easy to grep logs).
	for _, h := range []string{
		"CF-Connecting-IP", "True-Client-IP", "X-Real-IP", "X-Forwarded-For", "Forwarded",
		"X-Client-IP", "Fastly-Client-IP", "Fly-Client-IP", "X-Forwarded", "X-Cluster-Client-IP",
	} {
		if v := strings.TrimSpace(r.Header.Get(h)); v != "" {
			log.Printf("debug: header_get %s=%q", h, v)
		}
	}

	for _, c := range resp.Candidates {
		parsed := ""
		if c.Parsed != nil {
			parsed = *c.Parsed
		}
		log.Printf("debug: candidate source=%q raw=%q parsed=%q is_public=%v reject_reason=%q",
			c.Source, c.Raw, parsed, c.IsPublic, c.RejectReason)
	}

	d := resp.Diagnostics
	log.Printf("debug: result public_ipv4=%v public_ipv6=%v preferred_ipv4=%v preferred_ipv6=%v root_would_fail=%v root_fail_reason=%q dnsbl_client_key=%q",
		derefStr(d.PublicIPv4), derefStr(d.PublicIPv6), derefStr(d.PreferredIPv4), derefStr(d.PreferredIPv6),
		d.RootWouldFail, d.RootFailReason, d.DNSBLClientKey)

	if resp.ForwardedParsed != nil {
		b, _ := json.Marshal(resp.ForwardedParsed)
		log.Printf("debug: forwarded_parsed_json=%s", string(b))
	}

	if resp.TLS != nil {
		log.Printf("debug: tls version=%q cipher=%q alpn=%q sni=%q peer_certs=%d",
			resp.TLS.Version, resp.TLS.CipherSuite, resp.TLS.NegotiatedProtocol, resp.TLS.ServerName, resp.TLS.PeerCertificates)
	}

	keys := make([]string, 0, len(resp.Headers))
	for k := range resp.Headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		for _, v := range resp.Headers[k] {
			log.Printf("debug: header %s: %s", k, v)
		}
	}
	log.Printf("debug: --- end snapshot ---")
}

func derefStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func redactHeadersForDebug(r *http.Request) map[string][]string {
	out := make(map[string][]string)
	for k, vv := range r.Header {
		can := textproto.CanonicalMIMEHeaderKey(k)
		low := strings.ToLower(k)
		switch low {
		case "authorization", "cookie", "proxy-authorization", "x-api-key":
			out[can] = []string{"[redacted]"}
			continue
		default:
			cp := make([]string, len(vv))
			copy(cp, vv)
			out[can] = cp
		}
	}
	return out
}
