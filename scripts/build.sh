#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$project_dir"

command -v docker >/dev/null 2>&1 || { echo "Docker is required." >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Docker Compose is required." >&2; exit 1; }
[[ -f .env ]] || { echo "Missing .env. Create it from .env.sample and set the UAT values." >&2; exit 1; }

git_commit="${GIT_COMMIT:-unknown}"
if [[ "$git_commit" == "unknown" ]] && command -v git >/dev/null 2>&1; then
  git_commit="$(git rev-parse --short HEAD 2>/dev/null || echo unknown)"
fi

GIT_COMMIT="$git_commit" docker compose --profile backend build server
