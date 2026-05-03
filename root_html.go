package main

import (
	"fmt"
	"html"
	"net/http"
)

func writeRootHTML(w http.ResponseWriter, r *http.Request, ipStr string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodHead {
		return
	}

	esc := html.EscapeString(ipStr)
	fmt.Fprintf(w, rootHTMLDoc, esc)
}

// One printf verb: escaped primary IP.
const rootHTMLDoc = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Your IP</title>
<style>
:root {
  --bg: #0f1419;
  --fg: #e7e9ea;
  --muted: #8b98a5;
  --accent: #1d9bf0;
  --card: #16181c;
}
@media (prefers-color-scheme: light) {
  :root {
    --bg: #f7f9f9;
    --fg: #0f1419;
    --muted: #536471;
    --accent: #1d9bf0;
    --card: #ffffff;
  }
}
* { box-sizing: border-box; }
body {
  margin: 0;
  min-height: 100vh;
  font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif;
  background: var(--bg);
  color: var(--fg);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}
main {
  width: 100%%;
  max-width: 28rem;
  text-align: center;
}
.ip {
  margin: 0 0 1.25rem;
  font-size: clamp(1.5rem, 6vw, 2rem);
  font-weight: 600;
  letter-spacing: 0.02em;
  word-break: break-all;
  padding: 1.25rem 1rem;
  background: var(--card);
  border-radius: 12px;
  border: 1px solid color-mix(in srgb, var(--muted) 25%%, transparent);
}
.links {
  margin: 0;
  font-size: 0.9375rem;
  color: var(--muted);
}
.links a {
  color: var(--accent);
  text-decoration: none;
}
.links a:hover { text-decoration: underline; }
.sep { margin: 0 0.35em; color: var(--muted); }
</style>
</head>
<body>
<main>
  <p class="ip" translate="no">%s</p>
  <p class="links"><a href="/all">Details</a><span class="sep">·</span><a href="/json">JSON</a></p>
</main>
</body>
</html>
`
