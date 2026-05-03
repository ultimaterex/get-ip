# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
COPY *.go ./
COPY internal/ internal/
COPY cmd/ cmd/

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /get-ip . && \
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /resolve ./cmd/resolve

FROM alpine:3.20
RUN apk add --no-cache ca-certificates \
	&& mkdir -p /data \
	&& chown 65534:65534 /data

COPY --from=build /get-ip /usr/local/bin/get-ip
COPY --from=build /resolve /usr/local/bin/resolve

USER nobody
EXPOSE 8080
ENV PORT=8080
ENV GEOLITE_CITY_PATH=/data/GeoLite2-City.mmdb
ENV GEOLITE_ASN_PATH=/data/GeoLite2-ASN.mmdb

# Request logs go to stdout — use `docker logs` / your orchestrator’s log tail.
ENTRYPOINT ["/usr/local/bin/get-ip"]
