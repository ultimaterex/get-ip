# get-ip

Tiny HTTP service that echoes the caller’s IP: IPv4 when possible, otherwise IPv6. **`/all`** dumps IPv4/IPv6 plus connection and common forwarded headers.

## Run

```bash
go run .
# or
go build -o get-ip . && ./get-ip
```

`PORT` defaults to **8080** (e.g. `PORT=3000 go run .`).

**Docker**

```bash
docker build -t get-ip .
docker run --rm -p 8080:8080 get-ip
```

**Compose** (host port from `HOST_PORT`, default 8080)

```bash
docker compose up -d --build
```

## Examples

```bash
# one line: your IP
curl -s http://127.0.0.1:8080/
```

```text
203.0.113.7
```

```bash
# details
curl -s http://127.0.0.1:8080/all
```

```text
IPv4: 203.0.113.7
IPv6: (none)

Direct connection
  RemoteAddr: 203.0.113.7:41290
  Parsed IP: 203.0.113.7
  Port: 41290

  X-Forwarded-For: …
  …

Request
  Method: GET
  Host: 127.0.0.1:8080
  …
```

Behind a reverse proxy, set `X-Forwarded-For` / `X-Real-IP` (or your provider’s equivalent) so the app sees the real client.
