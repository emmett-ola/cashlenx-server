#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
. "$project_dir/scripts/lib/container_lifecycle.sh"
env_file="$(resolve_env_file)"
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
compose_args=(--env-file "$env_file" -f "$project_dir/docker/dependencies/mongodb/compose.yml")
compose_preflight "${compose_args[@]}"
diagnose_container "$(read_config_value MONGO_CONTAINER_NAME cashlenx-mongodb)" \
  "$(read_config_value MONGO_IMAGE mongo:7.0)" "$(resolve_network_name)" \
  sh -ec 'mongosh --quiet --username "$MONGO_INITDB_ROOT_USERNAME" --password "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin --eval "db.runCommand({ ping: 1 }).ok" "127.0.0.1:${MONGO_CONTAINER_PORT:-27017}/${MONGO_INITDB_DATABASE}" >/dev/null'
