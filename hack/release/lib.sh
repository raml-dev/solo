# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

repo_root() {
  cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd
}

RELEASE_REPO_ROOT="$(repo_root)"
RELEASE_LINUX_DESKTOP_ID="dev.raml.solo"
RELEASE_LINUX_DESKTOP_FILE_NAME="${RELEASE_LINUX_DESKTOP_ID}.desktop"
RELEASE_LINUX_METAINFO_FILE_NAME="${RELEASE_LINUX_DESKTOP_ID}.metainfo.xml"
RELEASE_LINUX_DESKTOP_FILE_SOURCE="${RELEASE_REPO_ROOT}/build/linux/${RELEASE_LINUX_DESKTOP_FILE_NAME}"
RELEASE_LINUX_METAINFO_SOURCE="${RELEASE_REPO_ROOT}/build/linux/${RELEASE_LINUX_METAINFO_FILE_NAME}"
RELEASE_LINUX_ICON_SOURCE="${RELEASE_REPO_ROOT}/build/linux/icon.png"
RELEASE_LINUX_NFPM_TEMPLATE_PATH="${RELEASE_REPO_ROOT}/build/linux/nfpm.yaml"
RELEASE_MACOS_BUNDLE_NAME="Solo.app"
RELEASE_MACOS_EXECUTABLE_NAME="Solo"
RELEASE_MACOS_VOLUME_NAME="Solo"
RELEASE_MACOS_BUNDLE_PATH="${RELEASE_REPO_ROOT}/build/bin/${RELEASE_MACOS_BUNDLE_NAME}"
RELEASE_DIST_DIR="${SOLO_DIST_DIR:-${RELEASE_REPO_ROOT}/dist}"
RELEASE_TMP_DIR="${RELEASE_DIST_DIR}/.tmp"
RELEASE_REPO_WORK_DIR="${RELEASE_DIST_DIR}/repos"

log() {
  printf '[release] %s\n' "$*" >&2
}

die() {
  log "ERROR: $*"
  exit 1
}

require_env() {
  local name="$1"
  if [[ -z "${!name:-}" ]]; then
    die "missing required environment variable: ${name}"
  fi
}

require_release_context() {
  require_env "SOLO_RELEASE_TAG"
  require_env "SOLO_RELEASE_VERSION"
  require_env "SOLO_RELEASE_CHANNEL"
  validate_release_channel "${SOLO_RELEASE_CHANNEL}"
}

validate_release_channel() {
  case "$1" in
    stable|prerelease) ;;
    *)
      die "invalid SOLO_RELEASE_CHANNEL: $1"
      ;;
  esac
}

raw_targets() {
  printf '%s\n' \
    windows-amd64 \
    windows-arm64 \
    linux-amd64 \
    linux-arm64 \
    darwin-amd64 \
    darwin-arm64
}

package_targets() {
  printf '%s\n' \
    deb-amd64 \
    deb-arm64 \
    rpm-amd64 \
    rpm-arm64 \
    arch-amd64 \
    arch-arm64
}

validate_raw_target() {
  local target="$1"
  if ! raw_targets | grep -Fx -- "$target" >/dev/null 2>&1; then
    die "unknown raw build target: ${target}"
  fi
}

raw_artifact_name() {
  case "$1" in
    windows-amd64) printf '%s\n' solo-windows-amd64.exe ;;
    windows-arm64) printf '%s\n' solo-windows-arm64.exe ;;
    linux-amd64) printf '%s\n' solo-linux-amd64 ;;
    linux-arm64) printf '%s\n' solo-linux-arm64 ;;
    darwin-amd64) printf '%s\n' solo-darwin-amd64.dmg ;;
    darwin-arm64) printf '%s\n' solo-darwin-arm64.dmg ;;
    *) die "cannot infer raw artifact name for target: $1" ;;
  esac
}

hash_command() {
  if command -v sha256sum >/dev/null 2>&1; then
    printf '%s\n' sha256sum
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    printf '%s\n' "shasum -a 256"
    return
  fi
  die "no SHA256 command available (expected sha256sum or shasum)"
}

hash_file() {
  local file="$1"
  local cmd
  cmd="$(hash_command)"
  if [[ "${cmd}" == "sha256sum" ]]; then
    sha256sum "$file" | awk '{print $1}'
    return
  fi
  shasum -a 256 "$file" | awk '{print $1}'
}

list_top_level_files() {
  local dist_dir
  dist_dir="${RELEASE_DIST_DIR}"
  if [[ ! -d "${dist_dir}" ]]; then
    return 0
  fi
  find "${dist_dir}" -mindepth 1 -maxdepth 1 -type f | LC_ALL=C sort
}
