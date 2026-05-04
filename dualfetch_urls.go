package main

import (
	"net/url"
	"os"
	"strings"
)

// envDualFetchJSONURLs returns GET_IP_DUAL_FETCH_IPV4_JSON_URL and GET_IP_DUAL_FETCH_IPV6_JSON_URL.
func envDualFetchJSONURLs() (v4, v6 string) {
	v4 = strings.TrimSpace(os.Getenv("GET_IP_DUAL_FETCH_IPV4_JSON_URL"))
	v6 = strings.TrimSpace(os.Getenv("GET_IP_DUAL_FETCH_IPV6_JSON_URL"))
	return v4, v6
}

// blocklistsJSONURLFromDualJSONURL returns the same scheme/host as jsonURL with path /blocklists/json,
// for parallel fetch on ipv4.ip / ipv6.ip hosts when dual-fetch JSON URLs are configured.
func blocklistsJSONURLFromDualJSONURL(jsonURL string) string {
	jsonURL = strings.TrimSpace(jsonURL)
	if jsonURL == "" {
		return ""
	}
	u, err := url.Parse(jsonURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	u.Path = "/blocklists/json"
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}
