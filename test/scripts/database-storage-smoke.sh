#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
cd "$project_dir"
suffix="$$"
env_file="$(mktemp "$project_dir/.env.storage-smoke.XXXXXX")"
bind_env="$(mktemp "$project_dir/.env.storage-bind.XXXXXX")"
bind_dir="$(mktemp -d "$project_dir/.storage-bind.XXXXXX")"
mongo_container="clx34-mongodb-$suffix"
mysql_container="clx34-mysql-$suffix"
mongo_volume="clx34-mongodb-data-$suffix"
mysql_volume="clx34-mysql-data-$suffix"
network="clx34-network-$suffix"
mongo_project="clx34-mongodb-$suffix"
mysql_project="clx34-mysql-$suffix"
mongo_password="storage-smoke-mongo-password"
mysql_root_password="storage-smoke-mysql-root-password"
mysql_password="storage-smoke-mysql-password"
database="cashlenx_storage_smoke"
mongo_image="$(sed -n 's/^MONGO_IMAGE=//p' docker/dependencies/images.env | sed -n '1p')"
mongo_version="$(sed -n 's/^MONGO_VERSION=//p' docker/dependencies/images.env | sed -n '1p')"
mysql_image="$(sed -n 's/^MYSQL_IMAGE=//p' docker/dependencies/images.env | sed -n '1p')"
mysql_version="$(sed -n 's/^MYSQL_VERSION=//p' docker/dependencies/images.env | sed -n '1p')"
[[ "$mongo_image" =~ ^mongo:[A-Za-z0-9_.-]+@sha256:[0-9a-f]{64}$ ]]
[[ "$mysql_image" =~ ^mysql:[A-Za-z0-9_.-]+@sha256:[0-9a-f]{64}$ ]]

cleanup() {
  ENV_FILE="${env_file#"$project_dir/"}" bash scripts/dependencies/mongodb/stop.sh >/dev/null 2>&1 || true
  ENV_FILE="${env_file#"$project_dir/"}" bash scripts/dependencies/mysql/stop.sh >/dev/null 2>&1 || true
  docker rm -f "$mongo_container" "$mysql_container" >/dev/null 2>&1 || true
  docker volume rm "$mongo_volume" "$mysql_volume" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
  rm -f "$env_file" "$bind_env"
  rm -rf "$bind_dir"
}
trap cleanup EXIT
trap 'printf "Database storage smoke failed at line %s.\n" "$LINENO" >&2' ERR

sed \
  -e 's/^CONTAINER_FRONTEND=.*$/CONTAINER_FRONTEND=docker/' \
  -e "s/^DOCKER_NETWORK_NAME=.*$/DOCKER_NETWORK_NAME=$network/" \
  -e "s/^DB_NAME=.*$/DB_NAME=$database/" \
  -e "s/^MONGO_ROOT_PASSWORD=.*$/MONGO_ROOT_PASSWORD=$mongo_password/" \
  -e "s/^MONGO_CONTAINER_NAME=.*$/MONGO_CONTAINER_NAME=$mongo_container/" \
  -e 's/^MONGO_PORT=.*$/MONGO_PORT=0/' \
  -e "s/^MONGO_PROJECT_NAME=.*$/MONGO_PROJECT_NAME=$mongo_project/" \
  -e "s/^MONGO_DATA_VOLUME_NAME=.*$/MONGO_DATA_VOLUME_NAME=$mongo_volume/" \
  -e 's/^MONGO_STOP_GRACE_PERIOD=.*$/MONGO_STOP_GRACE_PERIOD=60s/' \
  -e "s/^MYSQL_ROOT_PASSWORD=.*$/MYSQL_ROOT_PASSWORD=$mysql_root_password/" \
  -e "s/^MYSQL_PASSWORD=.*$/MYSQL_PASSWORD=$mysql_password/" \
  -e "s/^MYSQL_CONTAINER_NAME=.*$/MYSQL_CONTAINER_NAME=$mysql_container/" \
  -e 's/^MYSQL_PORT=.*$/MYSQL_PORT=0/' \
  -e "s/^MYSQL_PROJECT_NAME=.*$/MYSQL_PROJECT_NAME=$mysql_project/" \
  -e "s/^MYSQL_DATA_VOLUME_NAME=.*$/MYSQL_DATA_VOLUME_NAME=$mysql_volume/" \
  .env.example > "$env_file"

env_name="${env_file#"$project_dir/"}"

