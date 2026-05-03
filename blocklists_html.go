package main

import (
	"html"
	"net/http"
	"strings"
)

func writeBlocklistsHTML(w http.ResponseWriter, r *http.Request, primaryIP string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Link", `</blocklists/json>; rel="alternate"; type="application/json"`)
	w.Header().Add("Link", `</json>; rel="alternate"; type="application/json"`)
	if r.Method == http.MethodHead {
		return
	}

	esc := html.EscapeString(primaryIP)
	out := strings.ReplaceAll(blocklistsHTMLTemplate, "__PRIMARY_IP__", esc)
	_, _ = w.Write([]byte(out))
}

// blocklistsHTMLTemplate: HTTP prefix feeds + DNSBL; fetch /blocklists/json.
const blocklistsHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Blocklists · get-ip</title>
<style>
:root {
  --bg: #0c1014;
  --bg2: #121820;
  --fg: #e8ecf0;
  --muted: #8b9aab;
  --accent: #3ba4ff;
  --accent-dim: color-mix(in srgb, var(--accent) 35%, transparent);
  --card: #151b22;
  --bd: color-mix(in srgb, var(--muted) 22%, transparent);
  --shadow: 0 8px 32px rgba(0, 0, 0, 0.35);
  --bad: #f5495f;
  --ok: #3dd68c;
}
@media (prefers-color-scheme: light) {
  :root {
    --bg: #f3f6f9;
    --bg2: #e9eef4;
    --fg: #0f1419;
    --muted: #5c6d7e;
    --accent: #0d7bd4;
    --accent-dim: color-mix(in srgb, var(--accent) 18%, transparent);
    --card: #ffffff;
    --bd: color-mix(in srgb, var(--muted) 18%, transparent);
    --shadow: 0 8px 28px rgba(15, 20, 25, 0.08);
    --bad: #d42a3e;
    --ok: #0d8f5b;
  }
}
* { box-sizing: border-box; }
body {
  margin: 0;
  min-height: 100%;
  font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif;
  background: radial-gradient(1200px 600px at 80% -10%, var(--accent-dim), transparent 55%),
    radial-gradient(900px 500px at -10% 50%, color-mix(in srgb, var(--muted) 15%, transparent), transparent 50%),
    var(--bg);
  color: var(--fg);
  padding: 0 1rem 2rem;
}
.shell { max-width: 920px; margin: 0 auto; padding-top: clamp(1rem, 3vw, 1.75rem); }
.nav {
  font-size: 0.875rem;
  color: var(--muted);
  margin-bottom: 1rem;
}
.nav a { color: var(--accent); text-decoration: none; font-weight: 500; }
.nav a:hover { text-decoration: underline; }
.sep { margin: 0 0.45em; opacity: 0.7; }
h1 {
  margin: 0 0 0.35rem;
  font-size: clamp(1.5rem, 4vw, 2rem);
  font-weight: 700;
  letter-spacing: -0.02em;
}
.lead {
  margin: 0 0 1.25rem;
  font-size: 0.95rem;
  color: var(--muted);
  line-height: 1.5;
  max-width: 62ch;
}
.card {
  background: var(--card);
  border: 1px solid var(--bd);
  border-radius: 14px;
  padding: 1rem 1.1rem;
  box-shadow: var(--shadow);
  margin-bottom: 1rem;
}
.card-h {
  margin: 0 0 0.75rem;
  font-size: 0.72rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.09em;
  color: var(--muted);
}
.status { font-size: 0.9rem; color: var(--muted); padding: 0.25rem 0 0.75rem; }
.status.err { color: var(--bad); }
.pill {
  display: inline-block;
  font-size: 0.72rem;
  font-weight: 600;
  padding: 0.2rem 0.55rem;
  border-radius: 999px;
  margin-right: 0.35rem;
  vertical-align: middle;
}
.pill.bad { background: color-mix(in srgb, var(--bad) 22%, transparent); color: var(--bad); }
.pill.ok { background: color-mix(in srgb, var(--ok) 22%, transparent); color: var(--ok); }
.pill.neutral { background: color-mix(in srgb, var(--muted) 18%, transparent); color: var(--muted); }
.meta { font-size: 0.8125rem; color: var(--muted); margin-bottom: 0.75rem; }
table.data {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}
table.data th, table.data td {
  text-align: left;
  padding: 0.45rem 0.5rem;
  border-bottom: 1px solid var(--bd);
  vertical-align: top;
}
table.data th.col-num, table.data td.col-num {
  text-align: right;
  white-space: nowrap;
  width: 5rem;
}
table.data th {
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--muted);
}
.zone-tag { font-size: 0.72rem; color: var(--muted); margin-top: 0.15rem; }
.mono { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; font-size: 0.78em; }
.code-row { margin-bottom: 0.35rem; }
.code-row:last-child { margin-bottom: 0; }
.code-meaning { font-size: 0.85rem; color: var(--muted); line-height: 1.35; margin-top: 0.2rem; padding-left: 0.05rem; }
footer.site {
  margin-top: 2rem;
  text-align: center;
  font-size: 0.75rem;
  color: var(--muted);
}

