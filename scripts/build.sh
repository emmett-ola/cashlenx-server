#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$project_dir"
compose_file="$project_dir/docker/compose.yml"
. "$project_dir/scripts/lib/container_lifecycle.sh"

env_file="$(resolve_env_file)"
container_runtime_init "$(read_config_value CONTAINER_FRONTEND auto)"
load_env_defaults "$project_dir/docker/images.env" GO_BUILD_IMAGE GO_VERSION RUNTIME_IMAGE
git_commit="${GIT_COMMIT:-$(git rev-parse HEAD)}"
[[ "$git_commit" =~ ^[0-9a-fA-F]{40}$ ]] || { echo "GIT_COMMIT must be a full 40-character Git revision." >&2; exit 1; }
product_version="${PRODUCT_VERSION:-$(sed -n 's/^const Version = "\([^"]*\)"/\1/p' model/version.go | sed -n '1p' | tr -d '\r')}"
[[ "$product_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$ ]] || { echo "PRODUCT_VERSION must be a semantic version." >&2; exit 1; }
build_time="${BUILD_TIME:-$(git show -s --format=%cI "$git_commit")}"
[[ "$build_time" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2}T ]] || { echo "BUILD_TIME must be an ISO 8601 timestamp." >&2; exit 1; }
image_ref="$(resolve_image_ref SERVER_IMAGE_NAME cashlenx-server SERVER_IMAGE_TAG latest)"

export RUNTIME_ENV_FILE="../${env_file#"$project_dir"/}" GIT_COMMIT="$git_commit" PRODUCT_VERSION="$product_version" BUILD_TIME="$build_time"
compose_args=(--env-file "$env_file" -f "$compose_file")
compose_preflight "${compose_args[@]}"
compose "${compose_args[@]}" build server
"$project_dir/scripts/verify-image.sh" "$image_ref" "$product_version" "$git_commit"
