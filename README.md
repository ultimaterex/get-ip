# get-ip

Tiny HTTP service that echoes the caller’s IP: IPv4 when possible, otherwise IPv6. **`/`** serves a **small HTML page** when the client’s **`Accept`** header includes **`text/html`** (typical browsers); the page loads **live details** from **`/json`** (geo, ASN, forwarded headers, map). Non-browser clients get **plain text** (e.g. **`curl`**, scripts). **`/all`** is plain-text detail (optional estimated location + network); **`/json`** includes **`geo`** and **`asn`** when those MMDBs are loaded. Optional **HTTP prefix blocklists** and **DNSBL** live on **`/blocklists`** (HTML), **`/blocklists/json`**, and **`/blocklists/all`** — see [Blocklists](#blocklists). Summaries and forwarded headers use **public** addresses only.

## Run

```bash
go run .
# or
go build -o get-ip . && ./get-ip
```

`PORT` defaults to **8080** (e.g. `PORT=3000 go run .`).

**Health:** **`GET /health`** (and **`HEAD`**) return **200** with plain text **`ok`**. The Docker image’s **`HEALTHCHECK`** uses **`/health`** on port **8080**; if you change **`PORT`**, adjust your orchestrator’s probe to match.

**Discovery:** HTML responses include **`Link`** headers for **`</json>`** and **`</blocklists/json>`** (`rel="alternate"; type="application/json"`) where applicable ([RFC 8288](https://www.rfc-editor.org/rfc/rfc8288)).

**Client IP troubleshooting:** If **`/`** returns **503** “could not determine client IP”, no **public** address was found from **`CF-Connecting-IP`**, **`True-Client-IP`**, **`X-Real-IP`**, **`X-Forwarded-For`**, or the TCP **`RemoteAddr`**. Typical causes: the reverse proxy does not set forwarded headers; **`X-Forwarded-For`** only contains private or CDN edge addresses; or the visitor path is **CGNAT** (**`100.64.0.0/10`**, treated as non-public here). With **`GET_IP_DEBUG_HTTP=1`** (restart required), **`GET /debug`** and **`GET /debug/json`** return JSON with every IP candidate, parse errors, rejection reasons, all request headers (**`Authorization`** / **`Cookie`** redacted), optional TLS info, and the same forwarded summary as **`/json`**. Each **`GET`** (not **`HEAD`**) also writes the same snapshot to the process **log** (`debug: …` lines: request line, common **`Header.Get`** forwards, every candidate, diagnostics, TLS, then **every header** line). **Disable after troubleshooting** — the payload and logs are verbose and not for untrusted exposure.

If a **`.env`** file exists in the current working directory, it is loaded at startup (MaxMind and other variables). Values already set in the environment take precedence. **`.env`** is gitignored.

## Blocklist feeds (optional)

Point **`BLOCKLIST_URLS`** at one or more **HTTP(S) URLs** that serve plain-text **CIDR lists** (same general idea as Spamhaus DROP or FireHOL netsets: one network per line, **`#`** / **`;`** comments). Separate multiple URLs with **`;`**. Optionally suffix **`|tag`** for a stable label in **`/blocklists/json`** (e.g. `https://www.spamhaus.org/drop/drop.txt|spamhaus-drop`). Tags default from the URL filename if omitted.

**Compose / `.env` quoting:** Use either bare URLs or a single pair of quotes around the whole value — avoid YAML-style **`\"https://…\"`** inside the string (that embeds a backslash and breaks parsing). The service trims accidental quotes where possible; prefer `BLOCKLIST_URLS=https://a…;https://b…` or `BLOCKLIST_URLS="https://a…;https://b…"`.

On a schedule (**`BLOCKLIST_REFRESH`**, default **`24h`**), the service downloads each feed **on the server** (not in the visitor’s browser) into memory and tests the visitor’s **public IPv4 and IPv6** (when present) against those prefixes. The **`blocklists`** object ( **`listed`**, **`matched`**, per-feed **`matches`**, **`sources_loaded`**, **`last_refresh`**) is returned from **`/blocklists/json`** and the **`/blocklists`** HTML page, not from the default **`/json`** payload.

## Blocklists

| Route | Purpose |
|-------|---------|
| **`GET /blocklists`** | Browser-focused report (tables for prefix feeds + DNSBL). Plain **`curl`** without **`Accept: text/html`** prints pointers to JSON/plain endpoints. |
| **`GET /blocklists/json`** | Same core JSON shape as **`/json`**, plus **`blocklists`** and **`dnsbl`** when configured. |
| **`GET /blocklists/all`** | Plain-text blocklist report (prefix sections + DNSBL), plus forwarded/geo/asn summary. |

The home page (**`/`**) links to **`/blocklists`** so the default view stays lightweight. Old paths **`/spamlists`**, **`/spamlists/json`**, and **`/spamlists/all`** **301** to these routes.

**Licenses and acceptable use** belong to each feed’s publisher — comply with their terms and attribution (e.g. Spamhaus, FireHOL). Listings are **informational** (routing/blocklist membership), not a legal or abuse verdict.

**Copy-paste presets** (Spamhaus, FireHOL level 1–3, proxies, Feodo, DShield, Compose): **[documentation/blocklist-examples.md](documentation/blocklist-examples.md)**.

## DNS blocklists / DNSBL (optional)

For **MxToolbox-style** checks (live DNS queries against RBL zones), set **`DNSBL_ZONES`** to a semicolon-separated list of **zone suffixes** (what you append after the reversed IPv4). Optional **`zone|tag`** per entry for labels in JSON/HTML.

Example:

```env
DNSBL_ZONES="zen.spamhaus.org|spamhaus-zen;b.barracudacentral.org|barracuda;bl.spamcop.net|spamcop"
```

Lookups run **on the server** when you call **`/blocklists/json`**, **`/blocklists/all`**, or load **`/blocklists`** — the browser only receives results in JSON (or rendered HTML on **`/blocklists`**).

The service queries **`{reversed-ipv4}.{zone}`** for each zone and treats **`127.0.0.0/8`** A-record responses as “listed”. **IPv4 only** — if the visitor has no public IPv4, **`dnsbl.eligible`** is false (many real-world DNSBLs do not support IPv6 in this classic form).

**Operational note:** each publisher sets **rate limits and acceptable use**. Prefer short zone lists, sensible **`DNSBL_DEADLINE`** / **`DNSBL_PER_QUERY`**, and compliance with provider policies.

**Abuse controls:** results for a given **subject IPv4** are **cached** (see **`DNSBL_CACHE_TTL`**). **Fresh** DNS fan-out is also **rate-limited** per visitor and globally so one client cannot burn your resolver or third-party DNSBLs. **Cache hits** do not consume those limits.

**Copy-paste zone lists** (starter + longer options): **[documentation/dnsbl-examples.md](documentation/dnsbl-examples.md)**.

## CLI: `resolve`

The **`resolve`** tool queries local GeoLite MMDBs for any IP (not over HTTP). Full usage, **`fetch` / `--fetch`**, and environment variables are documented in **[documentation/cli/README.md](documentation/cli/README.md)**.

**Docker**

```bash
docker build -t get-ip .
docker run --rm -p 8080:8080 get-ip
```

**Compose** (host port from `HOST_PORT`, default 8080; **`.env.example`** → **`.env`** for optional MaxMind / GeoLite — see [GeoLite2](#geolite2-optional))

```bash
docker compose up -d --build
```

**Compose from GHCR** (image built in CI)

```bash
docker compose -f docker-compose.ghcr.yml pull
docker compose -f docker-compose.ghcr.yml up -d
```

## Examples

```bash
# one line: your IP (plain text; curl does not request text/html)
curl -s http://127.0.0.1:8080/
```

```text
203.0.113.7
```

Open **`/`** in a normal browser tab to see the lightweight HTML view (same IP, links to **`/all`**, **`/json`**, and **`/blocklists`**).

Offline mock preview (HTML file + **`python dev/gen_preview.py`**): **[documentation/browser-preview.md](documentation/browser-preview.md)**.

```bash
# details
curl -s http://127.0.0.1:8080/all
```

```text
IPv4: 203.0.113.7
IPv6: (none)

Forwarded headers (public addresses only)
  X-Forwarded-For: …
  …

Estimated location
  City: …
  Country: … (…)

Network
  ASN: …
  Organization: …
  Network: …

Blocklists (prefix feeds + DNSBL)
  GET /blocklists — HTML or pointers · GET /blocklists/json · GET /blocklists/all
  (legacy /spamlists → /blocklists)

Request
  Method: GET
  Host: 127.0.0.1:8080
  …
```

```bash
curl -s http://127.0.0.1:8080/json
```

```json
{
  "ipv4": "203.0.113.7",
  "ipv6": null,
  "forwarded": {
    "x_forwarded_for": ["203.0.113.7"]
  },
  "geo": {
    "city": "Montréal",
    "country": "CA",
    "loc": "45.5088,-73.5878",
    "timezone": "America/Toronto"
  },
  "asn": {
    "asn": 16276,
    "organization": "OVH SAS",
    "network": "149.56.0.0/16"
  },
  "request": {
    "method": "GET",
    "host": "127.0.0.1:8080",
    "protocol": "HTTP/1.1",
    "user_agent": "curl/8.x"
  }
}
```

The **`geo`** / **`asn`** objects (and the matching sections in **`/all`**) appear only when the corresponding MMDB is loaded and the lookup returns data. Optional **`blocklists`** / **`dnsbl`** are only in **`/blocklists/json`** (and the **`/blocklists`** page), not in the default **`/json`** example above.

## GeoLite2 (optional)

### 1. Get MaxMind credentials

1. Create a free **[MaxMind account](https://www.maxmind.com/en/geolite2/signup)** and accept the **GeoLite2 End User License Agreement**.
2. Note your **Account ID** ([how to find it](https://support.maxmind.com/knowledge-base/articles/find-your-maxmind-account-id)).
3. Under **[License keys](https://www.maxmind.com/en/accounts/current/license-key)**, generate a **license key** (used together with the Account ID for downloads).

Keep these secret — **do not commit them** to git (`.env` is gitignored; use your host or orchestrator’s secrets).

### 2. Define the license key (and account ID) for the process

The service reads **`MAXMIND_ACCOUNT_ID`** and **`MAXMIND_LICENSE_KEY`** from the environment.

**Shell (current session):**

```bash
export MAXMIND_ACCOUNT_ID="123456"
export MAXMIND_LICENSE_KEY="your_license_key_here"
./get-ip
```

**Docker:**

```bash
docker run --rm -p 8080:8080 \
  -e MAXMIND_ACCOUNT_ID="123456" \
  -e MAXMIND_LICENSE_KEY="your_license_key_here" \
  -v get-ip-data:/data \
  ghcr.io/ultimaterex/get-ip:latest
```

Mount **`/data`** so the downloaded MMDB survives container restarts (`GEOLITE_CITY_PATH` defaults to **`/data/GeoLite2-City.mmdb`** in the image).

**Compose** — `docker-compose.yml` (build) and `docker-compose.ghcr.yml` (pull from GHCR) already configure **`GEOLITE_CITY_PATH`** / **`GEOLITE_ASN_PATH`** under **`/data`**, a **`get-ip-data`** volume mounted at **`/data`**, and pass **`MAXMIND_ACCOUNT_ID`**, **`MAXMIND_LICENSE_KEY`**, **`GEOLITE_MAX_AGE_DAYS`**, and optional **`LOG_FILE`** via `${VAR:-…}` substitution.

Copy **`.env.example`** to **`.env`** next to the compose file (do not commit real keys). Compose loads **`.env`** automatically for that substitution.

Example `.env`:

```env
MAXMIND_ACCOUNT_ID=123456
MAXMIND_LICENSE_KEY=your_license_key_here
```

### 3. What happens on startup

1. If **both** `MAXMIND_ACCOUNT_ID` and `MAXMIND_LICENSE_KEY` are set **and** each MMDB file is **missing** or **older than** `GEOLITE_MAX_AGE_DAYS` (default **7**), the service **downloads** [GeoLite2-City](https://dev.maxmind.com/geoip/docs/databases/geolite2-city) and [GeoLite2-ASN](https://dev.maxmind.com/geoip/docs/databases/geolite2-asn) over HTTPS (**Basic auth**, following redirects per [MaxMind’s download docs](https://dev.maxmind.com/geoip/updating-databases/)).
2. It opens **`GEOLITE_CITY_PATH`** and **`GEOLITE_ASN_PATH`** (defaults: `data/GeoLite2-City.mmdb` and `data/GeoLite2-ASN.mmdb` locally; **`/data/...`** in Docker unless overridden).
3. **`/all`** may include **estimated location** and **network** sections; **`/json`** includes **`geo`** and **`asn`** — all use the resolved **public client IP** (IPv4 preferred, else IPv6).

If credentials are **not** set, nothing is downloaded automatically; the app still loads any MMDB files you placed at **`GEOLITE_CITY_PATH`** / **`GEOLITE_ASN_PATH`** yourself.

### Environment reference

| Variable | Meaning |
|----------|---------|
| `MAXMIND_ACCOUNT_ID` | MaxMind account ID (numeric string) |
| `MAXMIND_LICENSE_KEY` | License key from the MaxMind portal |
| `GEOLITE_CITY_PATH` | Path to the City MMDB (default `data/GeoLite2-City.mmdb`; Docker defaults to `/data/GeoLite2-City.mmdb`) |
| `GEOLITE_ASN_PATH` | Path to the ASN MMDB (default `data/GeoLite2-ASN.mmdb`; Docker defaults to `/data/GeoLite2-ASN.mmdb`) |
| `GEOLITE_MAX_AGE_DAYS` | Re-download if the file is older than this many days (default **7**) |
| `LOG_FILE` | If set, append the same log lines the process writes to stdout (Go **`log`** package) to this file. Stdout is unchanged, so **`docker logs`** still works. Example for Compose with the default **`/data`** volume: **`LOG_FILE=/data/get-ip.log`** in **`.env`**. Leave empty or unset to disable file logging. |
| `GET_IP_DEBUG_HTTP` | Set to **`1`**, **`true`**, **`yes`**, or **`on`** to register **`GET /debug`** and **`GET /debug/json`** (JSON client-IP diagnostics). **Off by default** — disable after troubleshooting. |
| `BLOCKLIST_URLS` | Semicolon-separated HTTP(S) URLs of CIDR blocklist feeds. Optional **`url|tag`** per [Blocklist feeds](#blocklist-feeds-optional). Empty disables the feature. |
| `BLOCKLIST_REFRESH` | How often to re-download feeds (Go **`time.ParseDuration`**, e.g. **`24h`**, **`6h`**). Default **`24h`**. Minimum effective refresh is **1m** (invalid values fall back to **`24h`**). |
| `DNSBL_ZONES` | Semicolon-separated DNSBL zone names (optional **`zone|tag`**). Empty disables live DNSBL lookups. See [DNS blocklists](#dns-blocklists--dnsbl-optional). |
| `DNSBL_PER_QUERY` | Per-zone DNS timeout (Go duration, default **`3s`**). |
| `DNSBL_DEADLINE` | Overall deadline for all zones in one request (default **`25s`**). |
| `DNSBL_CONCURRENCY` | Parallel DNS lookups (default **`12`**, max **256**). |
| `DNSBL_CACHE_TTL` | Cache DNSBL **`Info`** per subject IPv4 (Go duration). Default **`15m`**. Set **`0`** to disable caching. |
| `DNSBL_CLIENT_MAX` | Max **fresh** DNSBL runs per visitor key per **`DNSBL_CLIENT_WINDOW`** (default **`30`** per **`1h`**). Set **`0`** to disable this limit. **Cache hits do not count.** |
| `DNSBL_CLIENT_WINDOW` | Sliding window for **`DNSBL_CLIENT_MAX`** (Go duration, default **`1h`**). |
| `DNSBL_GLOBAL_MAX_PER_MINUTE` | Max **fresh** DNSBL runs per **UTC minute** for the whole process (default **`120`**). Set **`0`** to disable. **Cache hits do not count.** |
| `DNSBL_RL_MAX_CLIENT_KEYS` | When the per-client map grows past this many keys, half the entries are dropped (default **20000**). |

**Attribution:** GeoLite2 is © MaxMind; use requires [GeoLite2 attribution](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) wherever you display this data.

Behind a reverse proxy, set `X-Forwarded-For` / `X-Real-IP` (or your provider’s equivalent) so the app sees the real client.
