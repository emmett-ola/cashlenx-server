#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
fake_dir="$(mktemp -d)"
fake_log="$fake_dir/docker.log"
api_env="$(mktemp "$project_dir/.env.lifecycle-api.XXXXXX")"
mysql_env="$(mktemp "$project_dir/.env.lifecycle-mysql.XXXXXX")"
smtp_env="$(mktemp "$project_dir/.env.lifecycle-smtp.XXXXXX")"
invalid_boolean_env="$(mktemp "$project_dir/.env.lifecycle-boolean.XXXXXX")"
invalid_timezone_env="$(mktemp "$project_dir/.env.lifecycle-timezone.XXXXXX")"
invalid_storage_env="$(mktemp "$project_dir/.env.lifecycle-storage.XXXXXX")"
outside_env="$(mktemp)"

cleanup() {
  rm -f "$api_env" "$mysql_env" "$smtp_env" "$invalid_boolean_env" \
    "$invalid_timezone_env" "$outside_env"
  rm -f "$invalid_storage_env"
  rm -rf "$fake_dir"
}
trap cleanup EXIT

cat > "$fake_dir/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$FAKE_DOCKER_LOG"
if [[ "${1:-}" == "network" && "${2:-}" == "inspect" ]]; then
  if [[ "$*" == *"--format"* ]]; then
    printf '%s\n' "${FAKE_NETWORK_CONNECTIONS:-1}"
    exit 0
  fi
  [[ "${FAKE_NETWORK_EXISTS:-true}" == "true" ]] || exit 1
fi
EOF
chmod +x "$fake_dir/docker"

sed \
  -e 's/CHANGE_ME_MONGO_PASSWORD/lifecycle-mongo-password/g' \
  -e 's/CHANGE_ME_JWT_SECRET/lifecycle-jwt-secret-with-more-than-32-bytes/g' \
  -e 's/CHANGE_ME_ADMIN_PASSWORD/lifecycle-admin-password/g' \
  -e 's/^TIMEZONE=UTC$/TIMEZONE=Asia\/Shanghai/' \
  "$project_dir/.env.example" > "$api_env"
sed \
  -e 's/CHANGE_ME_MYSQL_ROOT_PASSWORD/lifecycle-mysql-root-password/g' \
  -e 's/CHANGE_ME_MYSQL_PASSWORD/lifecycle-mysql-password/g' \
  -e 's/^DB_TYPE=mongodb$/DB_TYPE=mysql/' \
  "$api_env" > "$mysql_env"
sed 's/^SMTP_ENABLED=false$/SMTP_ENABLED=true/' "$api_env" > "$smtp_env"
sed 's/^SMTP_ENABLED=false$/SMTP_ENABLED=ture/' "$api_env" > "$invalid_boolean_env"
printf 'ENV=dev\n' > "$outside_env"

api_env_name="${api_env#"$project_dir/"}"
mysql_env_name="${mysql_env#"$project_dir/"}"
smtp_env_name="${smtp_env#"$project_dir/"}"
invalid_boolean_env_name="${invalid_boolean_env#"$project_dir/"}"
invalid_timezone_env_name="${invalid_timezone_env#"$project_dir/"}"
invalid_storage_env_name="${invalid_storage_env#"$project_dir/"}"
test_path="$fake_dir:$PATH"

reset_log() {
  : > "$fake_log"
}

run_script() {
  local script="$1"
  local selected_env="${2:-$api_env_name}"
  PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" \
    FAKE_NETWORK_EXISTS="${FAKE_NETWORK_EXISTS:-true}" \
    FAKE_NETWORK_CONNECTIONS="${FAKE_NETWORK_CONNECTIONS:-1}" \
    ENV_FILE="$selected_env" \
    bash "$project_dir/$script"
}

assert_log_contains() {
  local expected="$1"
  grep -F -- "$expected" "$fake_log" >/dev/null || {
    echo "Expected fake Docker call containing: $expected" >&2
    return 1
  }
}

assert_log_not_contains() {
  local unexpected="$1"
  if grep -F -- "$unexpected" "$fake_log" >/dev/null; then
    echo "Unexpected fake Docker call containing: $unexpected" >&2
    return 1
  fi
}

