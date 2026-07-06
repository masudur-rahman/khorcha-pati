// Runtime config — overwritten by docker-entrypoint.sh in production.
// For local dev, vite proxy handles /api/ routing and the code defaults
// (KhorchaPatiBot / khorcha-pati repo) apply when a key is omitted.
// Supported keys: API_BASE, BOT_URL, REPO_URL.
window.__CONFIG__ = {};
