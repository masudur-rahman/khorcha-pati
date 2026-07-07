# syntax=docker/dockerfile:1
# ══════════════════════════════════════════════════════════════════════════════
# Application image. The heavy runtime layers (fonts, wkhtmltopdf, chromium) live
# in prebuilt base images (see Dockerfile.base) so this build only cross-compiles
# the Go binary and copies it in.
#
# Base references are injected as build args by the Makefile so each registry
# (ghcr / docker hub) resolves its own base. Pin to an immutable digest in CI
# for supply-chain safety, e.g. BASE_WK=<registry>/khorcha-pati-base@sha256:…
# ══════════════════════════════════════════════════════════════════════════════
ARG BASE_WK=masudjuly02/khorcha-pati-base:wkhtmltopdf
ARG BASE_CHROMEDP=masudjuly02/khorcha-pati-base:chromedp

# ── Build (runs natively on the builder platform, cross-compiles per target) ──
FROM --platform=$BUILDPLATFORM golang:1.26-bookworm AS builder

ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=none
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
      -ldflags "-s -w \
        -X github.com/masudur-rahman/khorcha-pati/cmd.Version=${VERSION} \
        -X github.com/masudur-rahman/khorcha-pati/cmd.BuildDate=${BUILD_DATE} \
        -X github.com/masudur-rahman/khorcha-pati/cmd.GitCommit=${GIT_COMMIT}" \
      -o /bin/khorcha-pati .

# ── wkhtmltopdf edition ───────────────────────────────────────────────────────
FROM ${BASE_WK} AS wkhtmltopdf
COPY --from=builder --chown=65535:65535 /bin/khorcha-pati /app/khorcha-pati

# ── chromedp edition ──────────────────────────────────────────────────────────
FROM ${BASE_CHROMEDP} AS chromedp
COPY --from=builder --chown=65535:65535 /bin/khorcha-pati /app/khorcha-pati
