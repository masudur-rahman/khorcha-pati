# syntax=docker/dockerfile:1
# ══════════════════════════════════════════════════════════════════════════════
# Stage 1 · Build
# Uses BuildKit cache mounts so Go modules and the build cache persist
# between CI runs, dramatically reducing build time.
# ══════════════════════════════════════════════════════════════════════════════
FROM golang:1.26-bookworm AS builder

ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG GIT_COMMIT=none

WORKDIR /app

# Download dependencies first — this layer rebuilds only when go.mod/go.sum change
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Build the binary — source changes only rebuild from here
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go build \
      -ldflags "-s -w \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.Version=${VERSION} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.BuildDate=${BUILD_DATE} \
        -X github.com/masudur-rahman/expense-tracker-bot/cmd.GitCommit=${GIT_COMMIT}" \
      -o /expense-tracker .

# ══════════════════════════════════════════════════════════════════════════════
# Stage 2 · Runtime
# debian:bookworm-slim with wkhtmltopdf for PDF report generation.
# wkhtmltopdf requires system Qt/X11 libs that distroless cannot provide.
# ══════════════════════════════════════════════════════════════════════════════
FROM debian:bookworm-slim AS runtime

ARG TARGETARCH=amd64
ARG WKHTMLTOPDF_VERSION=0.12.6.1-3
ARG DEBIAN_RELEASE_NAME=bookworm

RUN set -x \
 && apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y --no-install-recommends ca-certificates wget \
 && echo 'Etc/UTC' > /etc/timezone

RUN echo "deb http://deb.debian.org/debian ${DEBIAN_RELEASE_NAME} contrib" >> /etc/apt/sources.list \
 && apt-get update \
 && DEBIAN_FRONTEND=noninteractive apt-get install -y \
      fonts-lohit-beng-bengali \
      fonts-dejavu \
      fontconfig \
      ttf-mscorefonts-installer

RUN set -x \
 && wget -q https://github.com/wkhtmltopdf/packaging/releases/download/${WKHTMLTOPDF_VERSION}/wkhtmltox_${WKHTMLTOPDF_VERSION}.${DEBIAN_RELEASE_NAME}_${TARGETARCH}.deb \
 && dpkg -i wkhtmltox_${WKHTMLTOPDF_VERSION}.${DEBIAN_RELEASE_NAME}_${TARGETARCH}.deb || true \
 && apt-get install -f -y \
 && ldconfig \
 && rm wkhtmltox_${WKHTMLTOPDF_VERSION}.${DEBIAN_RELEASE_NAME}_${TARGETARCH}.deb \
 && rm -rf /var/lib/apt/lists/* /usr/share/doc /usr/share/man /tmp/*

COPY --from=builder /expense-tracker /expense-tracker

USER 65535:65535

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD ["wget", "-q", "--spider", "http://localhost:8080/healthz"]

ENTRYPOINT ["/expense-tracker"]
CMD ["serve"]
