#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
. "$project_dir/scripts/lib/container_lifecycle.sh"
. "$project_dir/scripts/dependencies/image_pins.sh"
env_file="$(resolve_env_file)"
load_dependency_image_pin mysql
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
compose_args=(--env-file "$env_file" -f "$project_dir/docker/dependencies/mysql/compose.yml")
compose_preflight "${compose_args[@]}"
diagnose_container "$(read_config_value MYSQL_CONTAINER_NAME cashlenx-mysql)" \
  "$MYSQL_IMAGE" "$(resolve_network_name)" \
  sh -ec 'mysqladmin ping --silent -h 127.0.0.1 -P "$MYSQL_CONTAINER_PORT" -u root -p"$MYSQL_ROOT_PASSWORD"'
