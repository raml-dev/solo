# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

validate_context() {
  require_release_context
  log "validated release context for ${SOLO_RELEASE_TAG} (${SOLO_RELEASE_CHANNEL})"
}

prepare_dist() {
  require_release_context
  local dist_dir tmp_dir repos_dir
  dist_dir="${RELEASE_DIST_DIR}"
  tmp_dir="${RELEASE_TMP_DIR}"
  repos_dir="${RELEASE_REPO_WORK_DIR}"

  rm -rf "${dist_dir}"
  mkdir -p "${dist_dir}" "${tmp_dir}" "${repos_dir}"
  log "prepared dist directory at ${dist_dir}"
}

inject_version() {
  require_release_context
  local config_path temp_path
  config_path="${RELEASE_REPO_ROOT}/wails.json"
  temp_path="$(mktemp)"
  jq --arg version "${SOLO_RELEASE_VERSION}" \
    '.info.productVersion = $version' \
    "${config_path}" > "${temp_path}"
  mv "${temp_path}" "${config_path}"
  log "injected SOLO_RELEASE_VERSION=${SOLO_RELEASE_VERSION} into wails.json"
}

main() {
  local command="${1:-}"
  case "${command}" in
    validate-context)
      validate_context
      ;;
    prepare-dist)
      prepare_dist
      ;;
    inject-version)
      inject_version
      ;;
    *)
      die "unknown prepare subcommand: ${command:-<empty>}"
      ;;
  esac
}

main "$@"
