# get-ip

Tiny HTTP service that echoes the caller’s IP: IPv4 when possible, otherwise IPv6. **`/`** serves a **small HTML page** when the client’s **`Accept`** header includes **`text/html`** (typical browsers); otherwise it returns **plain text** (e.g. **`curl`**, scripts). **`/all`** is plain-text detail (optional estimated location + network sections); **`/json`** includes **`geo`** and **`asn`** when those MMDBs are loaded. Summaries and forwarded headers use **public** addresses only.

## Run

```bash
go run .
# or
go build -o get-ip . && ./get-ip
```

`PORT` defaults to **8080** (e.g. `PORT=3000 go run .`).

If a **`.env`** file exists in the current working directory, it is loaded at startup (MaxMind and other variables). Values already set in the environment take precedence. **`.env`** is gitignored.

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

Open **`/`** in a normal browser tab to see the lightweight HTML view (same IP, links to **`/all`** and **`/json`**).

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

The **`geo`** / **`asn`** objects (and the matching sections in `/all`) appear only when the corresponding MMDB is loaded and the lookup returns data.

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

**Attribution:** GeoLite2 is © MaxMind; use requires [GeoLite2 attribution](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data) wherever you display this data.

Behind a reverse proxy, set `X-Forwarded-For` / `X-Real-IP` (or your provider’s equivalent) so the app sees the real client.
