package main

import (
	"testing"
)

func TestAcceptPrefersHTML(t *testing.T) {
	tests := []struct {
		accept string
		want   bool
	}{
		{"", false},
		{"*/*", false},
		{"application/json", false},
		{"text/html", true},
		{"text/html, application/xhtml+xml, application/xml;q=0.9, */*;q=0.8", true},
		{"text/html;q=0, */*", false},
		{"text/plain, */*", false},
	}
	for _, tt := range tests {
		if got := acceptPrefersHTML(tt.accept); got != tt.want {
			t.Errorf("acceptPrefersHTML(%q) = %v, want %v", tt.accept, got, tt.want)
		}
	}
}
