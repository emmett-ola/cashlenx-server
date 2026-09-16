#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
fake_dir="$(mktemp -d)"
fake_log="$fake_dir/docker.log"
fake_stop_marker="$fake_dir/stopped"
api_env="$(mktemp "$project_dir/.env.lifecycle-api.XXXXXX")"
mysql_env="$(mktemp "$project_dir/.env.lifecycle-mysql.XXXXXX")"
smtp_env="$(mktemp "$project_dir/.env.lifecycle-smtp.XXXXXX")"
invalid_boolean_env="$(mktemp "$project_dir/.env.lifecycle-boolean.XXXXXX")"
invalid_timezone_env="$(mktemp "$project_dir/.env.lifecycle-timezone.XXXXXX")"
invalid_storage_env="$(mktemp "$project_dir/.env.lifecycle-storage.XXXXXX")"
prod_env="$(mktemp "$project_dir/.env.lifecycle-prod.XXXXXX")"
outside_env="$(mktemp)"
invalid_images="$(mktemp "$project_dir/docker/dependencies/images.invalid.XXXXXX")"
inside_symlink="$project_dir/.env.lifecycle-link"
outside_symlink="$project_dir/.env.lifecycle-outside-link"

cleanup() {
  rm -f "$api_env" "$mysql_env" "$smtp_env" "$invalid_boolean_env" \
    "$invalid_timezone_env" "$outside_env" "$invalid_images"
  rm -f "$invalid_storage_env" "$inside_symlink" "$outside_symlink"
  rm -f "$prod_env"
  rm -rf "$fake_dir"
}
trap cleanup EXIT

cat > "$fake_dir/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$FAKE_DOCKER_LOG"
if [[ "${1:-}" == "--version" ]]; then
  if [[ "${FAKE_FRONTEND_KIND:-docker}" == "docker" ]]; then
    printf '%s\n' 'Docker version 29.0.0, build fake'
  elif [[ "${FAKE_FRONTEND_KIND:-docker}" == "old-nerdctl" ]]; then
    printf '%s\n' 'nerdctl version 2.1.0'
  else
    printf '%s\n' 'nerdctl version 2.2.0'
  fi
  exit 0
fi
if [[ "${1:-}" == "compose" && "${2:-}" == "version" ]]; then
  if [[ "${FAKE_FRONTEND_KIND:-docker}" == "docker" ]]; then
    printf '%s\n' 'Docker Compose version v2.29.0'
  elif [[ "${FAKE_FRONTEND_KIND:-docker}" == "old-nerdctl" ]]; then
    printf '%s\n' 'nerdctl compose version 2.1.0'
  else
    printf '%s\n' 'nerdctl compose version 2.2.0'
  fi
  if [[ "${FAKE_VERBOSE_VERSION:-false}" == "true" ]]; then
    for ((line = 0; line < 4096; line++)); do
      printf '%s\n' 'nerdctl compose version 2.2.0'
    done
  fi
  exit 0
fi
if [[ "${1:-}" == "info" ]]; then
  exit 0
fi
if [[ "${1:-}" == "compose" && "$*" == *" config --quiet"* ]]; then
  [[ "${FAKE_CONFIG_SUPPORTED:-true}" == "true" ]]
  exit
fi
if [[ "${1:-}" == "compose" && "$*" == *" up "* && -n "${FAKE_SECRET_OUTPUT:-}" ]]; then
  printf '%s\n' "$FAKE_SECRET_OUTPUT" >&2
fi
if [[ "${1:-}" == "network" && "${2:-}" == "inspect" ]]; then
  if [[ "$*" == *"--format"* ]]; then
    printf '%s\n' "${FAKE_NETWORK_CONNECTIONS:-1}"
    exit 0
  fi
  [[ "${FAKE_NETWORK_EXISTS:-true}" == "true" ]] || exit 1
fi
if [[ "${1:-}" == "image" && "${2:-}" == "inspect" ]]; then
  [[ "${FAKE_IMAGE_EXISTS:-true}" == "true" ]] || exit 1
  if [[ "$*" == *"org.opencontainers.image.version"* ]]; then
    printf '%s\n' "${FAKE_IMAGE_VERSION:-0.0.0}"
  elif [[ "$*" == *"org.opencontainers.image.revision"* ]]; then
    printf '%s\n' "${FAKE_IMAGE_REVISION:-unknown}"
  else
    printf '%s\n' 'sha256:fake'
  fi
  exit 0
