ARG BUILD_FROM=ghcr.io/hassio-addons/base:15.0.6

FROM golang:1.22 as builder

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

# Build arguments
ARG BUILD_ARCH
ARG BUILD_DATE
ARG BUILD_DESCRIPTION
ARG BUILD_NAME
ARG BUILD_REF
ARG BUILD_REPOSITORY
ARG BUILD_VERSION

# Labels
LABEL \
    io.hass.name="${BUILD_NAME}" \
    io.hass.description="${BUILD_DESCRIPTION}" \
    io.hass.arch="${BUILD_ARCH}" \
    io.hass.type="addon" \
    io.hass.version=${BUILD_VERSION}
