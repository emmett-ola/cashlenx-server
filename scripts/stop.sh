#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"
compose_file="$project_dir/docker/compose.yml"

command -v docker >/dev/null 2>&1 || { echo "Docker is required." >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Docker Compose is required." >&2; exit 1; }
[[ -f .env ]] || { echo "Missing .env. Create it from .env.sample and set the UAT values." >&2; exit 1; }

# This project owns only the API container. Images, bind-mounted logs, database
# dependency projects, and their named volumes are intentionally preserved.
docker compose -f "$compose_file" down --remove-orphans