mongo_start="$(ENV_FILE="$env_name" bash scripts/dependencies/mongodb/start.sh)"
grep -F "mongodb_image_version=$mongo_version" <<< "$mongo_start" >/dev/null
grep -E '^mongodb_storage_filesystem=(ext2/ext3|ext4|xfs|btrfs|zfs)$' <<< "$mongo_start" >/dev/null
test "$(docker inspect --format '{{.Config.Image}}' "$mongo_container")" = "$mongo_image"
docker exec "$mongo_container" mongosh --quiet \
  --username cashlenx --password "$mongo_password" --authenticationDatabase admin \
  "$database" --eval 'db.storage_probe.insertOne({_id: "persistence", value: "kept"}).acknowledged' | grep -F true >/dev/null
if ! mongo_stop="$(ENV_FILE="$env_name" bash scripts/dependencies/mongodb/stop.sh 2>&1)"; then
  printf '%s\n' "$mongo_stop" >&2
  false
fi
grep -F 'stop_result=graceful' <<< "$mongo_stop" >/dev/null
docker volume inspect "$mongo_volume" >/dev/null
ENV_FILE="$env_name" bash scripts/dependencies/mongodb/start.sh >/dev/null
docker exec "$mongo_container" mongosh --quiet \
  --username cashlenx --password "$mongo_password" --authenticationDatabase admin \
  "$database" --eval 'db.storage_probe.countDocuments({_id: "persistence"})' | grep -Fx 1 >/dev/null
ENV_FILE="$env_name" bash scripts/dependencies/mongodb/status.sh | grep -F 'image_identity=verified' >/dev/null
ENV_FILE="$env_name" bash scripts/dependencies/mongodb/stop.sh >/dev/null
docker volume inspect "$mongo_volume" >/dev/null

mysql_start="$(ENV_FILE="$env_name" bash scripts/dependencies/mysql/start.sh)"
grep -F "mysql_image_version=$mysql_version" <<< "$mysql_start" >/dev/null
test "$(docker inspect --format '{{.Config.Image}}' "$mysql_container")" = "$mysql_image"
docker exec "$mysql_container" mysql --batch --skip-column-names \
  -uroot -p"$mysql_root_password" "$database" \
  -e 'CREATE TABLE storage_probe (id VARCHAR(32) PRIMARY KEY, value VARCHAR(32)); INSERT INTO storage_probe VALUES ("persistence", "kept");'
ENV_FILE="$env_name" bash scripts/dependencies/mysql/stop.sh >/dev/null
docker volume inspect "$mysql_volume" >/dev/null
ENV_FILE="$env_name" bash scripts/dependencies/mysql/start.sh >/dev/null
docker exec "$mysql_container" mysql --batch --skip-column-names \
  -uroot -p"$mysql_root_password" "$database" \
  -e 'SELECT value FROM storage_probe WHERE id = "persistence";' | grep -Fx kept >/dev/null
ENV_FILE="$env_name" bash scripts/dependencies/mysql/status.sh | grep -F 'image_identity=verified' >/dev/null
ENV_FILE="$env_name" bash scripts/dependencies/mysql/stop.sh >/dev/null
docker volume inspect "$mysql_volume" >/dev/null

bind_path="$bind_dir"
if command -v cygpath >/dev/null 2>&1; then
  bind_path="$(cygpath -m "$bind_dir")"
fi
sed "s|^MONGO_DATA_PATH=$|MONGO_DATA_PATH=$bind_path|" "$env_file" > "$bind_env"
bind_env_name="${bind_env#"$project_dir/"}"
if bind_start="$(ENV_FILE="$bind_env_name" bash scripts/dependencies/mongodb/start.sh 2>&1)"; then
  grep -E '^mongodb_storage_filesystem=(ext2/ext3|ext4|xfs|btrfs|zfs)$' <<< "$bind_start" >/dev/null
  ENV_FILE="$bind_env_name" bash scripts/dependencies/mongodb/stop.sh >/dev/null
else
  grep -F "$bind_path" <<< "$bind_start" >/dev/null
  grep -F 'unsupported shared or remote filesystem' <<< "$bind_start" >/dev/null
  grep -F 'Leave MONGO_DATA_PATH empty' <<< "$bind_start" >/dev/null
  grep -F 'No data was changed' <<< "$bind_start" >/dev/null
  [[ -z "$(find "$bind_dir" -mindepth 1 -print -quit)" ]]
fi

printf '%s\n' 'Database storage smoke checks passed.'