body.page-loading {
  background: var(--bg);
}
body.page-loading::before {
  content: "";
  position: fixed;
  inset: 0;
  pointer-events: none;
  z-index: 0;
  background: linear-gradient(
    100deg,
    transparent 36%,
    color-mix(in srgb, var(--accent) 18%, transparent) 50%,
    transparent 64%
  );
  background-size: 240% 100%;
  animation: page-load-glimmer 1.75s ease-in-out infinite;
}
@keyframes page-load-glimmer {
  0% { background-position: 100% 0; }
  100% { background-position: -100% 0; }
}
@media (prefers-reduced-motion: reduce) {
  body.page-loading::before {
    animation: none;
    opacity: 0.5;
    background-position: 50% 0;
  }
}
body.page-loading .shell,
body.page-loading footer.site {
  position: relative;
  z-index: 1;
}
</style>
</head>
<body class="page-loading">
<div class="shell">
  <nav class="nav" aria-label="Primary">
    <a href="/">Home</a><span class="sep">·</span><a href="/blocklists/all">Plain report</a><span class="sep">·</span><a href="/blocklists/json" target="_blank" rel="noopener">Raw JSON</a>
  </nav>
  <header>
    <h1 translate="no">__PRIMARY_IP__</h1>
    <p class="lead">Blocklist results for your IP.</p>
  </header>
  <div class="status" id="status">Loading blocklist checks…</div>
  <div id="panels" hidden></div>
</div>
<footer class="site">
  © Hosted on the swarm <span class="mono" aria-hidden="true">(⌐■_■)</span>
