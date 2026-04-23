#!/bin/sh
set -e

CONFIG_FILE="/app/config.json"

# Read config file values (if present), env vars take precedence
cfg() {
  val=""
  if [ -f "$CONFIG_FILE" ]; then
    val=$(jq -r ".$1 // empty" "$CONFIG_FILE" 2>/dev/null)
  fi
  echo "$val"
}

LISTEN_HOST="${LISTEN_HOST:-$(cfg listenHost)}"
LISTEN_PORT="${LISTEN_PORT:-$(cfg listenPort)}"
SERVER_NAME="${SERVER_NAME:-$(cfg serverName)}"
API_BASE="${API_BASE:-$(cfg apiBase)}"

# Apply defaults
LISTEN_HOST="${LISTEN_HOST:-0.0.0.0}"
LISTEN_PORT="${LISTEN_PORT:-80}"
SERVER_NAME="${SERVER_NAME:-_}"
API_BASE="${API_BASE:-https://xpense-api.mrahman.xyz}"

# Generate runtime config for the frontend JS app
cat > /usr/share/nginx/html/config.js <<EOF
window.__CONFIG__ = {
  API_BASE: "${API_BASE}",
};
EOF

# Generate nginx config
cat > /etc/nginx/conf.d/default.conf <<EOF
server {
    listen ${LISTEN_HOST}:${LISTEN_PORT};
    server_name ${SERVER_NAME};

    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff2?)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
EOF

exec nginx -g 'daemon off;'
