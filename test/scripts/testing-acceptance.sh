#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-https://test-api.cashlenx.com/api/v1}"
RUN_ID="$(date +%s)-${RANDOM}"
CATEGORY_NAME="acceptance-expense-${RUN_ID}"
CATEGORY_EMOJI="🧪"
CATEGORY_BG_COLOR="#5B8FF9"
TMP_DIR="$(mktemp -d)"
RESP_FILE="${TMP_DIR}/response.json"
REQUEST_FILE="${TMP_DIR}/request.json"
AUTH_HEADER_FILE="${TMP_DIR}/authorization.header"
CATEGORY_ID=""

chmod 700 "$TMP_DIR"
touch "$RESP_FILE" "$REQUEST_FILE" "$AUTH_HEADER_FILE"
chmod 600 "$RESP_FILE" "$REQUEST_FILE" "$AUTH_HEADER_FILE"

usage() {
  cat <<'EOF'
Usage: test/scripts/testing-acceptance.sh

Runs the authenticated Testing acceptance checks against:
  https://test-api.cashlenx.com/api/v1

Credentials are never accepted as command arguments or environment values.
By default the script prompts for a username and silently prompts for a
password. For automation, provide two newline-delimited values on a dedicated
file descriptor:

  ACCEPTANCE_CREDENTIALS_FD=3 test/scripts/testing-acceptance.sh 3<credentials.txt

The credentials file must remain outside Git and should be readable only by
its owner. Set BASE_URL only to select another path on the canonical Testing
host. Set ACCEPTANCE_ALLOW_OTHER_HOST=true for an explicitly approved target.
EOF
}

for arg in "$@"; do
  case "$arg" in
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $arg" >&2
      usage >&2
      exit 1
      ;;
  esac
done

case "$BASE_URL" in
  https://test-api.cashlenx.com/api/v1|https://test-api.cashlenx.com/api/v1/)
    BASE_URL="${BASE_URL%/}"
    ;;
  *)
    if [[ "${ACCEPTANCE_ALLOW_OTHER_HOST:-false}" != "true" ]]; then
      echo "refusing non-canonical Testing target: $BASE_URL" >&2
      echo "set ACCEPTANCE_ALLOW_OTHER_HOST=true only for an explicitly approved target" >&2
      exit 1
    fi
    BASE_URL="${BASE_URL%/}"
    ;;
esac

for command_name in curl python3; do
  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "missing required command: $command_name" >&2
    exit 1
  fi
done

read_credentials() {
  if [[ -n "${ACCEPTANCE_CREDENTIALS_FD:-}" ]]; then
    if [[ ! "$ACCEPTANCE_CREDENTIALS_FD" =~ ^[0-9]+$ ]]; then
      echo "ACCEPTANCE_CREDENTIALS_FD must be a numeric file descriptor" >&2
      exit 1
    fi
    IFS= read -r ACCEPTANCE_USERNAME <&"$ACCEPTANCE_CREDENTIALS_FD"
    IFS= read -r ACCEPTANCE_PASSWORD <&"$ACCEPTANCE_CREDENTIALS_FD"
  elif [[ -t 0 && -t 1 ]]; then
    read -r -p "Testing acceptance username: " ACCEPTANCE_USERNAME
    read -r -s -p "Testing acceptance password: " ACCEPTANCE_PASSWORD
    printf '\n'
  else
    echo "credentials require an interactive terminal or ACCEPTANCE_CREDENTIALS_FD" >&2
    exit 1
  fi

  if [[ -z "${ACCEPTANCE_USERNAME:-}" || -z "${ACCEPTANCE_PASSWORD:-}" ]]; then
    echo "username and password must both be non-empty" >&2
    exit 1
  fi
}

json_value() {
  local path="$1"
  python3 - "$RESP_FILE" "$path" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as handle:
    value = json.load(handle)

for part in sys.argv[2].split("."):
    if not isinstance(value, dict) or part not in value:
        sys.exit(1)
    value = value[part]

if isinstance(value, (dict, list)):
    print(json.dumps(value, ensure_ascii=False))
elif value is not None:
    print(value)
PY
}