assert_rejected() {
  local env_file="$1"
  local script="$2"
  local expected_key="$3"
  local forbidden_value="${4:-}"
  local output
  if output="$(PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" ENV_FILE="$env_file" bash "$project_dir/$script" 2>&1)"; then
    echo "Expected $script to reject $env_file" >&2
    return 1
  fi
  grep -F -- "$expected_key" <<< "$output" >/dev/null || {
    echo "Expected rejection to identify $expected_key" >&2
    return 1
  }
  if grep -F -- 'CHANGE_ME_' <<< "$output" >/dev/null; then
    echo "Rejection output exposed a placeholder value" >&2
    return 1
  fi
  if [[ -n "$forbidden_value" ]] && grep -F -- "$forbidden_value" <<< "$output" >/dev/null; then
    echo "Rejection output exposed the configured value for $expected_key" >&2
    return 1
  fi
}

reset_log
run_script scripts/dependencies/mongodb/build.sh
assert_log_contains "-f $project_dir/docker/dependencies/mongodb/compose.yml pull mongodb"

reset_log
assert_rejected .env.example scripts/dependencies/mongodb/start.sh MONGO_ROOT_PASSWORD
if grep -F -- ' up ' "$fake_log" >/dev/null; then
  echo "MongoDB start reached Docker after rejecting placeholders" >&2
  exit 1
fi

reset_log
FAKE_NETWORK_EXISTS=false run_script scripts/dependencies/mongodb/start.sh
assert_log_contains "network create --driver bridge cashlenx-network"
assert_log_contains "--pull never --remove-orphans mongodb"
assert_log_contains "exec cashlenx-mongodb sh -ec mongosh"
assert_log_not_contains "--wait"

reset_log
ENV_FILE=.env.example PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" \
  FAKE_NETWORK_EXISTS=true FAKE_NETWORK_CONNECTIONS=1 \
  bash "$project_dir/scripts/dependencies/mongodb/stop.sh"
assert_log_contains "-f $project_dir/docker/dependencies/mongodb/compose.yml down --remove-orphans"
if grep -F -- 'network rm' "$fake_log" >/dev/null; then
  echo "MongoDB stop removed a shared network with attached containers" >&2
  exit 1
fi

reset_log
run_script scripts/dependencies/mysql/build.sh
assert_log_contains "-f $project_dir/docker/dependencies/mysql/compose.yml pull mysql"

reset_log
assert_rejected .env.example scripts/dependencies/mysql/start.sh MYSQL_ROOT_PASSWORD
if grep -F -- ' up ' "$fake_log" >/dev/null; then
  echo "MySQL start reached Docker after rejecting missing values" >&2
  exit 1
fi

reset_log
run_script scripts/dependencies/mysql/start.sh "$mysql_env_name"
assert_log_contains "--pull never --remove-orphans mysql"
assert_log_contains "exec cashlenx-mysql sh -ec mysqladmin ping"
assert_log_not_contains "--wait"

reset_log
ENV_FILE=.env.example PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" \
  bash "$project_dir/scripts/dependencies/mysql/stop.sh"
assert_log_contains "-f $project_dir/docker/dependencies/mysql/compose.yml down --remove-orphans"

reset_log
run_script scripts/start.sh
assert_log_contains "-f $project_dir/docker/compose.yml up -d --no-build --remove-orphans server"
assert_log_contains "exec cashlenx-server sh -ec wget"
assert_log_not_contains "--wait"
if grep -F -- 'dependencies/' "$fake_log" >/dev/null; then
  echo "API start unexpectedly invoked a dependency Compose project" >&2
  exit 1
fi

reset_log
FAKE_NETWORK_EXISTS=true FAKE_NETWORK_CONNECTIONS=0 run_script scripts/stop.sh
assert_log_contains "-f $project_dir/docker/compose.yml down --remove-orphans"
assert_log_contains "network rm cashlenx-network"

reset_log
run_script scripts/start.sh "$mysql_env_name"
assert_log_contains "-f $project_dir/docker/compose.yml up -d --no-build --remove-orphans server"
assert_log_contains "exec cashlenx-server sh -ec wget"
assert_log_not_contains "--wait"

reset_log
assert_rejected "$smtp_env_name" scripts/start.sh SMTP_PASSWORD
if grep -F -- ' up ' "$fake_log" >/dev/null; then
  echo "API start reached Docker with enabled placeholder SMTP credentials" >&2
  exit 1
