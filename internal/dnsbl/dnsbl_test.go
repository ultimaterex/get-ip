package dnsbl

import (
	"net"
	"sort"
	"strings"
	"testing"
)

func TestReverseIPv4(t *testing.T) {
	t.Parallel()
	ip := net.ParseIP("190.98.108.105")
	got := reverseIPv4(ip)
	want := "105.108.98.190"
	if got != want {
		t.Fatalf("reverseIPv4: got %q want %q", got, want)
	}
	if reverseIPv4(nil) != "" {
		t.Fatal("nil ip")
	}
}

func TestInterpretReturnAddrs(t *testing.T) {
	t.Parallel()
	ok, codes := interpretReturnAddrs([]string{"127.0.0.9", "127.0.0.4"})
	if !ok || len(codes) != 2 {
		t.Fatalf("listed=%v codes=%v", ok, codes)
	}
	ok2, _ := interpretReturnAddrs([]string{"192.0.2.1"})
	if ok2 {
		t.Fatal("non-127 should not list")
	}
}

func TestParseZones(t *testing.T) {
	t.Parallel()
	s := " zen.spamhaus.org|zen ; b.barracudacentral.org "
	got := parseZones(s)
	if len(got) != 2 {
		t.Fatalf("len=%d", len(got))
	}
	if got[0].zone != "zen.spamhaus.org" || got[0].tag != "zen" {
		t.Errorf("first %+v", got[0])
	}
	if got[1].tag != "b-barracudacentral-org" {
		t.Errorf("tag %+v", got[1])
	}
}

func TestItoaOctet(t *testing.T) {
	t.Parallel()
	if itoa(0) != "0" || itoa(255) != "255" || itoa(16) != "16" {
		t.Fatalf("itoa broken: %q %q %q", itoa(0), itoa(255), itoa(16))
	}
}

func TestEnrichReturnCodesSpamhaus(t *testing.T) {
	t.Parallel()
	got := enrichReturnCodes("zen.spamhaus.org", []string{"127.0.0.4"})
	if len(got) != 1 || !strings.Contains(got[0].Meaning, "XBL") {
		t.Fatalf("%+v", got)
	}
}

func TestSpamhausAdjustPureMirrorError(t *testing.T) {
	t.Parallel()
	chk := Check{Listed: true, ReturnCodes: []string{"127.255.255.254"}}
	spamhausAdjustCheck("zen.spamhaus.org", &chk)
	if chk.Listed || chk.Error == "" || len(chk.ReturnCodes) == 0 {
		t.Fatalf("want mirror error, got %+v", chk)
	}
	chk.ReturnCodeDetails = enrichReturnCodes("zen.spamhaus.org", chk.ReturnCodes)
	if len(chk.ReturnCodeDetails) == 0 || chk.ReturnCodeDetails[0].Meaning == "" {
		t.Fatal("expected meanings")
	}
}

func TestSpamhausAdjustDoesNotTouchOtherZones(t *testing.T) {
	t.Parallel()
	chk := Check{Listed: true, ReturnCodes: []string{"127.255.255.254"}}
	spamhausAdjustCheck("cbl.abuseat.org", &chk)
	if !chk.Listed || chk.Error != "" {
		t.Fatalf("cbl should keep raw listed bit: %+v", chk)
	}
}

func TestCheckReturnSummary(t *testing.T) {
	t.Parallel()
	c := Check{
		Listed: true,
		ReturnCodes: []string{"127.0.0.2"},
		ReturnCodeDetails: []ReturnCodeDetail{
			{Code: "127.0.0.2", Meaning: "SBL"},
		},
	}
	if !strings.Contains(c.ReturnSummary(), "127.0.0.2") || !strings.Contains(c.ReturnSummary(), "SBL") {
		t.Fatal(c.ReturnSummary())
	}
}

func TestChecksSortListedErrorsClean(t *testing.T) {
	t.Parallel()
	res := []Check{
		{Source: "c", Zone: "clean.example", Listed: false},
		{Source: "a", Zone: "listed.example", Listed: true},
		{Source: "b", Zone: "error.example", Error: "timeout"},
	}
	sort.Slice(res, func(i, j int) bool {
		ri, rj := dnsblCheckRank(res[i]), dnsblCheckRank(res[j])
		if ri != rj {
			return ri < rj
		}
		if res[i].Source != res[j].Source {
			return res[i].Source < res[j].Source
		}
		return res[i].Zone < res[j].Zone
	})
	if res[0].Zone != "listed.example" || res[1].Zone != "error.example" || res[2].Zone != "clean.example" {
		t.Fatalf("order: %+v", res)
	}
}
