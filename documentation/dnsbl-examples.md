# DNSBL zone examples (`DNSBL_ZONES`)

These are **DNS zone suffixes** for classic IPv4 DNSBL queries: the server looks up `**{reversed-ipv4}.{zone}`** and treats `**127.0.0.0/8**` answers as “listed”.

This is **not** the same as `[BLOCKLIST_URLS](./blocklist-examples.md)` (HTTP CIDR downloads). You can use **either or both**.

**Requirements**

- **Public IPv4** on the visitor — otherwise `dnsbl.eligible` is false in `/json`.
- **Publisher rules** — rate limits, attribution, and automation policies apply to **your** resolver traffic. Verify each zone’s current documentation before production use; zone names **change** over time.

**Environment**


| Variable                      | Default   | Meaning                                                                             |
| ----------------------------- | --------- | ----------------------------------------------------------------------------------- |
| `DNSBL_ZONES`                 | *(empty)* | `;`-separated zones, optional `zone|tag`                                            |
| `DNSBL_PER_QUERY`             | `3s`      | Timeout per zone                                                                    |
| `DNSBL_DEADLINE`              | `25s`     | Cap for one `/json` request                                                         |
| `DNSBL_CONCURRENCY`           | `12`      | Parallel lookups                                                                    |
| `DNSBL_CACHE_TTL`             | `15m`     | Per–subject-IPv4 result cache (set `0` to disable)                                  |
| `DNSBL_CLIENT_MAX`            | `30`      | Max **fresh** DNSBL runs per visitor per `DNSBL_CLIENT_WINDOW` (set `0` to disable) |
| `DNSBL_CLIENT_WINDOW`         | `1h`      | Sliding window for the per-client cap                                               |
| `DNSBL_GLOBAL_MAX_PER_MINUTE` | `120`     | Max **fresh** runs per UTC **minute** for the whole process (set `0` to disable)    |
| `DNSBL_RL_MAX_CLIENT_KEYS`    | `20000`   | Truncate the per-client map if it grows past this size                              |


**Cache vs rate limits:** a **cache hit** (same **subject IPv4** within `**DNSBL_CACHE_TTL`**) returns the previous result and **does not** run DNS or count against the per-client / global limits. Only **cache misses** that perform a full zone fan-out are limited.

---

## Starter (small, widely referenced)

Good for smoke-testing `**dnsbl`** without hammering dozens of zones:

```env
DNSBL_ZONES="zen.spamhaus.org|spamhaus-zen;b.barracudacentral.org|barracuda;bl.spamcop.net|spamcop;psbl.surriel.com|psbl"
```

---

## Broader set (names roughly aligned with common checker UIs)

The left column is a **human label** (like product names in reputation tools). The right column is the **zone string** to configure. **Confirm each zone with the operator’s docs** — this list is for orientation, not a legal guarantee.


| Label (informative)       | Typical zone suffix                                                          |
| ------------------------- | ---------------------------------------------------------------------------- |
| Spamhaus ZEN              | `zen.spamhaus.org`                                                           |
| Barracuda                 | `b.barracudacentral.org`                                                     |
| SpamCop                   | `bl.spamcop.net`                                                             |
| PSBL                      | `psbl.surriel.com`                                                           |
| Backscatterer             | `ips.backscatterer.org`                                                      |
| UCEPROTECT L1 / L2 / L3   | `dnsbl-1.uceprotect.net`, `dnsbl-2.uceprotect.net`, `dnsbl-3.uceprotect.net` |
| DRONE BL                  | `dnsbl.dronebl.org`                                                          |
| MAILSPIKE BL              | `bl.mailspike.net`                                                           |
| MAILSPIKE Z               | `z.mailspike.net`                                                            |
| SORBS aggregate (example) | `dnsbl.sorbs.net`                                                            |
| Spamhaus DBL is domain —  | *not applicable to raw IP DNSBL in this form*                                |