fi
if [[ "${1:-}" == "run" && "$*" == *"mongo:7.0@sha256:"* ]]; then
  if [[ "$*" == *"--mount"* ]]; then
    printf 'version=7.0.43\nfilesystem=%s\n' "${FAKE_STORAGE_FILESYSTEM:-ext2/ext3}"
  else
    printf '%s\n' '7.0.43'
  fi
  exit 0
fi
if [[ "${1:-}" == "run" && "$*" == *"mysql:8.0@sha256:"* ]]; then
  printf '%s\n' '8.0.46'
  exit 0
fi
if [[ "${1:-}" == "inspect" ]]; then
  [[ "${FAKE_CONTAINER_EXISTS:-true}" == "true" ]] || exit 1
  case "$*" in
    *State.Status*) if [[ -e "$FAKE_STOP_MARKER" ]]; then printf '%s\n' exited; else printf '%s\n' "${FAKE_CONTAINER_STATUS:-running}"; fi ;;
    *State.ExitCode*) printf '%s\n' "${FAKE_EXIT_CODE:-0}" ;;
    *Config.Image*)
      case "$*" in
        *cashlenx-mongodb*) printf '%s\n' 'mongo:7.0@sha256:9854f7139445d766a9523571d6f047530c45547460ffcf8259eb2bf4264632ca' ;;
        *cashlenx-mysql*) printf '%s\n' 'mysql:8.0@sha256:7dcddc01f13bab2f15cde676d44d01f61fc9f99fe7785e86196dfc07d358ae2b' ;;
        *) printf '%s\n' 'cashlenx-server:latest' ;;
      esac ;;
    *'{{.Image}}'*) printf '%s\n' 'sha256:fake' ;;
  esac
  exit 0
fi
if [[ "${1:-}" == "exec" ]]; then [[ "${FAKE_HEALTHY:-true}" == "true" ]]; exit; fi
if [[ "${1:-}" == "stop" ]]; then : > "$FAKE_STOP_MARKER"; exit 0; fi
if [[ "${1:-}" == "logs" ]]; then printf '%s\n' 'fake server log'; exit 0; fi
if [[ "${1:-}" == "run" && "$*" == *"--entrypoint /app/cashlenx-server"* ]]; then
  printf 'CashLenX v%s\nGit Commit: %s\n' "${FAKE_IMAGE_VERSION:-0.0.0}" "${FAKE_IMAGE_REVISION:-unknown}"
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
sed \
  -e 's/^ENV=dev$/ENV=prod/' \
  -e 's|^CORS_ORIGINS=.*$|CORS_ORIGINS=https://app.cashlenx.com|' \
  -e 's/^METRICS_BEARER_TOKEN=$/METRICS_BEARER_TOKEN=lifecycle-metrics-token-with-32-bytes/' \
  "$api_env" > "$prod_env"
printf 'ENV=dev\n' > "$outside_env"

