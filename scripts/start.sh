#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/compose.yml"
. "$project_dir/scripts/lib/container_lifecycle.sh"

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
    function positive_integer(key) {
      return values[key] ~ /^[1-9][0-9]*$/
    }
    /^[A-Za-z_][A-Za-z0-9_]*=/ {
      values[$1] = clean(substr($0, index($0, "=") + 1))
    }
    END {
      boolean_keys[1] = "SCHEMA_VALIDATION"
      boolean_keys[2] = "AUTH_REGISTRATION_ENABLED"
      boolean_keys[3] = "SMTP_ENABLED"
      boolean_keys[4] = "METRICS_ENABLED"
      for (i = 1; i <= 4; i++) {
        key = boolean_keys[i]
        if (values[key] != "" && values[key] != "true" && values[key] != "false") print key
      }

      if (!valid_timezone(values["TIMEZONE"])) print "TIMEZONE"

      require_value("JWT_SECRET")
      require_value("ADMIN_PASSWORD")

      if (values["ENV"] != "dev" && values["ENV"] != "test" && values["ENV"] != "prod") print "ENV"
      if (values["LOG_LEVEL"] !~ /^(debug|info|warn|error|dpanic|panic|fatal)$/) print "LOG_LEVEL"
      if (!positive_integer("JWT_EXPIRATION_MINUTES")) print "JWT_EXPIRATION_MINUTES"
      if (!positive_integer("REFRESH_TOKEN_EXPIRATION_DAYS")) print "REFRESH_TOKEN_EXPIRATION_DAYS"
      if (!positive_integer("VERIFICATION_CODE_EXPIRE_MINUTES")) print "VERIFICATION_CODE_EXPIRE_MINUTES"
      if (!positive_integer("VERIFICATION_CODE_SEND_INTERVAL_SECONDS")) print "VERIFICATION_CODE_SEND_INTERVAL_SECONDS"
      if (!positive_integer("API_RATE_LIMIT_REQUESTS_PER_MINUTE")) print "API_RATE_LIMIT_REQUESTS_PER_MINUTE"
      if (!positive_integer("API_RATE_LIMIT_BURST")) print "API_RATE_LIMIT_BURST"

      if (values["ENV"] == "prod") {
        if (length(values["JWT_SECRET"]) < 32) print "JWT_SECRET"
        if (length(values["ADMIN_PASSWORD"]) < 12) print "ADMIN_PASSWORD"
        if (values["CORS_ORIGINS"] == "" || index(values["CORS_ORIGINS"], "*") > 0 || index(values["CORS_ORIGINS"], "http://") > 0) print "CORS_ORIGINS"
        if (values["METRICS_ENABLED"] == "true" && (length(values["METRICS_BEARER_TOKEN"]) < 32 || unsafe("METRICS_BEARER_TOKEN", values["METRICS_BEARER_TOKEN"]))) print "METRICS_BEARER_TOKEN"
      }

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
        if (!positive_integer("SMTP_PORT")) print "SMTP_PORT"
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

env_file="$(resolve_env_file)"
env_relative="${env_file#"$project_dir"/}"
validate_start_configuration
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
load_env_defaults "$project_dir/docker/images.env" GO_BUILD_IMAGE RUNTIME_IMAGE
network_name="$(resolve_network_name)"
container_name="$(read_config_value BACKEND_CONTAINER_NAME cashlenx-server)"
export RUNTIME_ENV_FILE="../$env_relative"
compose_args=(--env-file "$env_file" -f "$compose_file")
compose_preflight "${compose_args[@]}"
ensure_network "$network_name"
compose_up_quiet "${compose_args[@]}" up -d --no-build --remove-orphans server
wait_for_container_command "$container_name" sh -ec \
  'wget -q -T 3 -O /dev/null "http://127.0.0.1:${SERVER_PORT:-10063}/api/${API_VERSION:-v1}/open/health"'
