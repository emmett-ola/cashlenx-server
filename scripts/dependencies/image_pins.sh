#!/usr/bin/env bash

if [[ -n "${CASHLENX_DEPENDENCY_IMAGE_PINS_LOADED:-}" ]]; then
  return 0
fi
CASHLENX_DEPENDENCY_IMAGE_PINS_LOADED=1

dependency_images_file="${project_dir:?project_dir must be set before sourcing image_pins.sh}/docker/dependencies/images.env"

load_dependency_image_pin() {
  local dependency="$1" image_key version_key expected_image version_pattern image version
  case "$dependency" in
    mongodb)
      image_key="MONGO_IMAGE"
      version_key="MONGO_VERSION"
      expected_image="mongo:7.0"
      version_pattern='^7\.0\.[0-9]+$'
      ;;
    mysql)
      image_key="MYSQL_IMAGE"
      version_key="MYSQL_VERSION"
      expected_image="mysql:8.0"
      version_pattern='^8\.0\.[0-9]+$'
      ;;
    *) lifecycle_error "Unknown database dependency: $dependency"; return 1 ;;
  esac

  image="$(read_env_value "$image_key" "$dependency_images_file")"
  version="$(read_env_value "$version_key" "$dependency_images_file")"
  [[ "$image" =~ ^${expected_image}@sha256:[0-9a-f]{64}$ ]] || {
    lifecycle_error "$image_key must be a digest-pinned $expected_image reference in docker/dependencies/images.env."
    return 1
  }
  [[ "$version" =~ $version_pattern ]] || {
    lifecycle_error "$version_key must identify the exact $expected_image patch version in docker/dependencies/images.env."
    return 1
  }
  printf -v "$image_key" '%s' "$image"
  printf -v "$version_key" '%s' "$version"
  export "${image_key?}" "${version_key?}"
}

verify_dependency_image_version() {
  local dependency="$1" output actual expected image
  case "$dependency" in
    mongodb)
      image="$MONGO_IMAGE"
      expected="$MONGO_VERSION"
      output="$(container run --rm --pull never --network none --entrypoint sh "$image" -ec \
        'mongod --version | sed -n "s/^db version v//p" | sed -n "1p"')" || {
          lifecycle_error "Could not verify the pinned MongoDB image without pulling a different image."
          return 1
        }
      ;;
    mysql)
      image="$MYSQL_IMAGE"
      expected="$MYSQL_VERSION"
      output="$(container run --rm --pull never --network none --entrypoint sh "$image" -ec \
        'mysqld --version | sed -n "s/.* Ver \([0-9][^ ]*\).*/\1/p"')" || {
          lifecycle_error "Could not verify the pinned MySQL image without pulling a different image."
          return 1
        }
      ;;
    *) lifecycle_error "Unknown database dependency: $dependency"; return 1 ;;
  esac
  actual="$(printf '%s\n' "$output" | sed -n '1p' | tr -d '\r')"
  [[ "$actual" == "$expected" ]] || {
    lifecycle_error "The pinned $dependency image reports version '${actual:-unknown}', expected '$expected'."
    return 1
  }
  printf '%s_image_version=%s\n' "$dependency" "$actual"
}

preflight_mongodb_storage() {
  local data_path volume_name mount_spec storage_source probe actual_version filesystem
  data_path="$(read_config_value MONGO_DATA_PATH)"
  volume_name="$(read_config_value MONGO_DATA_VOLUME_NAME cashlenx-mongodb-data)"
  if [[ -n "$data_path" ]]; then
    mount_spec="type=bind,source=$data_path,target=/data/db,readonly"
    storage_source="$data_path"
  else
    mount_spec="type=volume,source=$volume_name,target=/data/db,readonly"
    storage_source="named volume $volume_name"
  fi

  # Git Bash rewrites container-internal POSIX paths for native Windows
  # executables unless conversion is disabled for this one Docker invocation.
  probe="$(MSYS_NO_PATHCONV=1 container run --rm --pull never --network none --entrypoint sh \
    --mount "$mount_spec" "$MONGO_IMAGE" -ec '
      version="$(mongod --version | sed -n "s/^db version v//p" | sed -n "1p")"
      filesystem="$(stat -f -c %T /data/db)"
      printf "version=%s\nfilesystem=%s\n" "$version" "$filesystem"
    ')" || {
      lifecycle_error "Could not inspect MongoDB data source '$storage_source' from the pinned container image. Confirm that the source exists and is accessible."
      return 1
    }
  actual_version="$(printf '%s\n' "$probe" | sed -n 's/^version=//p' | sed -n '1p' | tr -d '\r')"
  filesystem="$(printf '%s\n' "$probe" | sed -n 's/^filesystem=//p' | sed -n '1p' | tr '[:upper:]' '[:lower:]' | tr -d '\r')"
  [[ "$actual_version" == "$MONGO_VERSION" ]] || {
    lifecycle_error "The pinned MongoDB image reports version '${actual_version:-unknown}', expected '$MONGO_VERSION'."
    return 1
  }

  case "$filesystem" in
    ext2/ext3 | ext4 | xfs | btrfs | zfs)
      printf 'mongodb_image_version=%s\nmongodb_storage_filesystem=%s\n' "$actual_version" "$filesystem"
      ;;
    9p | v9fs | drvfs | cifs | smbfs | nfs | nfs4 | fuse | fuse.* | fuseblk | virtiofs | afs | ceph | glusterfs)
      lifecycle_error "MongoDB data source '$storage_source' resolves inside the container as unsupported shared or remote filesystem '$filesystem'. Leave MONGO_DATA_PATH empty to use the named volume, or move retained data through a reviewed migration to native ext4 or XFS storage. No data was changed."
      return 1
      ;;
    *)
      lifecycle_error "MongoDB data source '$storage_source' resolves inside the container as unapproved filesystem '${filesystem:-unknown}'. Supported native filesystems include ext4 and XFS. Leave MONGO_DATA_PATH empty to use the named volume, or review the storage platform before retrying. No data was changed."
      return 1
      ;;
  esac
}
