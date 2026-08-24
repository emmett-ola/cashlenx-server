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

ensure_network() {
  local name="$1"
  if docker network inspect "$name" >/dev/null 2>&1; then
    return 0
  fi
  docker network create --driver bridge "$name" >/dev/null 2>&1 ||
    docker network inspect "$name" >/dev/null 2>&1
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
    function valid_volume_name(value) {
      return value ~ /^[A-Za-z0-9][A-Za-z0-9_.-]*$/
    }
    function valid_data_path(value, normalized) {
      if (value == "") return 1
      normalized = value
      gsub(/\\/, "/", normalized)
      if (normalized == "/" || normalized ~ /(^|\/)\.\.(\/|$)/) return 0
      if (substr(normalized, 1, 1) == "/" && length(normalized) > 1) return 1
      if (length(normalized) > 3 && substr(normalized, 2, 1) == ":" && substr(normalized, 3, 1) == "/") return 1
      return 0
    }
    /^[A-Za-z_][A-Za-z0-9_]*=/ {
      values[$1] = clean(substr($0, index($0, "=") + 1))
    }
    END {
      if (!valid_timezone(values["TIMEZONE"])) print "TIMEZONE"
      if (!valid_data_path(values["MYSQL_DATA_PATH"])) print "MYSQL_DATA_PATH"
      if (values["MYSQL_DATA_PATH"] == "" && !valid_volume_name(values["MYSQL_DATA_VOLUME_NAME"])) print "MYSQL_DATA_VOLUME_NAME"

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
network_name="$(resolve_network_name)"

ensure_network "$network_name"
docker compose --env-file "$env_file" -f "$compose_file" \
  up -d --no-build --pull never --remove-orphans --wait mysql
