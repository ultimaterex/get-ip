# IPv4 and IPv6 behind Docker and Caddy

**get-ip** is a **self-hosted, self-serve** “what is my IP” service—visitors hit your deployment to learn their public addresses; you operate the stack (reverse proxy, Docker networks, DNS). This guide explains how to run **get-ip** behind **Caddy** so browsers and **`curl`** see **correct client IPv4 or IPv6**. You only need **one** shared Docker network—**`proxy-all`**—with **IPv6 enabled** so Caddy and get-ip share a dual-stack bridge. **`reverse_proxy`** should target the **container name** on that network.

Use placeholders like **`ip.example.com`** wherever a hostname appears.

**Linux hosts:** enable IPv6 in the **Docker Engine** before user-defined networks can use IPv6 reliably—see [Enable IPv6 on Docker Engine (Linux)](#enable-ipv6-on-docker-engine-linux).

---

## Why `proxy-all` with `external: true`

```yaml
networks:
  proxy-all:
    external: true
```

**What it means:** Compose **does not create** this network. Docker must already have a network named **`proxy-all`** (from another compose project or `docker network create`).

**Why use it:**

1. **One shared LAN for proxies and apps.** Caddy and get-ip often live in **different** compose files or repos. Marking **`proxy-all`** as **`external: true`** lets **every** stack **join** the same bridge by name. Containers on that network can resolve **`http://get-ip:8080`** by DNS name.

2. **Avoid duplicate definitions.** If two projects each defined **`proxy-all`** without **`external`**, Compose would create **separate** networks (sometimes with name prefixes), and **Caddy would not** reliably reach **`get-ip`** by service name.

3. **Central IPv6 settings.** Create **`proxy-all` once** with **`enable_ipv6: true`** and fixed subnets so it is an **IPv6-capable** Docker network. Caddy and get-ip attach **only** to **`proxy-all`** for proxy traffic—no extra **`proxy`** / **`default`** networks are required for this pattern.

**Typical split:** An “infra” or root compose file **defines** `proxy-all`. Application compose files **reference** `proxy-all` as **`external: true`**.

---

## Why Caddy uses `reverse_proxy http://get-ip:8080`

```caddyfile
https://ip.example.com {
	encode gzip
	reverse_proxy http://get-ip:8080
}
http://ip.example.com {
	encode gzip
	reverse_proxy http://get-ip:8080
}
```

**Replace `ip.example.com` with your real hostname.**

### Same Docker network as get-ip

**`get-ip`** must be the **service/container name** resolvable on **`proxy-all`**—the **only** network both services need for reverse-proxy traffic. Then:

- Traffic goes **Caddy → get-ip** on that bridge, not through **`127.0.0.1`** on the host or **`host.docker.internal`** unless you deliberately publish ports that way.
- **IPv6 on `proxy-all`** (`enable_ipv6: true`) lets container IPv6 addressing work on that bridge; combined with **Docker Engine IPv6 on Linux**, you avoid misleading **`RemoteAddr`** / forwarded chains when debugging IPv6 paths.

### Correct client IP for get-ip

get-ip reads **`CF-Connecting-IP`**, **`True-Client-IP`**, **`X-Real-IP`**, **`X-Forwarded-For`**, then **`RemoteAddr`**. Caddy terminates the visitor’s TCP connection (IPv4 or IPv6) and forwards to get-ip over HTTP, setting forwarded headers so get-ip can recover the **original client** address. That only behaves well if Caddy is the **direct** HTTP peer on a stable Docker network—not if you bypass it with brittle host-port hairpins.

### HTTP and HTTPS blocks

- **`https://…`** — TLS on 443 (adjust TLS/automation to your setup).
- **`http://…`** — plain HTTP on 80 (redirect to HTTPS in production if you want; the example shows both for clarity).

### `encode gzip`

Compresses responses (HTML/JSON); optional but cheap for text-heavy endpoints.

---

## Compose: only `proxy-all`

Attach **Caddy** and **get-ip** to **`proxy-all`** only. That network is your **IPv6-capable** Docker bridge for traffic between the reverse proxy and the app.

```yaml
services:
  get-ip:
    # ... image, env, volumes ...
    networks:
      - proxy-all

  # caddy:
  #   networks:
  #     - proxy-all

networks:
  proxy-all:
    name: proxy-all
    enable_ipv6: true
    ipam:
      config:
        - subnet: 172.30.0.0/16       # IPv4 range for this Docker network
        - subnet: fd00:bad:c0de::/64  # IPv6 ULA range for this Docker network (pick your own ULA)
```

If **`proxy-all`** is created elsewhere, reference it without redefining **`ipam`**:

```yaml
services:
  get-ip:
    networks:
      - proxy-all

networks:
  proxy-all:
    external: true
```

### Why `enable_ipv6: true` and explicit subnets

- **`enable_ipv6: true`** — Docker assigns **IPv6** addresses to containers on **`proxy-all`**. Without IPv6 on that network, IPv6-only paths from the edge proxy into the app stack can break or show misleading **`RemoteAddr`** / forwarded chains when debugging.

- **Fixed IPv4 / IPv6 subnets** — Avoid overlaps with other Docker networks or your LAN; makes **`iptables`/nft** and troubleshooting predictable.

- **`fd00:bad:c0de::/64`** — This is a **ULA** range for **container addressing on the Docker bridge**. It is **not** your visitors’ global IPv6. Visitors still reach Caddy with **global** IPv6; Caddy forwards **`X-Forwarded-For`** (and related headers) so get-ip sees the **public** address.

**Important:** If **`proxy-all`** is declared **`external: true`** in **this** file, **do not** also define **`subnet` / `enable_ipv6`** here—those belong in the **one** place that **creates** the network. Use either:

- **Define** `proxy-all` (with `enable_ipv6` and `ipam`) in a single infra compose, **or**
- **`docker network create`** with IPv6, then **`external: true`** everywhere else.

---

## Enable IPv6 on Docker Engine (Linux)

User-defined networks with **`enable_ipv6: true`** need the Engine to support IPv6. On **Linux**, turn IPv6 on in the daemon, then restart Docker. Official overview: [Enable IPv6 support](https://docs.docker.com/config/daemon/ipv6/).

### 1. Host already has IPv6

Confirm the machine has a global or link IPv6 stack (`ip -6 route`, `ip -6 addr`). If the host has no IPv6, fix networking first; Docker cannot invent upstream IPv6 for your WAN.

### 2. Configure `/etc/docker/daemon.json`

Create or merge JSON so IPv6 is enabled. Use a **ULA** (or another non-colliding prefix) for Docker’s **default bridge** IPv6 pool—pick a **`fd00::/64`** (or larger) block that does not overlap your **`proxy-all`** **`ipam`** subnet.

```json
{
  "ipv6": true,
  "fixed-cidr-v6": "fd00:dead:beef::/64",
  "ip6tables": true
}
```

Notes:

- **`ipv6` / `fixed-cidr-v6`** — On many setups the Engine needs this so it allocates IPv6 on the **default** bridge; without it, **user-defined** IPv6 networks can fail or behave inconsistently depending on version.
- **`ip6tables: true`** — Lets Docker manage IPv6 NAT/filter rules where supported; omit if your package documents a different default.
- If **`daemon.json`** already exists, merge these keys—do not duplicate the outer **`{}`**.

Validate JSON, then:

```bash
sudo systemctl restart docker
```

### 3. Verify

```bash
docker network ls
docker network inspect proxy-all
```

Look for an IPv6 **`Subnet`** and **`EnableIPv6": true`**. Run a short-lived container on **`proxy-all`** and check **`GlobalIPv6Address`**.

### 4. Packet forwarding (optional)

If containers must forward IPv6 through the host (less common for a simple reverse-proxy hop), ensure forwarding is on:

```bash
sudo sysctl -w net.ipv6.conf.all.forwarding=1
```

Persist in **`/etc/sysctl.d/`** if needed. Your firewall (**nftables**, **firewalld**) may need rules for forwarded IPv6—consult your distro.

### 5. Rootless Docker

Rootless mode may restrict IPv6 bridging; if **`enable_ipv6`** networks fail after daemon changes, see [Docker Engine: Rootless mode](https://docs.docker.com/engine/security/rootless/) for your version.

---

## DNS and your main `ip` hostname

Point **`ip.example.com`** (and **`www`**, etc.) **A** and **AAAA** records at the machine or load balancer that runs **Caddy**, not at get-ip’s unpublished container IP unless that is intentional.

---

## Cloudflare: `ipv4.ip` and `ipv6.ip` split hostnames

These names mean **two DNS names under your zone**, e.g. **`ipv4.ip.example.com`** and **`ipv6.ip.example.com`** (three labels before **`example.com`**: **`ipv4`** / **`ipv6`**, then **`ip`**). The goal is to force **two different client connection families** so a single page can **`fetch()` both** and merge **`ipv4`** and **`ipv6`** from **`/json`** (same JSON shape as **`GET /json`**).

### What to create in Cloudflare DNS

| Name | Records | Purpose |
|------|---------|---------|
| **`ipv4.ip`** | **A** only → your origin **IPv4** (or proxied target). **No AAAA** on this name. | Resolvers and browsers that pick this name can only use **IPv4** toward whatever **A** returns (your edge or Cloudflare). |
| **`ipv6.ip`** | **AAAA** only → your origin **IPv6**. **No A** on this name. | Same stack only over **IPv6**. |

**Rules:**

1. **Exactly one address family per hostname.** If **`ipv4.ip`** gets an **AAAA** record, Happy Eyeballs may use IPv6 and you lose the “IPv4-only name” guarantee.
2. **Same origin service.** Both names terminate on **Caddy** on your host (orange cloud) or directly on your host (grey cloud). get-ip stays **`reverse_proxy`** behind Caddy; no separate app instance per name.
3. **Proxied (orange) vs DNS-only (grey)**  
   - **DNS-only:** **`A`/`AAAA`** point at **your** server’s global IPv4 / IPv6. Visitor ↔ origin path matches those records; **`X-Forwarded-For`** / **`RemoteAddr`** behave like a normal reverse proxy (no **`CF-Connecting-IP`** unless you add something else).  
   - **Proxied:** Cloudflare sits in front; client IPs visible to get-ip often arrive via **`CF-Connecting-IP`** / **`X-Forwarded-For`** (get-ip already prefers **`CF-Connecting-IP`** when present). TLS may terminate at Cloudflare; origin can use HTTP or HTTPS depending on your SSL mode.

4. **TLS on Caddy:** Issue certificates for **every** hostname you serve (`ip`, **`ipv4.ip`**, **`ipv6.ip`**, etc.). Caddy’s ACME will request names listed in your **`Caddyfile`**; ensure **DNS** and **firewall** allow challenges for each name.

### Frontend usage (reference)

From your main **`https://ip.example.com`** page (or any origin):

```javascript
const [v4, v6] = await Promise.all([
  fetch("https://ipv4.ip.example.com/json").then((r) => r.json()),
  fetch("https://ipv6.ip.example.com/json").then((r) => r.json()),
]);
```

Replace **`example.com`** with your apex. **`ipv4.ip`** and **`ipv6.ip`** are **DNS names** (labels under **`ip.<apex>`**), not URL paths on **`ip.example.com`**.

### Built-in **`/`** HTML: env-driven dual fetch

If you set **both** of these at runtime, the browser **home** page (**`GET /`** with **`Accept: text/html`**) loads JSON from **two absolute URLs** in parallel and merges **`ipv4`** / **`ipv6`** (plus **`forwarded`**, **`geo`**, **`asn`**, **`request`** from the first successful payload). If **both** fetches fail, it falls back to **`GET /json`** on the **same** origin.

The **`/blocklists`** HTML page uses the **same env vars**: it derives **`https://ipv4.ip.example.com/blocklists/json`** and **`https://ipv6.ip.example.com/blocklists/json`** from those **`…/json`** URLs (same scheme and host, path **`/blocklists/json`**), merges **`ipv4`** / **`ipv6`**, unions HTTP prefix **blocklist** hits, and prefers an **`eligible`** DNSBL payload when one response has public IPv4. If both cross-origin fetches fail, it falls back to **`GET /blocklists/json`** on the same origin.

Opening **`/json`** or **`/blocklists/json`** **directly** in a browser tab is still a **single** HTTP response (no merge)—only the **`/`** and **`/blocklists`** HTML views run client-side dual fetch.

| Variable | Purpose |
|----------|---------|
| **`GET_IP_DUAL_FETCH_IPV4_JSON_URL`** | Full URL, e.g. **`https://ipv4.ip.example.com/json`** (or **`…/ipv4/json`** on the same host). |
| **`GET_IP_DUAL_FETCH_IPV6_JSON_URL`** | Full URL, e.g. **`https://ipv6.ip.example.com/json`** (or **`…/ipv6/json`**). |

Cross-origin **`fetch()`** from **`https://ip.example.com`** to those hosts requires CORS on JSON responses. Set:

| Variable | Purpose |
|----------|---------|
| **`GET_IP_ACCESS_CONTROL_ALLOW_ORIGIN`** | Single origin allowed to read JSON cross-origin, e.g. **`https://ip.example.com`**. get-ip adds **`Access-Control-Allow-Origin`** to **`/json`**, **`/ipv4/json`**, **`/ipv6/json`**, and **`/blocklists/json`**, and answers **`OPTIONS`** preflight when this is set. |

Dedicated routes (same app): **`GET /ipv4`** / **`GET /ipv6`** return **plain text** for that family only. If there is no public address for that family, the response is **503** with plain-text body **`IPv4_NOT_AVAILABLE`** or **`IPv6_NOT_AVAILABLE`**. **`GET /ipv4/json`** and **`GET /ipv6/json`** return the **same JSON** as **`GET /json`** for the current request (useful aliases when each hostname only exposes one path).

---

## Caddy: add sites for `ipv4.ip` and `ipv6.ip`

Your existing **`ip.example.com`** block stays as-is. **Add** equivalent **`https://`** / **`http://`** blocks for the two split names, each proxying to the **same** **`get-ip`** upstream.

### Option A — Repeat blocks (explicit)

```caddyfile
https://ipv4.ip.example.com {
	encode gzip
	reverse_proxy http://get-ip:8080
}
http://ipv4.ip.example.com {
	encode gzip
	reverse_proxy http://get-ip:8080
}

https://ipv6.ip.example.com {
	encode gzip
	reverse_proxy http://get-ip:8080
}
http://ipv6.ip.example.com {
	encode gzip
	reverse_proxy http://get-ip:8080
}
```

Replace **`example.com`** with your zone apex.

### Option B — Snippet (DRY)

```caddyfile
(get_ip_upstream) {
	encode gzip
	reverse_proxy http://get-ip:8080
}

https://ipv4.ip.example.com {
	import get_ip_upstream
}
http://ipv4.ip.example.com {
	import get_ip_upstream
}

https://ipv6.ip.example.com {
	import get_ip_upstream
}
http://ipv6.ip.example.com {
	import get_ip_upstream
}
```

**You must:**

1. **Reload Caddy** after editing the **`Caddyfile`** (`caddy reload` or container restart).
2. **Let ACME run** for **`ipv4.ip.*`** and **`ipv6.ip.*`** (first request or Caddy startup may obtain certs).
3. Keep **`get-ip`** on the **same Docker network** as Caddy so **`http://get-ip:8080`** still resolves.

No change is required **inside** get-ip for these names; it already trusts **`CF-Connecting-IP`** when Cloudflare is proxied.

---

## Quick checklist

1. **Linux:** Docker Engine **`daemon.json`** has **`ipv6`** (and **`fixed-cidr-v6`** as needed); **`systemctl restart docker`** completed.
2. **`proxy-all`** exists as an **IPv6-capable** Docker network (**`enable_ipv6: true`** + subnets where you define it).
3. **Caddy** and **get-ip** attach **only** to **`proxy-all`** for proxy connectivity.
4. Caddy **`reverse_proxy http://get-ip:8080`** for **`ip`**, **`ipv4.ip`**, and **`ipv6.ip`** hostnames.
5. Public DNS: **`ipv4.ip`** = **A only**; **`ipv6.ip`** = **AAAA only**; main **`ip`** can keep **A + AAAA** as you prefer.

After this, **`curl -4 https://ipv4.ip.example.com/json`** should emphasize IPv4 path visibility; **`curl -6 https://ipv6.ip.example.com/json`** the IPv6 path. **`curl -4`** / **`curl -6`** against **`https://ip.example.com`** still depends on resolver choice on a dual-stack name.