Example **quoted** value (line breaks for readability — use a single line in `.env`):

```env
DNSBL_ZONES="zen.spamhaus.org|spamhaus-zen;b.barracudacentral.org|barracuda;bl.spamcop.net|spamcop;psbl.surriel.com|psbl;ips.backscatterer.org|backscatterer;dnsbl-1.uceprotect.net|uce-l1;dnsbl-2.uceprotect.net|uce-l2;dnsbl-3.uceprotect.net|uce-l3;dnsbl.dronebl.org|dronebl;bl.mailspike.net|mailspike-bl;z.mailspike.net|mailspike-z"
```

Tune `**DNSBL_DEADLINE**` upward if you configure many zones (e.g. `**45s**`) or reduce the number of zones.

---

## Largest practical IPv4 preset (public zones only)

This is a **large** set of **independently operated** DNSBL zones suitable for classic `**{reversed-ipv4}.{zone}`** queries. It is meant for **coverage**, not minimal redundancy: several Spamhaus-related zones overlap conceptually (you may drop `**zen`** if you keep `**sbl`/`xbl`/`pbl**`, or keep `**zen**` only to save queries).

**Not included on purpose**


| Item                             | Why                                                                                       |
| -------------------------------- | ----------------------------------------------------------------------------------------- |
| **SORBS** (`dnsbl.sorbs.net`, …) | Service **shut down** (mid‑2024); zones are dead weight.                                  |
| **WPBL** (`db.wpbl.info`)        | Listed as **shutdown**; lookups may hang or error.                                        |
| **SpamRATS** (`*.spamrats.com`)  | Often uses **per‑subscriber / keyed** zone labels — no single public suffix for everyone. |
| **Invaluement ivmSIP**           | **Paid / assigned hostname** — not a fixed public zone string.                            |


**Recommended tuning** when using ~40+ zones:

```env
DNSBL_DEADLINE=120s
DNSBL_PER_QUERY=4s
DNSBL_CONCURRENCY=24
```

**Max preset (`DNSBL_ZONES` — single line):**

```env
DNSBL_ZONES="zen.spamhaus.org;sbl.spamhaus.org;xbl.spamhaus.org;pbl.spamhaus.org;cbl.abuseat.org;b.barracudacentral.org;bl.spamcop.net;psbl.surriel.com;ips.backscatterer.org;dnsbl-1.uceprotect.net;dnsbl-2.uceprotect.net;dnsbl-3.uceprotect.net;dnsbl.dronebl.org;bl.mailspike.net;z.mailspike.net;rep.mailspike.net;ix.dnsbl.manitu.net;dnsbl.nixspam.de;bl.blocklist.de;tor.dan.me.uk;torexit.dan.me.uk;phishing.rbl.msrbl.net;spam.rbl.msrbl.net;virus.rbl.msrbl.net;images.rbl.msrbl.net;combined.rbl.msrbl.net;dnsbl.spfbl.net;dnsbl.zapbl.net;all.s5h.net;rbl.interserver.net;rblspamassassin.interserver.net;rbl.efnetrbl.org;bl.suomispam.net;gl.suomispam.net;truncate.gbudb.net;ubl.unsubscore.com"
```

Expect **some errors/timeouts** on individual zones (maintenance, rate limits, or resolver policies). That is normal — inspect `**checks[].error`** in `**/json**`.

---

## Docker Compose

```yaml
environment:
  DNSBL_ZONES: "zen.spamhaus.org|spamhaus-zen;bl.spamcop.net|spamcop"
  DNSBL_DEADLINE: "30s"
```

---

## JSON shape (`/json`)

When `DNSBL_ZONES` is non-empty, responses include `**dnsbl**`:

- `**eligible**` — whether a public IPv4 was checked.
- `**listed**` — true if **any** zone returned a `127.0.0.0/8` answer.
- `**checks`** — per-zone `**listed**`, `**return_codes**`, `**response_ms**`, `**error**` (if the lookup failed).

