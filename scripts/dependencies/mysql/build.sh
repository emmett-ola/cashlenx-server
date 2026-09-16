#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/dependencies/mysql/compose.yml"
. "$project_dir/scripts/lib/container_lifecycle.sh"
. "$project_dir/scripts/dependencies/image_pins.sh"

env_file="$(resolve_env_file)"
load_dependency_image_pin mysql
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
compose_args=(--env-file "$env_file" -f "$compose_file")
compose_preflight "${compose_args[@]}"
compose "${compose_args[@]}" pull mysql
verify_dependency_image_version mysql