fi

reset_log
assert_rejected "$invalid_boolean_env_name" scripts/start.sh SMTP_ENABLED

for invalid_timezone in UTC+8 CST Etc/GMT+8; do
  sed "s|^TIMEZONE=Asia/Shanghai$|TIMEZONE=$invalid_timezone|" \
    "$api_env" > "$invalid_timezone_env"
  reset_log
  assert_rejected "$invalid_timezone_env_name" scripts/start.sh TIMEZONE "$invalid_timezone"
  assert_rejected "$invalid_timezone_env_name" scripts/dependencies/mongodb/start.sh TIMEZONE "$invalid_timezone"
  assert_rejected "$invalid_timezone_env_name" scripts/dependencies/mysql/start.sh TIMEZONE "$invalid_timezone"
  if grep -F -- ' up ' "$fake_log" >/dev/null; then
    echo "Start reached Docker with an unsupported timezone" >&2
    exit 1
  fi
done

for invalid_path in relative/data /; do
  sed "s|^MONGO_DATA_PATH=$|MONGO_DATA_PATH=$invalid_path|" \
    "$api_env" > "$invalid_storage_env"
  reset_log
  assert_rejected "$invalid_storage_env_name" scripts/dependencies/mongodb/start.sh MONGO_DATA_PATH "$invalid_path"

  sed "s|^MYSQL_DATA_PATH=$|MYSQL_DATA_PATH=$invalid_path|" \
    "$mysql_env" > "$invalid_storage_env"
  reset_log
  assert_rejected "$invalid_storage_env_name" scripts/dependencies/mysql/start.sh MYSQL_DATA_PATH "$invalid_path"
done

sed 's|^MONGO_DATA_PATH=$|MONGO_DATA_PATH=/srv/cashlenx/mongodb|' \
  "$api_env" > "$invalid_storage_env"
reset_log
run_script scripts/dependencies/mongodb/start.sh "$invalid_storage_env_name"
assert_log_contains "--remove-orphans mongodb"
assert_log_contains "exec cashlenx-mongodb sh -ec mongosh"
assert_log_not_contains "--wait"

sed 's|^MYSQL_DATA_PATH=$|MYSQL_DATA_PATH=C:/cashlenx/mysql|' \
  "$mysql_env" > "$invalid_storage_env"
reset_log
run_script scripts/dependencies/mysql/start.sh "$invalid_storage_env_name"
assert_log_contains "--remove-orphans mysql"
assert_log_contains "exec cashlenx-mysql sh -ec mysqladmin ping"
assert_log_not_contains "--wait"

sed 's/^MONGO_DATA_VOLUME_NAME=.*$/MONGO_DATA_VOLUME_NAME=invalid name/' \
  "$api_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/dependencies/mongodb/start.sh MONGO_DATA_VOLUME_NAME 'invalid name'

sed 's/^MYSQL_DATA_VOLUME_NAME=.*$/MYSQL_DATA_VOLUME_NAME=invalid name/' \
  "$mysql_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/dependencies/mysql/start.sh MYSQL_DATA_VOLUME_NAME 'invalid name'

sed 's|^DOCKER_MONGO_DB_URI=.*$|DOCKER_MONGO_DB_URI=|' \
  "$api_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/start.sh DOCKER_MONGO_DB_URI
run_script scripts/build.sh "$invalid_storage_env_name"
assert_log_contains "build server"

sed 's|^DOCKER_MYSQL_DB_URI=.*$|DOCKER_MYSQL_DB_URI=|' \
  "$mysql_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/start.sh DOCKER_MYSQL_DB_URI

sed 's/^DOCKER_NETWORK_NAME=.*$/DOCKER_NETWORK_NAME=invalid network name/' \
  "$api_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/start.sh DOCKER_NETWORK_NAME 'invalid network name'
if grep -F -- ' up ' "$fake_log" >/dev/null; then
  echo "Start reached Compose with an invalid Docker network name" >&2
  exit 1
fi

reset_log
assert_rejected .env.lifecycle-missing scripts/dependencies/mongodb/build.sh "Missing environment file"
assert_rejected "$outside_env" scripts/dependencies/mongodb/build.sh "ENV_FILE must stay inside"

echo "Dependency lifecycle smoke checks passed."
