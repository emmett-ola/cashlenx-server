#!/usr/bin/env bash
set -euo pipefail
project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd -P)"
cd "$project_dir"
. "$project_dir/scripts/lib/container_lifecycle.sh"
env_file="$(resolve_env_file)"
: "$env_file"
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
show_container_logs "$(read_config_value MONGO_CONTAINER_NAME cashlenx-mongodb)" "${1:-100}"
