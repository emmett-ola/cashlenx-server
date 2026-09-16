#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/dependencies/mysql/compose.yml"
. "$project_dir/scripts/lib/container_lifecycle.sh"
. "$project_dir/scripts/dependencies/image_pins.sh"

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
      if (normalized == "/" || normalized ~ /(^|\/)\.\.(\/|$)/ || index(normalized, ",") > 0) return 0
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

env_file="$(resolve_env_file)"
validate_start_configuration
load_dependency_image_pin mysql
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
network_name="$(resolve_network_name)"
container_name="$(read_config_value MYSQL_CONTAINER_NAME cashlenx-mysql)"
compose_args=(--env-file "$env_file" -f "$compose_file")
compose_preflight "${compose_args[@]}"
verify_dependency_image_version mysql
ensure_network "$network_name"
compose_up_quiet "${compose_args[@]}" up -d --no-build --pull never --remove-orphans mysql
wait_for_container_command "$container_name" sh -ec \
  'mysqladmin ping -h 127.0.0.1 -uroot -p"$MYSQL_ROOT_PASSWORD" --silent'
