package main

import (
	"html"
	"net/http"
	"strings"
)

func writeRootHTML(w http.ResponseWriter, r *http.Request, ipStr string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Link", `</json>; rel="alternate"; type="application/json"`)
	w.Header().Add("Link", `</blocklists/json>; rel="alternate"; type="application/json"`)
	if r.Method == http.MethodHead {
		return
	}

	esc := html.EscapeString(ipStr)
	out := strings.ReplaceAll(rootHTMLTemplate, "__PRIMARY_IP__", esc)
	_, _ = w.Write([]byte(out))
}

// rootHTMLTemplate: responsive grid (details + map), auto-fetch /json. Placeholder __PRIMARY_IP__.
const rootHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Your IP</title>
<link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" crossorigin>
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
  }
}
* { box-sizing: border-box; }
html, body { height: 100%; }
body {
  margin: 0;
  min-height: 100%;
  font-family: ui-sans-serif, system-ui, -apple-system, Segoe UI, Roboto, Helvetica, Arial, sans-serif;
  background: radial-gradient(1200px 600px at 80% -10%, var(--accent-dim), transparent 55%),
    radial-gradient(900px 500px at -10% 50%, color-mix(in srgb, var(--muted) 15%, transparent), transparent 50%),
    var(--bg);
  color: var(--fg);
  display: flex;
  flex-direction: column;
  padding: 0 1rem 1.25rem;
}
.shell {
  flex: 1;
  width: 100%;
  max-width: 1080px;
  margin: 0 auto;
  padding-top: clamp(1rem, 3vw, 1.75rem);
}
.top {
  margin-bottom: 1.25rem;
}
.ip {
  margin: 0 0 0.35rem;
  font-size: clamp(1.75rem, 5vw, 2.35rem);
  font-weight: 700;
  letter-spacing: -0.02em;
  word-break: break-all;
  line-height: 1.15;
}
.hero-sub {
  margin: 0 0 1rem;
  font-size: 0.95rem;
  color: var(--muted);
  line-height: 1.45;
  max-width: 52ch;
}
.links {
  margin: 0;
  font-size: 0.875rem;
  color: var(--muted);
}
.links a {
  color: var(--accent);
  text-decoration: none;
  font-weight: 500;
}
.links a:hover { text-decoration: underline; }
.sep { margin: 0 0.4em; opacity: 0.7; }

.grid {
  display: grid;
  gap: 1.25rem;
  align-items: start;
}
@media (min-width: 860px) {
  .grid {
    grid-template-columns: minmax(0, 1fr) minmax(300px, 400px);
    gap: 1.5rem;
  }
  .map-aside {
    position: sticky;
    top: 1rem;
  }
}

