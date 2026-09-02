#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/dependencies/mongodb/compose.yml"

resolve_env_file() {
  local requested="${ENV_FILE:-.env}"
  local candidate
  if [[ "$requested" == /* ]]; then
    candidate="$requested"
  else
    candidate="$project_dir/$requested"
  fi

  if [[ ! -e "$candidate" ]]; then
    echo "Missing environment file: $requested" >&2
    echo "Create it with: cp .env.example \"$requested\"" >&2
    return 1
  fi
  [[ -f "$candidate" ]] || { echo "Environment path is not a file: $requested" >&2; return 1; }

  local resolved
  resolved="$(realpath "$candidate")"
  case "$resolved" in
    "$project_dir"/*) printf '%s\n' "$resolved" ;;
    *) echo "ENV_FILE must stay inside $project_dir: $requested" >&2; return 1 ;;
  esac
}

read_env_value() {
  local key="$1"
  awk -F= -v wanted="$key" '
    $0 ~ "^[[:space:]]*(export[[:space:]]+)?" wanted "[[:space:]]*=" {
      value = substr($0, index($0, "=") + 1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      if ((substr(value, 1, 1) == "\"" && substr(value, length(value), 1) == "\"") ||
          (substr(value, 1, 1) == "\047" && substr(value, length(value), 1) == "\047")) {
        value = substr(value, 2, length(value) - 2)
      }
      result = value
    }
    END { print result }
  ' "$env_file"
}

resolve_network_name() {
  local name
  name="$(read_env_value DOCKER_NETWORK_NAME)"
  name="${name:-cashlenx-network}"
  if [[ ! "$name" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]*$ ]]; then
    echo "Invalid Docker network setting: DOCKER_NETWORK_NAME" >&2
    return 1
  fi
  printf '%s\n' "$name"
}

remove_network_if_unused() {
  local name="$1"
  local connected
  if ! docker network inspect "$name" >/dev/null 2>&1; then
    return 0
  fi
  connected="$(docker network inspect --format '{{len .Containers}}' "$name" 2>/dev/null || true)"
  if [[ "$connected" == "0" ]]; then
    docker network rm "$name" >/dev/null 2>&1 || true
  fi
}

command -v docker >/dev/null 2>&1 || { echo "Docker is required." >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Docker Compose is required." >&2; exit 1; }

env_file="$(resolve_env_file)"
network_name="$(resolve_network_name)"

# Stop remains available with incomplete credentials. The named data volume and
# configured image are intentionally preserved.
MONGO_ROOT_USERNAME=dependency-stop MONGO_ROOT_PASSWORD=dependency-stop \
  docker compose --env-file "$env_file" -f "$compose_file" down --remove-orphans
remove_network_if_unused "$network_name"
