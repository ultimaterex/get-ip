package main

import "testing"

func TestBlocklistsJSONURLFromDualJSONURL(t *testing.T) {
	t.Parallel()
	got := blocklistsJSONURLFromDualJSONURL("https://ipv4.ip.example.com/json")
	want := "https://ipv4.ip.example.com/blocklists/json"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
	if blocklistsJSONURLFromDualJSONURL("") != "" {
		t.Fatal("empty in should give empty out")
	}
	if blocklistsJSONURLFromDualJSONURL("not-a-url") != "" {
		t.Fatal("bad url should give empty")
	}
}
