#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/dependencies/mysql/compose.yml"
. "$project_dir/scripts/lib/container_lifecycle.sh"

env_file="$(resolve_env_file)"
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
network_name="$(resolve_network_name)"
container_name="$(read_config_value MYSQL_CONTAINER_NAME cashlenx-mysql)"
stop_grace_period="$(read_config_value MYSQL_STOP_GRACE_PERIOD 30s)"
compose_args=(--env-file "$env_file" -f "$compose_file")
# Stop remains available with incomplete credentials.
export MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-dependency-stop}" MYSQL_USER="${MYSQL_USER:-dependency-stop}" MYSQL_PASSWORD="${MYSQL_PASSWORD:-dependency-stop}"
compose_preflight "${compose_args[@]}"
stop_result=0
stop_container_bounded "$container_name" "$stop_grace_period" || stop_result=$?
compose "${compose_args[@]}" down --remove-orphans
remove_network_if_unused "$network_name"
exit "$stop_result"