api_env_name="${api_env#"$project_dir/"}"
mysql_env_name="${mysql_env#"$project_dir/"}"
smtp_env_name="${smtp_env#"$project_dir/"}"
invalid_boolean_env_name="${invalid_boolean_env#"$project_dir/"}"
invalid_timezone_env_name="${invalid_timezone_env#"$project_dir/"}"
invalid_storage_env_name="${invalid_storage_env#"$project_dir/"}"
prod_env_name="${prod_env#"$project_dir/"}"
inside_symlink_name="${inside_symlink#"$project_dir/"}"
outside_symlink_name="${outside_symlink#"$project_dir/"}"
ln -s "$(basename "$api_env")" "$inside_symlink"
ln -s "$outside_env" "$outside_symlink"
resolved_api_env="$(realpath "$inside_symlink")"
test_path="$fake_dir:$PATH"
fake_image_version="$(sed -n 's/^const Version = "\([^"]*\)"/\1/p' "$project_dir/model/version.go" | head -n 1)"
fake_image_revision="$(git -C "$project_dir" rev-parse HEAD)"

reset_log() {
  : > "$fake_log"
  rm -f "$fake_stop_marker"
}

run_script() {
  local script="$1"
  local selected_env="${2:-$api_env_name}"
  PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" \
    FAKE_FRONTEND_KIND="${FAKE_FRONTEND_KIND:-docker}" \
    FAKE_VERBOSE_VERSION="${FAKE_VERBOSE_VERSION:-false}" \
    FAKE_SECRET_OUTPUT="${FAKE_SECRET_OUTPUT:-}" \
    FAKE_CONFIG_SUPPORTED="${FAKE_CONFIG_SUPPORTED:-true}" \
    FAKE_NETWORK_EXISTS="${FAKE_NETWORK_EXISTS:-true}" \
    FAKE_NETWORK_CONNECTIONS="${FAKE_NETWORK_CONNECTIONS:-1}" \
    FAKE_CONTAINER_EXISTS="${FAKE_CONTAINER_EXISTS:-true}" \
    FAKE_CONTAINER_STATUS="${FAKE_CONTAINER_STATUS:-running}" \
    FAKE_EXIT_CODE="${FAKE_EXIT_CODE:-0}" \
    FAKE_HEALTHY="${FAKE_HEALTHY:-true}" \
    FAKE_IMAGE_EXISTS="${FAKE_IMAGE_EXISTS:-true}" \
    FAKE_STORAGE_FILESYSTEM="${FAKE_STORAGE_FILESYSTEM:-ext2/ext3}" \
    FAKE_STOP_MARKER="$fake_stop_marker" \
    FAKE_IMAGE_VERSION="$fake_image_version" \
    FAKE_IMAGE_REVISION="$fake_image_revision" \
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
assert_log_not_contains "config --images"

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

reset_log
output="$(FAKE_SECRET_OUTPUT=lifecycle-sensitive-value run_script scripts/dependencies/mongodb/start.sh 2>&1)"
if grep -F -- 'lifecycle-sensitive-value' <<< "$output" >/dev/null; then
  echo "Dependency start output exposed a configured value" >&2
  exit 1
fi
assert_log_contains "exec cashlenx-mongodb sh -ec test"
assert_log_contains "mongosh --quiet"
assert_log_not_contains "--wait"

reset_log
FAKE_FRONTEND_KIND=nerdctl run_script scripts/dependencies/mongodb/start.sh
status_output="$(FAKE_FRONTEND_KIND=nerdctl run_script scripts/dependencies/mongodb/status.sh)"
grep -F 'frontend=nerdctl' <<< "$status_output" >/dev/null
grep -F 'health=healthy' <<< "$status_output" >/dev/null
doctor_output="$(FAKE_FRONTEND_KIND=nerdctl run_script scripts/dependencies/mongodb/doctor.sh)"
grep -F 'diagnostic=doctor' <<< "$doctor_output" >/dev/null
FAKE_FRONTEND_KIND=nerdctl run_script scripts/dependencies/mongodb/logs.sh
assert_log_contains "logs --tail 100 cashlenx-mongodb"
assert_log_contains "--version"
assert_log_contains "--pull never --remove-orphans mongodb"
assert_log_not_contains "config --images"
assert_log_not_contains "--wait"

reset_log
FAKE_FRONTEND_KIND=nerdctl FAKE_VERBOSE_VERSION=true run_script scripts/dependencies/mongodb/start.sh
assert_log_contains "--pull never --remove-orphans mongodb"

reset_log
if output="$(FAKE_FRONTEND_KIND=old-nerdctl run_script scripts/dependencies/mongodb/start.sh 2>&1)"; then
  echo "Expected nerdctl 2.1 to be rejected" >&2
  exit 1
fi
grep -F -- 'Install nerdctl 2.2 or newer' <<< "$output" >/dev/null
assert_log_not_contains "network create"
assert_log_not_contains " up "

reset_log
if output="$(FAKE_CONFIG_SUPPORTED=false run_script scripts/dependencies/mongodb/start.sh 2>&1)"; then
  echo "Expected unsupported Compose configuration to be rejected" >&2
  exit 1
fi
grep -F -- 'cannot validate this Compose configuration' <<< "$output" >/dev/null
assert_log_not_contains "network create"
assert_log_not_contains " up "

reset_log
if [[ -L "$inside_symlink" ]]; then
  run_script scripts/dependencies/mongodb/start.sh "$inside_symlink_name"
  assert_log_contains "--env-file $resolved_api_env"
  assert_log_contains "--remove-orphans mongodb"
fi

reset_log
ENV_FILE=.env.example PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" FAKE_STOP_MARKER="$fake_stop_marker" \
  FAKE_NETWORK_EXISTS=true FAKE_NETWORK_CONNECTIONS=1 \
  bash "$project_dir/scripts/dependencies/mongodb/stop.sh"
assert_log_contains "-f $project_dir/docker/dependencies/mongodb/compose.yml down --remove-orphans"
# Both supported engines reject removal while containers remain attached. The
# lifecycle may safely attempt cleanup without relying on engine-specific
# network-inspect JSON fields.
assert_log_contains "network rm cashlenx-network"

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
status_output="$(run_script scripts/dependencies/mysql/status.sh "$mysql_env_name")"
grep -F 'health=healthy' <<< "$status_output" >/dev/null
doctor_output="$(run_script scripts/dependencies/mysql/doctor.sh "$mysql_env_name")"
grep -F 'diagnostic=doctor' <<< "$doctor_output" >/dev/null
run_script scripts/dependencies/mysql/logs.sh "$mysql_env_name"
assert_log_contains "logs --tail 100 cashlenx-mysql"

reset_log
ENV_FILE=.env.example PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" FAKE_STOP_MARKER="$fake_stop_marker" \
  bash "$project_dir/scripts/dependencies/mysql/stop.sh"
assert_log_contains "-f $project_dir/docker/dependencies/mysql/compose.yml down --remove-orphans"

reset_log
run_script scripts/start.sh
assert_log_contains "-f $project_dir/docker/compose.yml up -d --no-build --pull never --remove-orphans server"
assert_log_contains "exec cashlenx-server sh -ec wget"
assert_log_not_contains "--wait"
status_output="$(run_script scripts/status.sh)"
grep -F 'dependency_state=running' <<< "$status_output" >/dev/null
grep -F 'health=healthy' <<< "$status_output" >/dev/null
doctor_output="$(run_script scripts/doctor.sh)"
grep -F 'diagnostic=doctor' <<< "$doctor_output" >/dev/null
run_script scripts/logs.sh
assert_log_contains "logs --tail 100 cashlenx-server"
if grep -F -- 'dependencies/' "$fake_log" >/dev/null; then
  echo "API start unexpectedly invoked a dependency Compose project" >&2
  exit 1
fi

reset_log
FAKE_NETWORK_EXISTS=true FAKE_NETWORK_CONNECTIONS=0 run_script scripts/stop.sh
assert_log_contains "-f $project_dir/docker/compose.yml down --remove-orphans"
assert_log_contains "network rm cashlenx-network"
assert_log_contains "stop --time 30 cashlenx-server"

reset_log
if output="$(FAKE_HEALTHY=false run_script scripts/status.sh 2>&1)"; then
  echo "Expected degraded API health to fail status" >&2
  exit 1
fi
grep -F 'health=unhealthy' <<< "$output" >/dev/null

reset_log
if output="$(FAKE_EXIT_CODE=137 run_script scripts/stop.sh 2>&1)"; then
  echo "Expected forced API stop to fail the graceful-stop check" >&2
  exit 1
fi
grep -F 'stop_result=forced' <<< "$output" >/dev/null
assert_log_contains "down --remove-orphans"

reset_log
if output="$(FAKE_IMAGE_EXISTS=false run_script scripts/status.sh 2>&1)"; then
  echo "Expected missing API image to fail status" >&2
  exit 1
fi
grep -F 'image=missing' <<< "$output" >/dev/null

reset_log
if output="$(FAKE_NETWORK_EXISTS=false run_script scripts/status.sh 2>&1)"; then
  echo "Expected missing API network to fail status" >&2
  exit 1
fi
grep -F 'network_state=missing' <<< "$output" >/dev/null

reset_log
if output="$(FAKE_CONTAINER_EXISTS=false run_script scripts/status.sh 2>&1)"; then
  echo "Expected missing API dependency to fail status" >&2
  exit 1
fi
grep -F 'dependency_state=missing' <<< "$output" >/dev/null

reset_log
FAKE_NETWORK_EXISTS=true FAKE_NETWORK_CONNECTIONS=0 run_script scripts/stop.sh
output="$(FAKE_NETWORK_EXISTS=true FAKE_NETWORK_CONNECTIONS=0 run_script scripts/stop.sh)"
grep -F 'stop_result=already-stopped' <<< "$output" >/dev/null
assert_log_contains "down --remove-orphans"

reset_log
run_script scripts/start.sh "$mysql_env_name"
assert_log_contains "-f $project_dir/docker/compose.yml up -d --no-build --pull never --remove-orphans server"
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

reset_log
run_script scripts/start.sh "$prod_env_name"
assert_log_contains "-f $project_dir/docker/compose.yml up -d --no-build --pull never --remove-orphans server"

sed 's/^METRICS_BEARER_TOKEN=.*$/METRICS_BEARER_TOKEN=short/' \
  "$prod_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/start.sh METRICS_BEARER_TOKEN short

sed 's|^CORS_ORIGINS=.*$|CORS_ORIGINS=http://localhost:*|' \
  "$prod_env" > "$invalid_storage_env"
reset_log
assert_rejected "$invalid_storage_env_name" scripts/start.sh CORS_ORIGINS 'http://localhost:*'

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
assert_log_contains "exec cashlenx-mongodb sh -ec test"
assert_log_contains "mongosh --quiet"
assert_log_not_contains "--wait"

for native_filesystem in ext4 xfs; do
  reset_log
  FAKE_STORAGE_FILESYSTEM="$native_filesystem" run_script scripts/dependencies/mongodb/start.sh "$invalid_storage_env_name"
  assert_log_contains "--remove-orphans mongodb"
done

reset_log
if output="$(FAKE_STORAGE_FILESYSTEM=v9fs run_script scripts/dependencies/mongodb/start.sh "$invalid_storage_env_name" 2>&1)"; then
  echo "Expected MongoDB to reject a shared filesystem before initialization" >&2
  exit 1
fi
grep -F -- '/srv/cashlenx/mongodb' <<< "$output" >/dev/null
grep -F -- "unsupported shared or remote filesystem 'v9fs'" <<< "$output" >/dev/null
grep -F -- 'Leave MONGO_DATA_PATH empty' <<< "$output" >/dev/null
grep -F -- 'No data was changed' <<< "$output" >/dev/null
assert_log_not_contains " up "
assert_log_not_contains "network create"

reset_log
if output="$(FAKE_STORAGE_FILESYSTEM=overlay run_script scripts/dependencies/mongodb/start.sh "$invalid_storage_env_name" 2>&1)"; then
  echo "Expected MongoDB to fail closed on an unapproved filesystem" >&2
  exit 1
fi
grep -F -- "unapproved filesystem 'overlay'" <<< "$output" >/dev/null
assert_log_not_contains " up "

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
if [[ -L "$outside_symlink" ]]; then
  assert_rejected "$outside_symlink_name" scripts/dependencies/mongodb/build.sh "ENV_FILE must stay inside"
fi

sed 's|^MONGO_IMAGE=.*$|MONGO_IMAGE=mongo:7.0|' \
  docker/dependencies/images.env > "$invalid_images"
if output="$({
  project_dir="$project_dir"
  . "$project_dir/scripts/lib/container_lifecycle.sh"
  . "$project_dir/scripts/dependencies/image_pins.sh"
  export dependency_images_file="$invalid_images"
  load_dependency_image_pin mongodb
} 2>&1)"; then
  echo "Expected a mutable MongoDB image pin to be rejected" >&2
  exit 1
fi
grep -F -- 'MONGO_IMAGE must be a digest-pinned' <<< "$output" >/dev/null

echo "Dependency lifecycle smoke checks passed."
