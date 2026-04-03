# syntax=docker/dockerfile:1
# ══════════════════════════════════════════════════════════════════════════════
# Stage 1 · Build
# ══════════════════════════════════════════════════════════════════════════════
FROM golang:1.24-bookworm AS builder

ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=none

WORKDIR /app

# Only copy go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source code after dependencies are cached
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags "-s -w \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.Version=${VERSION} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.BuildDate=${BUILD_DATE} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.GitCommit=${GIT_COMMIT}" \
      -o /bin/expense-tracker .

# ══════════════════════════════════════════════════════════════════════════════
# Stage 2 · Runtime base
# ══════════════════════════════════════════════════════════════════════════════
FROM debian:bookworm-slim AS runtime-base

ARG DEBIAN_RELEASE_NAME=bookworm

# Use apt cache mounts to speed up package installation
RUN rm -f /etc/apt/apt.conf.d/docker-clean; echo 'Binary::apt::APT::Keep-Downloaded-Packages "true";' > /etc/apt/apt.conf.d/keep-cache
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y --no-install-recommends ca-certificates wget gnupg fontconfig && \
    echo 'Etc/UTC' > /etc/timezone

# Pre-seed EULA for ttf-mscorefonts-installer to avoid interactive prompt
RUN echo "ttf-mscorefonts-installer msttcorefonts/accepted-mscorefonts-eula select true" | debconf-set-selections

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    echo "deb http://deb.debian.org/debian ${DEBIAN_RELEASE_NAME} contrib" >> /etc/apt/sources.list && \
    apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
      fonts-lohit-beng-bengali \
      fonts-dejavu \
      ttf-mscorefonts-installer && \
    fc-cache -f

WORKDIR /app
COPY --from=builder /bin/expense-tracker /app/expense-tracker

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["wget", "-q", "--spider", "http://localhost:8080/healthz"]

ENTRYPOINT ["/app/expense-tracker"]
CMD ["serve"]

# ══════════════════════════════════════════════════════════════════════════════
# Stage 3a · wkhtmltopdf edition
# ══════════════════════════════════════════════════════════════════════════════
FROM runtime-base AS wkhtmltopdf

ARG TARGETARCH=amd64
ARG WKHTMLTOPDF_VERSION=0.12.6.1-3
ARG DEBIAN_RELEASE_NAME=bookworm

# Cache the heavy download
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    set -x && \
    apt-get update && \
    wget -q https://github.com/wkhtmltopdf/packaging/releases/download/${WKHTMLTOPDF_VERSION}/wkhtmltox_${WKHTMLTOPDF_VERSION}.${DEBIAN_RELEASE_NAME}_${TARGETARCH}.deb && \
    apt-get install -y --no-install-recommends ./wkhtmltox_${WKHTMLTOPDF_VERSION}.${DEBIAN_RELEASE_NAME}_${TARGETARCH}.deb && \
    ldconfig && \
    rm wkhtmltox_${WKHTMLTOPDF_VERSION}.${DEBIAN_RELEASE_NAME}_${TARGETARCH}.deb

RUN mkdir -p /app/configs /app/.sqlite && \
    chown -R 65535:65535 /app

USER 65535:65535

# ══════════════════════════════════════════════════════════════════════════════
# Stage 2b · Chromium Base (Stable Dependencies)
# ══════════════════════════════════════════════════════════════════════════════
FROM runtime-base AS chromium-base

# Install Chromium and its dependencies with cache mounts
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && \
    apt-get install -y --no-install-recommends \
      chromium \
      libnss3 \
      libxss1 \
      libasound2 \
      libatk-bridge2.0-0 \
      libgtk-3-0 && \
    fc-cache -f

# ══════════════════════════════════════════════════════════════════════════════
# Stage 3b · chromedp edition
# ══════════════════════════════════════════════════════════════════════════════
FROM chromium-base AS chromedp

ENV CHROME_PATH=/usr/bin/chromium
RUN mkdir -p /app/configs /app/.sqlite && \
    chown -R 65535:65535 /app

USER 65535:65535