.status {
  font-size: 0.9rem;
  color: var(--muted);
  padding: 0.5rem 0;
}
.status.err { color: #f5495f; }

.cards {
  display: grid;
  gap: 0.85rem;
}
@media (min-width: 520px) {
  .cards {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
@media (min-width: 860px) {
  .cards .card.span2 { grid-column: span 2; }
}

.card {
  background: var(--card);
  border: 1px solid var(--bd);
  border-radius: 14px;
  padding: 1rem 1.05rem;
  box-shadow: var(--shadow);
}
.card-h {
  margin: 0 0 0.65rem;
  font-size: 0.7rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.09em;
  color: var(--muted);
}
.dl > div { margin-bottom: 0.55rem; }
.dl > div:last-child { margin-bottom: 0; }
.dt {
  font-size: 0.65rem;
  text-transform: uppercase;
  letter-spacing: 0.06em;
  color: var(--muted);
  margin-bottom: 0.12rem;
}
.dd {
  margin: 0;
  font-size: 0.8125rem;
  line-height: 1.45;
  word-break: break-word;
}

.map-aside { width: 100%; }
.map-stage {
  height: min(42vh, 380px);
  min-height: 240px;
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--bd);
  box-shadow: var(--shadow);
  background: var(--bg2);
}
.map-placeholder {
  height: min(42vh, 380px);
  min-height: 240px;
  border-radius: 14px;
  border: 1px dashed var(--bd);
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 1rem;
  font-size: 0.85rem;
  color: var(--muted);
  background: color-mix(in srgb, var(--card) 60%, transparent);
}
.map-foot {
  margin: 0.5rem 0 0;
  font-size: 0.65rem;
  color: var(--muted);
  text-align: center;
  line-height: 1.4;
}

footer.site {
  margin-top: auto;
  padding-top: 2rem;
  text-align: center;
  font-size: 0.75rem;
  color: var(--muted);
  line-height: 1.5;
}
footer.site .mono { font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace; }

/* During fetch: single flat base + shimmer (avoids layered radial repaints; reduced-motion = static tint). */
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
  <header class="top">
    <p class="ip" id="hero-ip" translate="no">__PRIMARY_IP__</p>
    <p class="hero-sub" id="hero-sub" aria-live="polite"></p>
    <p class="links">
      <a href="/all">Plain report</a><span class="sep">·</span><a href="/json" target="_blank" rel="noopener">Raw JSON</a><span class="sep">·</span><a href="/blocklists">Blocklists</a>
    </p>
  </header>

  <div class="grid">
    <div class="stack">
      <div class="status" id="status">Loading details…</div>
      <div class="cards" id="cards" hidden></div>
    </div>
    <aside class="map-aside" aria-label="Map">
      <div id="map-mount"></div>
    </aside>
  </div>
</div>

<footer class="site">
  © Hosted on the swarm <span class="mono" aria-hidden="true">(⌐■_■)</span>.
</footer>

<script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js" crossorigin></script>
<script>
(function () {
  var statusEl = document.getElementById('status');
  var cardsEl = document.getElementById('cards');
  var heroIpEl = document.getElementById('hero-ip');
  var heroSubEl = document.getElementById('hero-sub');
  var mapMount = document.getElementById('map-mount');
  var map = null;

  function clearPageLoading() {
    document.body.classList.remove('page-loading');
  }

  function esc(s) {
    var d = document.createElement('div');
    d.textContent = s;
    return d.innerHTML;
  }

  function row(label, val) {
    if (val === undefined || val === null || val === '') return '';
    return '<div><div class="dt">' + esc(label) + '</div><p class="dd">' + esc(String(val)) + '</p></div>';
  }

  function card(title, inner) {
    if (!inner || !String(inner).trim()) return '';
    return '<article class="card"><h2 class="card-h">' + esc(title) + '</h2><div class="dl">' + inner + '</div></article>';
  }

  function cardSpan(title, inner, span2) {
    var c = card(title, inner);
    if (!c) return '';
    return span2 ? c.replace('<article class="card"', '<article class="card span2"') : c;
  }

  function listForwarded(f) {
    if (!f) return '';
    var parts = [];
    if (f.cf_connecting_ip) parts.push(row('CF-Connecting-IP', f.cf_connecting_ip));
    if (f.true_client_ip) parts.push(row('True-Client-IP', f.true_client_ip));
    if (f.x_real_ip) parts.push(row('X-Real-IP', f.x_real_ip));
    if (f.x_forwarded_for && f.x_forwarded_for.length)
      parts.push(row('X-Forwarded-For', f.x_forwarded_for.join(', ')));
    return parts.join('');
  }

  function geoInner(g) {
    if (!g) return '';
    return [
      row('City', g.city),
      row('Region', g.region_iso ? g.region + ' (' + g.region_iso + ')' : g.region),
      row('Country', g.country_name && g.country ? g.country_name + ' (' + g.country + ')' : (g.country || g.country_name)),
      row('Continent', g.continent_code ? g.continent + ' (' + g.continent_code + ')' : g.continent),
      row('Coordinates', g.loc),
      row('Timezone', g.timezone)
    ].join('');
  }

  function asnInner(a) {
    if (!a) return '';
    return [
      row('ASN', a.asn != null ? String(a.asn) : ''),
      row('Organization', a.organization),
      row('Network', a.network)
    ].join('');
  }

  function reqInner(q) {
    if (!q) return '';
    return [
      row('Method', q.method),
      row('Host', q.host),
      row('Protocol', q.protocol),
      row('User-Agent', q.user_agent)
    ].join('');
  }

  function addrInner(j) {
    return row('IPv4', j.ipv4) + row('IPv6', j.ipv6);
  }

  function parseLoc(loc) {
    if (!loc || typeof loc !== 'string') return null;
    var p = loc.split(',');
    if (p.length !== 2) return null;
    var lat = parseFloat(p[0]);
    var lng = parseFloat(p[1]);
    if (!isFinite(lat) || !isFinite(lng)) return null;
    return [lat, lng];
  }

  function heroLine(j) {
    var bits = [];
    if (j.geo) {
      if (j.geo.city) bits.push(j.geo.city);
      if (j.geo.country_name) bits.push(j.geo.country_name);
    }
    if (j.asn && j.asn.organization) bits.push(j.asn.organization);
    return bits.length ? bits.join(' · ') : '';
  }

  function setPrimaryIP(j) {
    var ip = j.ipv4 || j.ipv6 || (heroIpEl.textContent || '').trim();
    if (ip) heroIpEl.textContent = ip;
  }

  function renderMap(lat, lng) {
    mapMount.innerHTML = '';
    var stage = document.createElement('div');
    stage.className = 'map-stage';
    stage.id = 'map-stage';
    mapMount.appendChild(stage);

    var foot = document.createElement('p');
    foot.className = 'map-foot';
    foot.innerHTML = 'Tiles © <a href="https://www.openstreetmap.org/copyright" target="_blank" rel="noopener">OpenStreetMap</a> contributors';
    mapMount.appendChild(foot);

    if (typeof L === 'undefined') {
      stage.outerHTML = '<div class="map-placeholder">Map library failed to load.</div>';
      return;
    }

    map = L.map(stage).setView([lat, lng], 12);
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '',
      maxZoom: 19
    }).addTo(map);
    L.marker([lat, lng]).addTo(map);
    setTimeout(function () { map.invalidateSize(); }, 120);
    window.addEventListener('resize', function () {
      if (map) map.invalidateSize();
    });
  }

  function renderPlaceholderMap(msg) {
    mapMount.innerHTML = '<div class="map-placeholder">' + esc(msg) + '</div>';
  }

  function renderPage(j) {
    setPrimaryIP(j);
    heroSubEl.textContent = heroLine(j);

    var html = '';
    html += card('Addresses', addrInner(j));
    var fwd = listForwarded(j.forwarded);
    if (fwd) html += card('Forwarded', fwd);
    html += card('Location', geoInner(j.geo));
    html += card('Network', asnInner(j.asn));
    html += cardSpan('Request', reqInner(j.request), true);

    statusEl.hidden = true;
    statusEl.textContent = '';
    cardsEl.hidden = false;
    cardsEl.innerHTML = html || '<article class="card span2"><p class="dd">No detail fields.</p></article>';

    var coords = j.geo && parseLoc(j.geo.loc);
    if (coords) renderMap(coords[0], coords[1]);
    else renderPlaceholderMap('No GeoLite coordinates — map needs a loc field in JSON.');
    clearPageLoading();
  }

  function fail(msg) {
    clearPageLoading();
    statusEl.classList.add('err');
    statusEl.textContent = msg;
    renderPlaceholderMap('Load /json to see the map here.');
  }

  function runFetch() {
    fetch('/json', { credentials: 'same-origin', headers: { Accept: 'application/json' } })
      .then(function (r) {
        if (!r.ok) throw new Error('HTTP ' + r.status);
        return r.json();
      })
      .then(renderPage)
      .catch(function (e) {
        fail('Could not load /json: ' + e.message);
      });
  }

  if (typeof window.__GETIP_MOCK__ === 'object' && window.__GETIP_MOCK__ !== null) {
    statusEl.textContent = '';
    statusEl.hidden = true;
    renderPage(window.__GETIP_MOCK__);
  } else {
    runFetch();
  }
})();
</script>
</body>
</html>
`
