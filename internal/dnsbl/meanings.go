package dnsbl

import (
	"strings"
)

// enrichReturnCodes attaches human-readable notes for known DNSBL return addresses (mostly 127.0.0.0/8).
func enrichReturnCodes(zone string, codes []string) []ReturnCodeDetail {
	if len(codes) == 0 {
		return nil
	}
	z := normZone(zone)
	out := make([]ReturnCodeDetail, 0, len(codes))
	for _, c := range codes {
		out = append(out, ReturnCodeDetail{
			Code:    c,
			Meaning: meaningForZoneCode(z, c),
		})
	}
	return out
}

func normZone(z string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(z), "."))
}

func isSpamhausZone(zone string) bool {
	return strings.Contains(normZone(zone), "spamhaus.org")
}

// spamhausAdjustCheck corrects pure Spamhaus mirror/query error responses (127.255.255.x)
// so they are not treated as reputation listings. See Spamhaus DNSBL FAQ.
func spamhausAdjustCheck(zone string, chk *Check) {
	if !isSpamhausZone(zone) || len(chk.ReturnCodes) == 0 {
		return
	}
	for _, c := range chk.ReturnCodes {
		if c != "127.255.255.252" && c != "127.255.255.254" && c != "127.255.255.255" {
			return
		}
	}
	chk.Listed = false
	chk.Error = spamhausMirrorErrorMessage(chk.ReturnCodes)
}

func spamhausMirrorErrorMessage(codes []string) string {
	for _, c := range codes {
		switch c {
		case "127.255.255.254":
			return "Spamhaus mirror: query via public/open resolver blocked — not a reputation response (Spamhaus DNSBL FAQ)."
		case "127.255.255.255":
			return "Spamhaus mirror: excessive queries — not a reputation response (Spamhaus DNSBL FAQ)."
		case "127.255.255.252":
			return "Spamhaus mirror: DNSBL name typo — not a reputation response (Spamhaus DNSBL FAQ)."
		}
	}
	return "Spamhaus mirror: query error — not a reputation response (Spamhaus DNSBL FAQ)."
}

func meaningForZoneCode(zone, code string) string {
	if m := zoneExactTable[zone][code]; m != "" {
		return m
	}
	if isSpamhausZone(zone) {
		if m := spamhausIPMeanings[code]; m != "" {
			return m
		}
		if strings.HasPrefix(code, "127.255.255.") {
			return spamhausErrorRange(code)
		}
	}
	if strings.HasPrefix(code, "127.") {
		return "127/8 response — see this zone publisher’s DNSBL return-code documentation."
	}
	return ""
}

func spamhausErrorRange(code string) string {
	switch code {
	case "127.255.255.252":
		return "Spamhaus: DNSBL name typo (not a reputation listing)."
	case "127.255.255.254":
		return "Spamhaus: query via public/open resolver blocked (not a reputation listing)."
	case "127.255.255.255":
		return "Spamhaus: excessive queries (not a reputation listing)."
	default:
		return "Spamhaus: error range 127.255.255.0/24 — not a reputation listing."
	}
}

// Spamhaus IP zone return codes — https://www.spamhaus.org/faq/section/DNSBL%20Usage#200
var spamhausIPMeanings = map[string]string{
	"127.0.0.2":  "SBL — Spamhaus blocklist data",
	"127.0.0.3":  "SBL — Spamhaus CSS (combined spam sources) data",
	"127.0.0.4":  "XBL — exploit/malware-infected (includes CBL data)",
	"127.0.0.9":  "SBL — Spamhaus DROP data (as well as 127.0.0.2 since 2016)",
	"127.0.0.10": "PBL — ISP-maintained policy block",
	"127.0.0.11": "PBL — Spamhaus-maintained policy block",
}

func joinMeaningLine(ds []ReturnCodeDetail) string {
	var b strings.Builder
	for i, d := range ds {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(d.Code)
		if d.Meaning != "" {
			b.WriteString(" — ")
			b.WriteString(d.Meaning)
		}
	}
	return b.String()
}

var zoneExactTable = map[string]map[string]string{
	"cbl.abuseat.org": {
		"127.0.0.2":       "CBL — infected / spam-sending host (see abuseat.org).",
		"127.255.255.254": "Unusual for this zone — often a resolver/forwarder artifact; verify with dig or abuseat.org.",
	},
	"b.barracudacentral.org": {
		"127.0.0.2": "Barracuda — listed as a spam source (see Barracuda Central for detail/removal).",
	},
}
