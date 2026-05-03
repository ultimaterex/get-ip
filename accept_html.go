package main

import (
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// prefersHTML reports whether the client asked for an HTML document via Accept.
// Browsers send text/html; curl and many APIs use */* or application/json without text/html.
func prefersHTML(r *http.Request) bool {
	return acceptPrefersHTML(r.Header.Get("Accept"))
}

func acceptPrefersHTML(accept string) bool {
	accept = strings.TrimSpace(accept)
	if accept == "" {
		return false
	}
	for _, part := range strings.Split(accept, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		mediatype, params, err := mime.ParseMediaType(part)
		if err != nil {
			continue
		}
		if mediatype != "text/html" {
			continue
		}
		if q, ok := params["q"]; ok {
			f, err := strconv.ParseFloat(q, 64)
			if err == nil && f == 0 {
				return false
			}
		}
		return true
	}
	return false
}
