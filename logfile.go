package main

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// configureLogOutput duplicates log output to LOG_FILE (append) when set.
// Stdout is always included so `docker logs` and orchestrators still see output.
// Returns a cleanup function to close the file on exit (optional but tidy).
func configureLogOutput() (cleanup func()) {
	path := strings.TrimSpace(os.Getenv("LOG_FILE"))
	if path == "" {
		return func() {}
	}

	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			log.Printf("warning: LOG_FILE mkdir %s: %v", dir, err)
			return func() {}
		}
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Printf("warning: LOG_FILE open %s: %v", path, err)
		return func() {}
	}

	log.SetOutput(io.MultiWriter(os.Stdout, f))
	return func() { _ = f.Close() }
}
