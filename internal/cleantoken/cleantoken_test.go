package cleantoken

import "testing"

func TestSemicolonPart(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{`"https://example.com/x"`, "https://example.com/x"},
		// YAML-style: leading \" before URL (bytes: \ " h t t p …)
		{"\x5c\x22https://example.com/x\"", "https://example.com/x"},
		{`  "https://a.com"  `, "https://a.com"},
		{`'https://a.com'`, "https://a.com"},
		{`https://raw.githubusercontent.com/x`, "https://raw.githubusercontent.com/x"},
	}
	for _, c := range cases {
		if got := SemicolonPart(c.in); got != c.want {
			t.Errorf("SemicolonPart(%q) = %q want %q", c.in, got, c.want)
		}
	}
}