</footer>
<script>
(function () {
  var statusEl = document.getElementById('status');
  var panelsEl = document.getElementById('panels');

  function clearPageLoading() {
    document.body.classList.remove('page-loading');
  }

  function esc(s) {
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  function pill(label, cls) {
    return '<span class="pill ' + cls + '">' + esc(label) + '</span>';
  }

  function card(title, inner) {
    if (!inner || !String(inner).trim()) return '';
    var slug = esc(title).replace(/\s+/g, '-').replace(/[^a-zA-Z0-9\-]/g, '');
    return '<section class="card" aria-labelledby="h-' + slug + '"><h2 class="card-h" id="h-' + slug + '">' + esc(title) + '</h2>' + inner + '</section>';
  }

  function renderBlocklists(bl) {
    if (!bl) return '<p class="meta">No <code>BLOCKLIST_URLS</code> configured.</p>';
    var listed = bl.listed ? pill('Listed', 'bad') : pill('Not listed', 'ok');
    var meta = '<div class="meta">' + listed +
      ' Sources loaded: <span class="mono">' + (bl.sources_loaded != null ? esc(String(bl.sources_loaded)) : '—') + '</span>' +
      (bl.last_refresh ? ' · Last refresh <span class="mono">' + esc(bl.last_refresh) + '</span>' : '') + '</div>';
    var rows = '';
    if (bl.matches && bl.matches.length) {
      rows = '<table class="data"><thead><tr><th>Source</th><th>IP</th><th>Prefix</th><th>Family</th></tr></thead><tbody>';
      bl.matches.forEach(function (m) {
        rows += '<tr><td class="mono">' + esc(m.source || '') + '</td><td class="mono">' + esc(m.ip || '') + '</td><td class="mono">' + esc(m.prefix || '') + '</td><td>' + esc(m.family || '') + '</td></tr>';
      });
      rows += '</tbody></table>';
    } else {
      rows = '<p class="meta">No prefix hits for your IPs.</p>';
    }
    return meta + rows;
  }

  function renderDnsbl(d) {
    if (!d) return '<p class="meta">No <code>DNSBL_ZONES</code> configured.</p>';
    if (!d.eligible && (d.rate_limited || d.skipped_reason === 'rate_limited')) {
      return '<p class="meta">' + pill('Rate limited', 'bad') + ' Too many fresh DNSBL lookups. Retry later or relax server limits.</p>';
    }
    if (!d.eligible) {
      var msg = d.skipped_reason === 'no_public_ipv4'
        ? 'No public IPv4 — classic DNSBL queries need IPv4.'
        : esc(d.skipped_reason || 'skipped');
      return '<p class="meta">' + pill('Skipped', 'neutral') + ' ' + msg + '</p>';
    }
    var head = '';
    if (d.cached) head += pill('From cache', 'neutral');
    head += (d.listed ? pill('Listed on ≥1 zone', 'bad') : pill('Clean on sampled zones', 'ok'));
    head = '<div class="meta">' + head + ' · Zones: <span class="mono">' + esc(String(d.zones_checked)) + '</span> · Subject <span class="mono">' + esc(d.ipv4 || '') + '</span>';
    if (d.cache_expires) head += ' · cache expires <span class="mono">' + esc(d.cache_expires) + '</span>';
    head += '</div>';

    function dnsblDetailCell(c) {
      if (c.error) {
        var errLine = esc(c.error);
        if (c.return_code_details && c.return_code_details.length) {
          errLine += '<br>' + c.return_code_details.map(function (x) {
            var m = x.meaning ? (' — ' + esc(x.meaning)) : '';
            return '<span class="mono">' + esc(x.code) + '</span>' + m;
          }).join('<br>');
        }
        return errLine;
      }
      if (c.listed) {
        if (c.return_code_details && c.return_code_details.length) {
          return c.return_code_details.map(function (x) {
            var m = x.meaning ? ('<div class="code-meaning">' + esc(x.meaning) + '</div>') : '';
            return '<div class="code-row"><span class="mono">' + esc(x.code) + '</span>' + m + '</div>';
          }).join('');
        }
        if (c.return_codes && c.return_codes.length) {
          return '<span class="mono">' + esc(c.return_codes.join(', ')) + '</span>';
        }
        return '—';
      }
      return '—';
    }

    var rows = '<table class="data"><thead><tr><th>Zone</th><th>Status</th><th>Detail</th><th class="col-num">RTT (ms)</th></tr></thead><tbody>';
    if (d.checks && d.checks.length) {
      d.checks.forEach(function (c) {
        var st, detail;
        if (c.error) {
          st = pill('error', 'bad');
          detail = dnsblDetailCell(c);
        } else if (c.listed) {
          st = pill('LISTED', 'bad');
          detail = dnsblDetailCell(c);
        } else {
          st = pill('ok', 'ok');
          detail = '—';
        }
        rows += '<tr><td><span class="mono">' + esc(c.zone || '') + '</span><div class="zone-tag">' + esc(c.source || '') + '</div></td><td>' + st + '</td><td>' + detail + '</td><td class="mono col-num">' + esc(String(c.response_ms)) + '</td></tr>';
      });
    }
    rows += '</tbody></table>';
    return head + rows;
  }

  function render(j) {
    var html = '';
    html += card('HTTP prefix blocklists', renderBlocklists(j.blocklists));
    html += card('DNS blocklists (DNSBL)', renderDnsbl(j.dnsbl));
    panelsEl.innerHTML = html || '<p class="meta">Nothing to show.</p>';
    statusEl.hidden = true;
    panelsEl.hidden = false;
    clearPageLoading();
  }

  function fail(msg) {
    clearPageLoading();
    statusEl.classList.add('err');
    statusEl.textContent = msg;
  }

  function runFetch() {
    fetch('/blocklists/json', { credentials: 'same-origin', headers: { Accept: 'application/json' } })
      .then(function (r) {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.json();
      })
      .then(render)
      .catch(function (e) {
        fail('Could not load /blocklists/json: ' + e.message);
      });
  }

  if (typeof window.__GETIP_BLOCKLISTS_MOCK__ === 'object' && window.__GETIP_BLOCKLISTS_MOCK__ !== null) {
    render(window.__GETIP_BLOCKLISTS_MOCK__);
  } else {
    runFetch();
  }
})();
</script>
</body>
</html>
`
