FROM golang:latest AS builder

LABEL org.opencontainers.image.source https://github.com/telenornms/skogul

RUN mkdir -p src

WORKDIR src

COPY go.mod go.sum ./

COPY . .

RUN make skogul

# Runtime
FROM debian:stable-slim

RUN apt update && apt install -y \
	ca-certificates

COPY --from=builder /go/src/skogul /usr/local/bin/skogul
COPY --from=builder /go/src/docs/examples/basics/default.json /etc/skogul/conf.d/default.json

ENTRYPOINT ["skogul", "-loglevel=i"]
