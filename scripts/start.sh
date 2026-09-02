#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/compose.yml"
. "$project_dir/scripts/lib/container_readiness.sh"

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
    function unsafe(key, value) {
      if (index(value, "CHANGE_ME") > 0) return 1
      if (key == "JWT_SECRET" && value == "your-secret-key-here-change-in-production") return 1
      if (key == "ADMIN_PASSWORD" && value == "admin") return 1
      if ((key == "MONGO_ROOT_PASSWORD" || key == "MYSQL_PASSWORD") && value == "cashlenx123") return 1
      if ((key == "DOCKER_MONGO_DB_URI" || key == "DOCKER_MYSQL_DB_URI") && index(value, "cashlenx123") > 0) return 1
      return 0
    }
    function valid_timezone(value) {
      if (value == "" || value == "UTC") return 1
      if (index(value, "Etc/GMT") == 1) return 0
      return value ~ /^[A-Za-z][A-Za-z0-9._+-]*(\/[A-Za-z][A-Za-z0-9._+-]*)+$/
    }
    function require_value(key) {
      if (values[key] == "" || unsafe(key, values[key])) print key
    }
    /^[A-Za-z_][A-Za-z0-9_]*=/ {
      values[$1] = clean(substr($0, index($0, "=") + 1))
    }
    END {
      boolean_keys[1] = "SCHEMA_VALIDATION"
      boolean_keys[2] = "AUTH_REGISTRATION_ENABLED"
      boolean_keys[3] = "SMTP_ENABLED"
      for (i = 1; i <= 3; i++) {
        key = boolean_keys[i]
        if (values[key] != "" && values[key] != "true" && values[key] != "false") print key
      }

      if (!valid_timezone(values["TIMEZONE"])) print "TIMEZONE"

      require_value("JWT_SECRET")
      require_value("ADMIN_PASSWORD")

      db_type = values["DB_TYPE"] == "" ? "mongodb" : values["DB_TYPE"]
      if (db_type == "mongodb") {
        docker_uri = values["DOCKER_MONGO_DB_URI"]
        if (docker_uri == "") {
          print "DOCKER_MONGO_DB_URI"
        } else if (index(docker_uri, "MONGO_ROOT_USERNAME") > 0 || index(docker_uri, "MONGO_ROOT_PASSWORD") > 0) {
          require_value("MONGO_ROOT_USERNAME")
          require_value("MONGO_ROOT_PASSWORD")
        } else if (unsafe("DOCKER_MONGO_DB_URI", docker_uri)) {
          print "DOCKER_MONGO_DB_URI"
        }
      } else if (db_type == "mysql") {
        docker_uri = values["DOCKER_MYSQL_DB_URI"]
        if (docker_uri == "") {
          print "DOCKER_MYSQL_DB_URI"
        } else if (index(docker_uri, "MYSQL_USER") > 0 || index(docker_uri, "MYSQL_PASSWORD") > 0) {
          require_value("MYSQL_USER")
          require_value("MYSQL_PASSWORD")
        } else if (unsafe("DOCKER_MYSQL_DB_URI", docker_uri)) {
          print "DOCKER_MYSQL_DB_URI"
        }
      } else {
        print "DB_TYPE"
      }

      if (values["SMTP_ENABLED"] == "true") {
        require_value("SMTP_HOST")
        require_value("SMTP_PORT")
        require_value("SMTP_USERNAME")
        require_value("SMTP_PASSWORD")
        require_value("SMTP_FROM_ADDRESS")
      }
    }
  ' "$env_file"
}

validate_start_configuration() {
  local invalid_keys
  invalid_keys="$(invalid_configuration_keys | sort -u)"
  if [[ -n "$invalid_keys" ]]; then
    echo "Unsafe, placeholder, or missing environment values must be fixed before start:" >&2
    while IFS= read -r key; do
      [[ -n "$key" ]] && echo "  - $key" >&2
    done <<< "$invalid_keys"
    return 1
  fi
}

command -v docker >/dev/null 2>&1 || { echo "Docker is required." >&2; exit 1; }
docker compose version >/dev/null 2>&1 || { echo "Docker Compose is required." >&2; exit 1; }

env_file="$(resolve_env_file)"
env_relative="${env_file#"$project_dir"/}"
validate_start_configuration
network_name="$(resolve_network_name)"

ensure_network "$network_name"
container_name="$(read_env_value BACKEND_CONTAINER_NAME)"
container_name="${container_name:-cashlenx-server}"
RUNTIME_ENV_FILE="../$env_relative" \
  docker compose --env-file "$env_file" -f "$compose_file" up -d --no-build --remove-orphans server
wait_for_container_command "$container_name" sh -ec \
  'wget -q -T 3 -O /dev/null "http://127.0.0.1:${SERVER_PORT:-10063}/api/${API_VERSION:-v0}/open/health"'
