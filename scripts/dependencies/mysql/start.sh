#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/dependencies/mysql/compose.yml"

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
  [[ ! -L "$candidate" ]] || { echo "Environment file symlinks are not allowed: $requested" >&2; return 1; }

  local resolved_dir resolved
  resolved_dir="$(cd "$(dirname "$candidate")" && pwd -P)"
  resolved="$resolved_dir/$(basename "$candidate")"
  case "$resolved" in
    "$project_dir"/*) printf '%s\n' "$resolved" ;;
    *) echo "ENV_FILE must stay inside $project_dir: $requested" >&2; return 1 ;;
  esac
}

invalid_configuration_keys() {
  awk -F= '
    function clean(value) {
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      if ((substr(value, 1, 1) == "\"" && substr(value, length(value), 1) == "\"") ||
          (substr(value, 1, 1) == "\047" && substr(value, length(value), 1) == "\047")) {
        value = substr(value, 2, length(value) - 2)
      }
      return value
    }
    function valid_timezone(value) {
      if (value == "" || value == "UTC") return 1
      if (index(value, "Etc/GMT") == 1) return 0
      return value ~ /^[A-Za-z][A-Za-z0-9._+-]*(\/[A-Za-z][A-Za-z0-9._+-]*)+$/
    }
    /^[A-Za-z_][A-Za-z0-9_]*=/ {
      values[$1] = clean(substr($0, index($0, "=") + 1))
    }
    END {
      if (!valid_timezone(values["TIMEZONE"])) print "TIMEZONE"

      required[1] = "MYSQL_ROOT_PASSWORD"
      required[2] = "MYSQL_USER"
      required[3] = "MYSQL_PASSWORD"
      for (i = 1; i <= 3; i++) {
        key = required[i]
        value = values[key]
        if (value == "" || index(value, "CHANGE_ME") > 0 ||
            ((key == "MYSQL_ROOT_PASSWORD" || key == "MYSQL_PASSWORD") && value == "cashlenx123")) {
          print key
        }
      }
    }
  ' "$env_file"
}

validate_start_configuration() {
  local invalid_keys
  invalid_keys="$(invalid_configuration_keys | sort -u)"
  if [[ -n "$invalid_keys" ]]; then
    echo "Unsafe, placeholder, or missing MySQL values must be fixed before start:" >&2
    while IFS= read -r key; do
      [[ -n "$key" ]] && echo "  - $key" >&2
    done <<< "$invalid_keys"
    return 1
  fi
}

command -v docker >/dev/null 2>&1 || { echo "Docker is required." >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Docker Compose is required." >&2; exit 1; }

env_file="$(resolve_env_file)"
validate_start_configuration

docker compose --env-file "$env_file" -f "$compose_file" \
  up -d --no-build --pull never --remove-orphans --wait mysql