write_json() {
  python3 - "$REQUEST_FILE" "$@" <<'PY'
import json
import sys

target = sys.argv[1]
payload = {}
for pair in sys.argv[2:]:
    key, value = pair.split("=", 1)
    payload[key] = value

with open(target, "w", encoding="utf-8") as handle:
    json.dump(payload, handle, ensure_ascii=False)
PY
}

write_login_json() {
  printf '%s\0%s\0' "$ACCEPTANCE_USERNAME" "$ACCEPTANCE_PASSWORD" |
    python3 -c '
import json
import sys

values = sys.stdin.buffer.read().split(b"\0")
if len(values) < 3:
    raise SystemExit("invalid credential input")

payload = {
    "username": values[0].decode("utf-8"),
    "password": values[1].decode("utf-8"),
    "device_id": "testing-acceptance",
    "device_name": "Testing Acceptance",
}
with open(sys.argv[1], "w", encoding="utf-8") as handle:
    json.dump(payload, handle, ensure_ascii=False)
' "$REQUEST_FILE"
}

request() {
  local method="$1"
  local path="$2"
  local expected="$3"
  local authenticated="${4:-false}"
  local body_file="${5:-}"
  local status
  local args=(
    --silent --show-error
    --request "$method"
    --output "$RESP_FILE"
    --write-out '%{http_code}'
    --header "X-Request-ID: testing-acceptance-${RUN_ID}"
  )

  if [[ "$authenticated" == "true" ]]; then
    args+=(--header "@${AUTH_HEADER_FILE}")
  fi
  if [[ -n "$body_file" ]]; then
    args+=(--header "Content-Type: application/json" --data-binary "@${body_file}")
  fi

  status="$(curl "${args[@]}" "${BASE_URL}${path}")"
  if [[ "$status" != "$expected" ]]; then
    echo "FAILED ${method} ${path}: HTTP ${status}, expected ${expected}" >&2
    return 1
  fi
  echo "OK ${method} ${path} -> ${status}"
}

cleanup() {
  local exit_status=$?
  trap - EXIT
  if [[ -s "$AUTH_HEADER_FILE" && -n "$CATEGORY_ID" ]]; then
    if ! request DELETE "/category/${CATEGORY_ID}" 200 true >/dev/null; then
      echo "warning: failed to remove acceptance category ${CATEGORY_ID}" >&2
      exit_status=1
    else
      echo "OK cleanup category ${CATEGORY_ID}"
    fi
  fi
  rm -rf "$TMP_DIR"
  unset ACCEPTANCE_PASSWORD
  exit "$exit_status"
}
trap cleanup EXIT

read_credentials

write_login_json
request POST "/open/auth/login" 200 false "$REQUEST_FILE"
access_token="$(json_value data.access_token)"
if [[ -z "$access_token" ]]; then
  echo "login response did not include an access token" >&2
  exit 1
fi
printf 'Authorization: Bearer %s\n' "$access_token" >"$AUTH_HEADER_FILE"
unset access_token
unset ACCEPTANCE_PASSWORD

write_json \
  "name=${CATEGORY_NAME}" \
  "type=expense" \
  "remark=temporary Testing acceptance data" \
  "emoji=${CATEGORY_EMOJI}" \
  "bg_color=${CATEGORY_BG_COLOR}"
request POST "/category" 201 true "$REQUEST_FILE"
CATEGORY_ID="$(json_value data.id 2>/dev/null || json_value data.Id)"
[[ -n "$CATEGORY_ID" ]] || {
  echo "category response did not include an id" >&2
  exit 1
}

if [[ "$(json_value data.emoji)" != "$CATEGORY_EMOJI" ]]; then
  echo "created category did not retain emoji" >&2
  exit 1
fi
if [[ "$(json_value data.bg_color)" != "$CATEGORY_BG_COLOR" ]]; then
  echo "created category did not retain bg_color" >&2
  exit 1
fi

request GET "/category/${CATEGORY_ID}" 200 true
if [[ "$(json_value data.emoji)" != "$CATEGORY_EMOJI" || "$(json_value data.bg_color)" != "$CATEGORY_BG_COLOR" ]]; then
  echo "read category did not retain presentation fields" >&2
  exit 1
fi

request GET "/statistic/chart/monthly-comparison/$(date +%Y)" 200 true

request DELETE "/category/${CATEGORY_ID}" 200 true
CATEGORY_ID=""

echo "Testing authenticated acceptance completed successfully."
