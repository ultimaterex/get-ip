package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIPv4Text_publicFromHeader(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/ipv4", nil)
	req.Header.Set("CF-Connecting-IP", "8.8.8.8")
	rec := httptest.NewRecorder()
	handleIPv4Text(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code=%d body=%q", rec.Code, rec.Body.String())
	}
	if strings.TrimSpace(rec.Body.String()) != "8.8.8.8" {
		t.Fatalf("body=%q", rec.Body.String())
	}
}

func TestIPv6Text_publicFromHeader(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/ipv6", nil)
	req.Header.Set("CF-Connecting-IP", "2001:4860:4860::8888")
	rec := httptest.NewRecorder()
	handleIPv6Text(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code=%d body=%q", rec.Code, rec.Body.String())
	}
	want := "2001:4860:4860::8888"
	if strings.TrimSpace(rec.Body.String()) != want {
		t.Fatalf("body=%q want %q", rec.Body.String(), want)
	}
}

func TestIPv4Text_missing(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/ipv4", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	handleIPv4Text(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != errIPv4NotAvailable {
		t.Fatalf("body=%q want %q", rec.Body.String(), errIPv4NotAvailable)
	}
}

func TestIPv6Text_missing(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/ipv6", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	handleIPv6Text(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("code=%d", rec.Code)
	}
	if strings.TrimSpace(rec.Body.String()) != errIPv6NotAvailable {
		t.Fatalf("body=%q want %q", rec.Body.String(), errIPv6NotAvailable)
	}
}

func TestFamilyJSON_CORSOptions(t *testing.T) {
	t.Setenv("GET_IP_ACCESS_CONTROL_ALLOW_ORIGIN", "")
	req := httptest.NewRequest("OPTIONS", "/ipv4/json", nil)
	rec := httptest.NewRecorder()
	handleFamilyJSON(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("without env, OPTIONS got %d", rec.Code)
	}

	t.Setenv("GET_IP_ACCESS_CONTROL_ALLOW_ORIGIN", "https://ip.example.com")
	req2 := httptest.NewRequest("OPTIONS", "/ipv4/json", nil)
	rec2 := httptest.NewRecorder()
	handleFamilyJSON(rec2, req2)
	if rec2.Code != 204 {
		t.Fatalf("code=%d", rec2.Code)
	}
	if rec2.Header().Get("Access-Control-Allow-Origin") != "https://ip.example.com" {
		t.Fatalf("cors header missing")
	}
}

func TestFamilyJSON_GET(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/ipv4/json", nil)
	req.Header.Set("CF-Connecting-IP", "8.8.8.8")
	rec := httptest.NewRecorder()
	handleFamilyJSON(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code=%d", rec.Code)
	}
	b, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(b), `"ipv4": "8.8.8.8"`) {
		t.Fatalf("json missing ipv4: %s", b)
	}
}
