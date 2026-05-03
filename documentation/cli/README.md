# `resolve` CLI

Look up **any** IPv4/IPv6 against your local GeoLite MMDBs. This is **not** exposed as an HTTP route on the get-ip server.

## Build

```bash
go build -o resolve ./cmd/resolve
```

## Usage

```bash
./resolve 190.98.108.105
```

Prints JSON with **`ip`**, **`geo`**, and **`asn`** (omitted when a database is missing or has no row). You need at least one `.mmdb` on disk (e.g. after the server has downloaded them, or copy the files in).

### Commands and flags

| Invocation | Behavior |
|------------|----------|
| `resolve <ip>` | JSON lookup for that IP. |
| `resolve fetch` | Download or refresh MMDBs (requires MaxMind env vars). |
| `resolve --fetch` | Same as `fetch`. |
| `resolve --fetch <ip>` | Update MMDBs, then JSON lookup. |

`--fetch` may be written as `-fetch`.

## Optional `.env` file

The **`resolve`** binary (and the **`get-ip`** server) load a **`.env`** file from the **current working directory** when the file exists. Lines use the usual `KEY=value` form. Variables already set in the real environment are **not** overridden (so CI and shell exports win). The repo’s **`.gitignore`** already ignores `.env`.

## Downloading or refreshing MMDBs

Same behavior as get-ip server startup: **`MAXMIND_ACCOUNT_ID`**, **`MAXMIND_LICENSE_KEY`**, **`GEOLITE_MAX_AGE_DAYS`**, and **`GEOLITE_CITY_PATH`** / **`GEOLITE_ASN_PATH`** (see the main [README](../../README.md#geolite2-optional)).

```bash
export MAXMIND_ACCOUNT_ID=…
export MAXMIND_LICENSE_KEY=…
./resolve fetch
```

Or put those keys in **`.env`** next to where you run the command and run **`./resolve fetch`** without exporting them in the shell.

Fetch then look up in one shot:

```bash
./resolve --fetch 190.98.108.105
```

## Docker image

The container image installs **`resolve`** at `/usr/local/bin/resolve` (same `GEOLITE_*_PATH` under `/data` when using the default image environment). Pass the same environment variables and volume mounts as the server when running **`resolve fetch`** or **`resolve <ip>`** inside the container.
