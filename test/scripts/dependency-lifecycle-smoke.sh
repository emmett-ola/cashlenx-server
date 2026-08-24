#!/usr/bin/env bash
set -euo pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"
fake_dir="$(mktemp -d)"
fake_log="$fake_dir/docker.log"
test_env="$(mktemp "$project_dir/.env.lifecycle-test.XXXXXX")"
outside_env="$(mktemp)"

cleanup() {
  rm -f "$test_env" "$outside_env"
  rm -rf "$fake_dir"
}
trap cleanup EXIT

cat > "$fake_dir/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$*" >> "$FAKE_DOCKER_LOG"
EOF
chmod +x "$fake_dir/docker"

sed \
  -e 's/CHANGE_ME_MONGO_PASSWORD/lifecycle-mongo-password/g' \
  -e 's/CHANGE_ME_JWT_SECRET/lifecycle-jwt-secret-with-more-than-32-bytes/g' \
  -e 's/CHANGE_ME_ADMIN_PASSWORD/lifecycle-admin-password/g' \
  "$project_dir/.env.example" > "$test_env"
cat >> "$test_env" <<'EOF'
MYSQL_ROOT_PASSWORD=lifecycle-mysql-root-password
MYSQL_USER=cashlenx
MYSQL_PASSWORD=lifecycle-mysql-password
MYSQL_DB_URI=${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(localhost:3306)
EOF
printf 'ENV=dev\n' > "$outside_env"

env_name="${test_env#"$project_dir/"}"
test_path="$fake_dir:$PATH"

reset_log() {
  : > "$fake_log"
}

run_script() {
  local script="$1"
  PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" ENV_FILE="$env_name" \
    bash "$project_dir/$script"
}

assert_log_contains() {
  local expected="$1"
  grep -F -- "$expected" "$fake_log" >/dev/null || {
    echo "Expected fake Docker call containing: $expected" >&2
    return 1
  }
}

assert_rejected() {
  local env_file="$1"
  local script="$2"
  local expected_key="$3"
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
run_script scripts/dependencies/mongodb/start.sh
assert_log_contains "--pull never --remove-orphans --wait mongodb"

reset_log
ENV_FILE=.env.example PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" \
  bash "$project_dir/scripts/dependencies/mongodb/stop.sh"
assert_log_contains "-f $project_dir/docker/dependencies/mongodb/compose.yml down --remove-orphans"

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
run_script scripts/dependencies/mysql/start.sh
assert_log_contains "--pull never --remove-orphans --wait mysql"

reset_log
ENV_FILE=.env.example PATH="$test_path" FAKE_DOCKER_LOG="$fake_log" \
  bash "$project_dir/scripts/dependencies/mysql/stop.sh"
assert_log_contains "-f $project_dir/docker/dependencies/mysql/compose.yml down --remove-orphans"

reset_log
run_script scripts/start.sh
assert_log_contains "-f $project_dir/docker/compose.yml up -d --no-build --remove-orphans --wait server"
if grep -F -- 'dependencies/' "$fake_log" >/dev/null; then
  echo "API start unexpectedly invoked a dependency Compose project" >&2
  exit 1
fi

reset_log
assert_rejected .env.lifecycle-missing scripts/dependencies/mongodb/build.sh "Missing environment file"
assert_rejected "$outside_env" scripts/dependencies/mongodb/build.sh "ENV_FILE must stay inside"

echo "Dependency lifecycle smoke checks passed."
