package main

import (
	"net/http"
	"os"
	"strings"
)

// jsonAccessControlAllowOrigin returns GET_IP_ACCESS_CONTROL_ALLOW_ORIGIN when set (single origin
// for Access-Control-Allow-Origin on JSON responses). Required for browser dual-fetch from a main
// hostname to ipv4.* / ipv6.* JSON URLs (cross-origin GET + Accept: application/json).
func jsonAccessControlAllowOrigin() string {
	return strings.TrimSpace(os.Getenv("GET_IP_ACCESS_CONTROL_ALLOW_ORIGIN"))
}

func applyJSONCORS(w http.ResponseWriter) {
	if o := jsonAccessControlAllowOrigin(); o != "" {
		w.Header().Set("Access-Control-Allow-Origin", o)
		w.Header().Set("Vary", "Origin")
	}
}

func handleJSONCORSOptions(w http.ResponseWriter, r *http.Request) {
	if jsonAccessControlAllowOrigin() == "" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	applyJSONCORS(w)
	w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept")
	w.WriteHeader(http.StatusNoContent)
}
