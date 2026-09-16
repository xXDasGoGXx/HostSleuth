# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24-alpine3.23 AS build
WORKDIR /src

COPY go.mod ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal

ARG TARGETOS
ARG TARGETARCH
ARG VERSION=0.1.0-dev
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" \
    -o /out/hostsleuth ./cmd/hostsleuth

FROM alpine:3.24
RUN apk add --no-cache ca-certificates docker-cli iproute2 procps-ng

COPY --from=build /out/hostsleuth /usr/local/bin/hostsleuth

ENV HOSTSLEUTH_MODE=docker \
    HOSTSLEUTH_STATE_DIR=/var/lib/hostsleuth \
    HOSTSLEUTH_OS_RELEASE_PATH=/host/etc/os-release

ENTRYPOINT ["/usr/local/bin/hostsleuth"]
CMD ["serve", "--state-dir", "/var/lib/hostsleuth", "--listen", "127.0.0.1:8787", "--interval", "60s"]
