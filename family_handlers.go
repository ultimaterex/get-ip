package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Plain-text 503 bodies for GET /ipv4 and GET /ipv6 when no public address exists for that family.
const (
	errIPv4NotAvailable = "IPv4_NOT_AVAILABLE"
	errIPv6NotAvailable = "IPv6_NOT_AVAILABLE"
)

func writeFamilyUnavailable(w http.ResponseWriter, r *http.Request, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusServiceUnavailable)
	if r.Method == http.MethodHead {
		return
	}
	fmt.Fprint(w, body)
}

func handleIPv4Text(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	v4, _ := publicIPv4IPv6(r)
	if v4 == nil {
		log.Printf("%s %s -> %s", r.Method, r.URL.Path, errIPv4NotAvailable)
		writeFamilyUnavailable(w, r, errIPv4NotAvailable)
		return
	}

	log.Printf("%s %s -> %s", r.Method, r.URL.Path, v4.String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	fmt.Fprint(w, v4.String())
}

func handleIPv6Text(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, v6 := publicIPv4IPv6(r)
	if v6 == nil {
		log.Printf("%s %s -> %s", r.Method, r.URL.Path, errIPv6NotAvailable)
		writeFamilyUnavailable(w, r, errIPv6NotAvailable)
		return
	}

	log.Printf("%s %s -> %s", r.Method, r.URL.Path, v6.String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}
	fmt.Fprint(w, v6.String())
}

// handleFamilyJSON serves the same payload as GET /json for GET /ipv4/json and GET /ipv6/json
// (split-DNS hostnames). OPTIONS handles CORS preflight when GET_IP_ACCESS_CONTROL_ALLOW_ORIGIN is set.
func handleFamilyJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		handleJSONCORSOptions(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	resp := buildJSONResponse(r, false)
	v4, v6 := publicIPv4IPv6(r)
	log.Printf("%s %s -> v4=%s v6=%s", r.Method, r.URL.Path, formatIP(v4), formatIP(v6))

	applyJSONCORS(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Printf("json marshal: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Write(out)
	w.Write([]byte("\n"))
}
