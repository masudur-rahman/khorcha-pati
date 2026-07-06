# syntax=docker/dockerfile:1
# ══════════════════════════════════════════════════════════════════════════════
# Stage 1 · Build
# ══════════════════════════════════════════════════════════════════════════════
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# ══════════════════════════════════════════════════════════════════════════════
# Stage 2 · Serve
# ══════════════════════════════════════════════════════════════════════════════
FROM nginx:alpine

LABEL org.opencontainers.image.source="https://github.com/masudur-rahman/khorcha-pati"

RUN apk add --no-cache jq

# Copy built assets
COPY --from=build /app/dist /usr/share/nginx/html

# Remove default nginx config — entrypoint generates it at runtime
RUN rm -f /etc/nginx/conf.d/default.conf

COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Runtime-configurable via env vars or /app/config.json
# See config.example.json for file format

EXPOSE 65262

ENTRYPOINT ["/docker-entrypoint.sh"]
