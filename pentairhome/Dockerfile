ARG BUILD_FROM=ghcr.io/home-assistant/base:3.23

FROM golang:1.26.4 as builder

WORKDIR /usr/src/app
COPY src/. ./

RUN go mod download && go mod verify && go build -o /usr/bin/pentairhome main.go

# hadolint ignore=DL3006
FROM ${BUILD_FROM}

# Set shell
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

RUN apk add --no-cache libc6-compat=1.1.0-r4

COPY rootfs /
COPY --from=builder /usr/bin/pentairhome /usr/bin/pentairhome
RUN chmod +x /usr/bin/pentairhome

LABEL \
    org.opencontainers.image.title="Home Assistant Pentair Home" \
    org.opencontainers.image.description="This hardly works and is not affiliated with Pentair." \
    org.opencontainers.image.source="https://github.com/ThomasLomas/ha-pentairhome-addon" \
    org.opencontainers.image.licenses="Apache License 2.0"
