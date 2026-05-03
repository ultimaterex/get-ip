// Package cleantoken normalizes semicolon-separated env values (BLOCKLIST_URLS, DNSBL_ZONES).
package cleantoken

import "strings"

// SemicolonPart strips accidental surrounding quotes and YAML-style leading \"
// that often appear when values are set via docker-compose or shell-escaped .env.
func SemicolonPart(s string) string {
	s = strings.TrimSpace(s)
	for i := 0; i < 4; i++ {
		n := len(s)
		if n == 0 {
			return s
		}
		if n >= 2 && s[0] == '\\' && s[1] == '"' {
			s = strings.TrimSpace(s[2:])
			continue
		}
		if n >= 2 && s[0] == '"' && s[n-1] == '"' {
			s = strings.TrimSpace(s[1 : n-1])
			continue
		}
		if n >= 2 && s[0] == '\'' && s[n-1] == '\'' {
			s = strings.TrimSpace(s[1 : n-1])
			continue
		}
		break
	}
	s = strings.TrimSpace(s)
	// After stripping \"…\" prefix, a lone trailing " often remains (compose .env).
	for i := 0; i < 2; i++ {
		s = strings.TrimPrefix(s, `"`)
		s = strings.TrimSuffix(s, `"`)
		s = strings.TrimSpace(s)
	}
	return s
}
