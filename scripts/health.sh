#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"

command -v curl >/dev/null 2>&1 || { echo "curl is required." >&2; exit 1; }
server_port="$(sed -n 's/^SERVER_PORT=//p' .env 2>/dev/null | tail -n 1)"
api_version="$(sed -n 's/^API_VERSION=//p' .env 2>/dev/null | tail -n 1)"
server_port="${server_port:-8080}"
api_version="${api_version:-v0}"
health_url="${SERVER_HEALTH_URL:-http://127.0.0.1:${server_port}/api/${api_version}/open/health}"

for attempt in $(seq 1 30); do
  if curl --fail --silent --show-error "$health_url" >/dev/null; then
    echo "Server is healthy: $health_url"
    exit 0
  fi
  sleep 2
done

echo "Server health check failed: $health_url" >&2
docker compose ps server >&2
exit 1
