#!/usr/bin/env bash

if [[ -n "${CASHLENX_CONTAINER_LIFECYCLE_LOADED:-}" ]]; then
  return 0
fi
CASHLENX_CONTAINER_LIFECYCLE_LOADED=1
project_dir="${project_dir:?project_dir must be set before sourcing container_lifecycle.sh}"

lifecycle_error() {
  printf 'Container lifecycle error: %s\n' "$*" >&2
}

resolve_env_file() {
  local requested="${ENV_FILE:-.env}"
  local candidate
  if [[ "$requested" == /* ]]; then
    candidate="$requested"
  else
    candidate="$project_dir/$requested"
  fi

  if [[ ! -e "$candidate" ]]; then
    lifecycle_error "Missing environment file: $requested"
    printf 'Create it with: cp .env.example "%s"\n' "$requested" >&2
    return 1
  fi
  [[ -f "$candidate" ]] || { lifecycle_error "Environment path is not a file: $requested"; return 1; }

  local resolved
  resolved="$(realpath "$candidate")"
  case "$resolved" in
    "$project_dir"/*) printf '%s\n' "$resolved" ;;
    *) lifecycle_error "ENV_FILE must stay inside $project_dir: $requested"; return 1 ;;
  esac
}

read_env_value() {
  local key="$1"
  local source_file="${2:-${env_file:?env_file is not set}}"
  [[ "$key" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || { lifecycle_error "Invalid environment key: $key"; return 1; }
  awk -F= -v wanted="$key" '
    $0 ~ "^[[:space:]]*(export[[:space:]]+)?" wanted "[[:space:]]*=" {
      value = substr($0, index($0, "=") + 1)
      gsub(/^[[:space:]]+|[[:space:]]+$/, "", value)
      if ((substr(value, 1, 1) == "\"" && substr(value, length(value), 1) == "\"") ||
          (substr(value, 1, 1) == "\047" && substr(value, length(value), 1) == "\047")) {
        value = substr(value, 2, length(value) - 2)
      }
      result = value
    }
    END { print result }
  ' "$source_file"
}

read_config_value() {
  local key="$1"
  local fallback="${2:-}"
  local source_file="${3:-${env_file:?env_file is not set}}"
  local value
  if [[ -n "${!key+x}" ]]; then
    value="${!key}"
  else
    value="$(read_env_value "$key" "$source_file")"
  fi
  printf '%s\n' "${value:-$fallback}"
}

load_env_defaults() {
  local source_file="$1"
  shift
  local key value
  for key in "$@"; do
    if [[ -z "${!key+x}" ]]; then
      value="$(read_env_value "$key" "$source_file")"
      [[ -n "$value" ]] || { lifecycle_error "$key is required from ${source_file#"$project_dir"/}"; return 1; }
      printf -v "$key" '%s' "$value"
      export "${key?}"
    fi
  done
}

container_runtime_init() {
  local requested="${1:-auto}"
  local cli version_output compose_output combined version major minor
  case "$requested" in
    auto | docker | nerdctl) ;;
    *) lifecycle_error "CONTAINER_FRONTEND must be auto, docker, or nerdctl."; return 1 ;;
  esac

  if [[ -n "${CONTAINER_CLI:-}" ]]; then
    cli="$CONTAINER_CLI"
  elif [[ "$requested" == "nerdctl" ]]; then
    cli="nerdctl"
  elif [[ "$requested" == "docker" ]] || command -v docker >/dev/null 2>&1; then
    cli="docker"
  else
    cli="nerdctl"
  fi

  command -v "$cli" >/dev/null 2>&1 || {
    lifecycle_error "Container command '$cli' was not found. Install Docker Compose v2 or nerdctl 2.2+, or set CONTAINER_CLI."
    return 1
  }
  version_output="$("$cli" --version 2>&1)" || {
    lifecycle_error "Container command '$cli' cannot report its version. Install Docker Compose v2 or nerdctl 2.2+."
    return 1
  }
  compose_output="$("$cli" compose version 2>&1)" || {
    lifecycle_error "Container command '$cli' has no usable Compose frontend. Install Docker Compose v2 or nerdctl 2.2+."
    return 1
  }
  combined="$(printf '%s\n%s\n' "$version_output" "$compose_output" | tr '[:upper:]' '[:lower:]')"

  if [[ "$combined" == *nerdctl* ]]; then
    CONTAINER_FRONTEND_KIND="nerdctl"
    version="$(printf '%s\n' "$combined" | sed -nE 's/.*nerdctl( compose)? version v?([0-9]+\.[0-9]+(\.[0-9]+)?).*/\2/p' | sed -n '1p')"
    [[ -n "$version" ]] || { lifecycle_error "Could not determine the nerdctl version reported by '$cli'."; return 1; }
    major="${version%%.*}"
    minor="${version#*.}"
    minor="${minor%%.*}"
    if ((10#$major < 2 || (10#$major == 2 && 10#$minor < 2))); then
      lifecycle_error "nerdctl $version is unsupported. Install nerdctl 2.2 or newer before retrying."
      return 1
    fi
  elif [[ "$combined" == *docker* && "$compose_output" == *"Compose"* ]]; then
    CONTAINER_FRONTEND_KIND="docker"
    version="$(printf '%s\n' "$compose_output" | sed -nE 's/.*version[[:space:]]+v?([0-9]+\.[0-9]+(\.[0-9]+)?).*/\1/p' | sed -n '1p')"
  else
    lifecycle_error "Unknown container frontend reported by '$cli'. Install Docker Compose v2 or nerdctl 2.2+."
    return 1
  fi

  if [[ "$requested" != "auto" && "$requested" != "$CONTAINER_FRONTEND_KIND" ]]; then
    lifecycle_error "CONTAINER_FRONTEND requested '$requested', but '$cli' reports '$CONTAINER_FRONTEND_KIND'."
    return 1
  fi
  "$cli" info >/dev/null 2>&1 || {
    lifecycle_error "The $CONTAINER_FRONTEND_KIND runtime is unavailable. Start its daemon before retrying."
    return 1
  }

  CONTAINER_CLI="$cli"
  CONTAINER_FRONTEND_VERSION="${version:-unknown}"
  export CONTAINER_CLI CONTAINER_FRONTEND_KIND CONTAINER_FRONTEND_VERSION
}

container() {
  "${CONTAINER_CLI:?container_runtime_init must run first}" "$@"
}

compose() {
  container compose "$@"
}

compose_preflight() {
  if ! compose "$@" config --quiet >/dev/null 2>&1; then
    lifecycle_error "The selected $CONTAINER_FRONTEND_KIND frontend cannot validate this Compose configuration. Check the environment file and use Docker Compose v2 or nerdctl 2.2+."
    return 1
  fi
}

compose_up_quiet() {
  compose "$@" >/dev/null 2>&1 || {
    lifecycle_error "Compose could not start the requested services. Inspect the selected runtime without printing environment values, then retry."
    return 1
  }
}

resolve_image_ref() {
  local name_key="$1"
  local default_name="$2"
  local tag_key="$3"
  local default_tag="$4"
  local name tag
  name="$(read_config_value "$name_key" "$default_name")"
  tag="$(read_config_value "$tag_key" "$default_tag")"
  [[ "$name" =~ ^[a-z0-9][a-z0-9._/:@-]*$ && "$name" != *@* ]] || {
    lifecycle_error "$name_key must be a lowercase container image repository without a digest."
    return 1
  }
  [[ "$tag" =~ ^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$ ]] || {
    lifecycle_error "$tag_key is not a valid container image tag."
    return 1
  }
  printf '%s:%s\n' "$name" "$tag"
}

resolve_network_name() {
  local name
  name="$(read_config_value DOCKER_NETWORK_NAME cashlenx-network)"
  [[ "$name" =~ ^[A-Za-z0-9][A-Za-z0-9_.-]*$ ]] || {
    lifecycle_error "Invalid container network setting: DOCKER_NETWORK_NAME"
    return 1
  }
  printf '%s\n' "$name"
}

ensure_network() {
  local name="$1"
  if container network inspect "$name" >/dev/null 2>&1; then
    return 0
  fi
  container network create --driver bridge "$name" >/dev/null 2>&1 ||
    container network inspect "$name" >/dev/null 2>&1
}

remove_network_if_unused() {
  local name="$1"
  container network rm "$name" >/dev/null 2>&1 || true
}

wait_for_container_command() {
  local container_name="$1"
  shift
  local timeout_seconds="${CONTAINER_READY_TIMEOUT_SECONDS:-600}"
  local deadline=$((SECONDS + timeout_seconds))
  local status

  while ((SECONDS < deadline)); do
    status="$(container inspect --format '{{.State.Status}}' "$container_name" 2>/dev/null || true)"
    case "$status" in
      exited | dead)
        lifecycle_error "Container stopped before becoming ready: $container_name"
        container logs --tail 50 "$container_name" >&2 || true
        return 1
        ;;
    esac
    if container exec "$container_name" "$@" >/dev/null 2>&1; then
      printf 'Container is ready: %s\n' "$container_name"
      return 0
    fi
    sleep 2
  done

  lifecycle_error "Timed out after ${timeout_seconds}s waiting for container: $container_name"
  container logs --tail 50 "$container_name" >&2 || true
  return 1
}

save_image() {
  local output_file="$1"
  local image_ref="$2"
  if [[ "$CONTAINER_FRONTEND_KIND" == "nerdctl" ]]; then
    container save --output "$output_file" "$image_ref"
  else
    container image save --output "$output_file" "$image_ref"
  fi
}
