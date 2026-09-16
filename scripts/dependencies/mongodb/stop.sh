#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/dependencies/mongodb/compose.yml"
. "$project_dir/scripts/lib/container_lifecycle.sh"

env_file="$(resolve_env_file)"
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
network_name="$(resolve_network_name)"
compose_args=(--env-file "$env_file" -f "$compose_file")
# Stop remains available with incomplete credentials.
export MONGO_ROOT_USERNAME="${MONGO_ROOT_USERNAME:-dependency-stop}" MONGO_ROOT_PASSWORD="${MONGO_ROOT_PASSWORD:-dependency-stop}"
compose_preflight "${compose_args[@]}"
compose "${compose_args[@]}" down --remove-orphans
remove_network_if_unused "$network_name"
