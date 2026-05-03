# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src

COPY go.mod ./
COPY main.go ./

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /get-ip .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates

COPY --from=build /get-ip /usr/local/bin/get-ip

USER nobody
EXPOSE 8080
ENV PORT=8080

# Request logs go to stdout — use `docker logs` / your orchestrator’s log tail.
ENTRYPOINT ["/usr/local/bin/get-ip"]
