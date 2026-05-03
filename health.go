package main

import (
	"fmt"
	"net/http"
)

// handleHealth serves plain-text "ok" for orchestrator probes (Swarm, Kubernetes, Nomad, etc.).
func handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodGet {
		_, _ = fmt.Fprint(w, "ok\n")
	}
}
