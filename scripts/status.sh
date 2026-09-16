#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$project_dir"
. "$project_dir/scripts/lib/container_lifecycle.sh"
env_file="$(resolve_env_file)"
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
load_env_defaults "$project_dir/docker/images.env" GO_BUILD_IMAGE RUNTIME_IMAGE
env_relative="${env_file#"$project_dir"/}"
export RUNTIME_ENV_FILE="../$env_relative"
compose_args=(--env-file "$env_file" -f "$project_dir/docker/compose.yml")
compose_preflight "${compose_args[@]}"
db_type="$(read_config_value DB_TYPE mongodb)"
case "$db_type" in
  mongodb) dependency_name="$(read_config_value MONGO_CONTAINER_NAME cashlenx-mongodb)" ;;
  mysql) dependency_name="$(read_config_value MYSQL_CONTAINER_NAME cashlenx-mysql)" ;;
  *) lifecycle_error "DB_TYPE must be mongodb or mysql."; exit 1 ;;
esac
dependency_state="$(container inspect --format '{{.State.Status}}' "$dependency_name" 2>/dev/null || true)"
printf 'dependency=%s\ndependency_state=%s\n' "$dependency_name" "${dependency_state:-missing}"
result=0
[[ "$dependency_state" == running ]] || result=1
diagnose_container "$(read_config_value BACKEND_CONTAINER_NAME cashlenx-server)" \
  "$(resolve_image_ref SERVER_IMAGE_NAME cashlenx-server SERVER_IMAGE_TAG latest)" "$(resolve_network_name)" \
  sh -ec 'wget -q -T 3 -O /dev/null "http://127.0.0.1:${SERVER_PORT:-10063}/api/${API_VERSION:-v1}/open/health"' || result=1
exit "$result"
