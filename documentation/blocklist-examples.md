# Blocklist feed examples (`BLOCKLIST_URLS`)

`get-ip` accepts **`BLOCKLIST_URLS`** as **semicolon-separated** HTTPS URLs. Each feed is one URL; optional **`|tag`** sets the label in **`/json`** → **`blocklists.matched`**. If you omit `|tag`, the label is derived from the filename (e.g. `firehol_level1` from `firehol_level1.netset`).

**Quoting:** The whole value usually needs **double quotes** in shell or Compose because of **`;`**.

**Limits:** Each URL download is capped at **64 MiB**. Very large feeds may need to be split or omitted.

**Semantics:** These lists are **routing / reputation style** data — use only in ways allowed by each publisher. Review **Spamhaus** fair-use / acceptable-use, **FireHOL** [CC BY-SA](https://github.com/firehol/blocklist-ipsets/blob/master/LICENSE) and list pages, **Emerging Threats**, **Abuse.ch**, etc.

**Formats:** The parser accepts **CIDR** (`203.0.113.0/24`) or **single IPs** (treated as `/32` or `/128`). Lines starting with **`#`** or **`;`** are ignored. Most FireHOL **`.netset`** and **`.ipset`** files in [firehol/blocklist-ipsets](https://github.com/firehol/blocklist-ipsets) match this.

---

## 1. Minimal — Spamhaus DROP + EDROP (official)

High-confidence “do not route” style prefixes. Use Spamhaus’s published HTTPS URLs and their terms.

```env
BLOCKLIST_URLS="https://www.spamhaus.org/drop/drop.txt|spamhaus-drop;https://www.spamhaus.org/drop/edrop.txt|spamhaus-edrop"
```

---

## 2. Recommended starter — FireHOL level 1 only

Single aggregated netset (often a good balance of coverage vs. noise). Raw URL pattern:

`https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/<filename>`

```env
BLOCKLIST_URLS="https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset|firehol-level1"
```

---

## 3. Broader — FireHOL level 1–3

Larger, more aggressive aggregation. Expect **more memory**, longer downloads, and **more false positives** if you treat “listed” as a hard block.

```env
BLOCKLIST_URLS="https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset|firehol-l1;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level2.netset|firehol-l2;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level3.netset|firehol-l3"
```

---

## 4. Spamhaus via FireHOL mirrors (optional)

If you prefer the same **Spamhaus DROP** content as a `.netset` in the FireHOL repo (still comply with Spamhaus + FireHOL terms):

```env
BLOCKLIST_URLS="https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/spamhaus_drop.netset|spamhaus-drop-fh"
```

You can combine **official** Spamhaus URLs (§1) **or** FireHOL mirrors — usually **not both**, to avoid duplicate work and conflicting refresh semantics.

---

## 5. Application-oriented FireHOL lists (high false-positive risk)

Use when you explicitly want “talks to lots of webservers” / “open proxies” style signals — **not** for a generic “is this IP dirty?” badge on normal users.

```env
BLOCKLIST_URLS="https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_webserver.netset|firehol-webserver;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_proxies.netset|firehol-proxies"
```

---

## 6. Threat / abuse extras (pick what matches your policy)

All below use the same **`raw.githubusercontent.com/.../master/<file>`** base. Examples that are often useful for **labeling** (not necessarily blocking):

| File (examples) | Typical use |
|-----------------|-------------|
| `feodo.ipset` | Feodo / banking trojan C2 IPs ([Abuse.ch](https://feodotracker.abuse.ch/)) |
| `dshield_1d.netset`, `dshield_7d.netset`, `dshield_30d.netset` | DShield aggressive nets (window size in name) |
| `et_compromised.ipset`, `et_tor.ipset`, `et_spamhaus.netset` | Emerging Threats–sourced sets in FireHOL |
| `abuseipdb_1_netset.ipset` | Very large; check size before enabling |

Example combining a **small** set of extras:

```env
BLOCKLIST_URLS="https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset|firehol-l1;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/feodo.ipset|feodo;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/dshield_7d.netset|dshield-7d"
```

Browse all available **`*.netset` / `*.ipset`** files in the repo root:  
[https://github.com/firehol/blocklist-ipsets/tree/master](https://github.com/firehol/blocklist-ipsets/tree/master)

---

## Docker Compose fragment

```yaml
services:
  get-ip:
    environment:
      BLOCKLIST_REFRESH: "24h"
      BLOCKLIST_URLS: "https://www.spamhaus.org/drop/drop.txt|spamhaus-drop;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset|firehol-level1"
```

On Windows PowerShell, when setting the variable for a one-off run:

```powershell
$env:BLOCKLIST_URLS = 'https://www.spamhaus.org/drop/drop.txt|spamhaus-drop;https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset|firehol-level1'
```

---

## Verifying

After startup, check logs for lines like `blocklist: <tag>: loaded N prefixes`. Then hit **`/json`** from a client IP; **`blocklists`** shows **`sources_loaded`**, **`listed`**, and **`matches`** when applicable.
