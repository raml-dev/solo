# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

package_artifact_names() {
  printf '%s\n' \
    "solo_${SOLO_RELEASE_VERSION}_amd64.deb" \
    "solo_${SOLO_RELEASE_VERSION}_arm64.deb" \
    "solo-${SOLO_RELEASE_VERSION}-1.x86_64.rpm" \
    "solo-${SOLO_RELEASE_VERSION}-1.aarch64.rpm" \
    "solo-${SOLO_RELEASE_VERSION}-1-x86_64.pkg.tar.zst" \
    "solo-${SOLO_RELEASE_VERSION}-1-aarch64.pkg.tar.zst"
}

raw_dispatch_artifact_names() {
  printf '%s\n' \
    "solo-windows-amd64.exe" \
    "solo-windows-arm64.exe" \
    "solo-darwin-amd64.dmg" \
    "solo-darwin-arm64.dmg"
}

assert_file_exists() {
  local path="$1"
  [[ -f "${path}" ]] || die "expected file is missing: ${path}"
}

main() {
  local dist_dir checksum_path asset
  dist_dir="${RELEASE_DIST_DIR}"
  checksum_path="${dist_dir}/SHA256SUMS"

  require_release_context
  [[ -d "${dist_dir}" ]] || die "dist directory does not exist: ${dist_dir}"
  assert_file_exists "${checksum_path}"

  while IFS= read -r asset; do
    assert_file_exists "${dist_dir}/${asset}"
  done < <(raw_dispatch_artifact_names)

  while IFS= read -r asset; do
    assert_file_exists "${dist_dir}/${asset}"
  done < <(package_artifact_names)

  if [[ "${SOLO_RELEASE_CHANNEL}" == "prerelease" ]]; then
    log "publish-input validation passed for prerelease channel; external repository publication remains disabled"
    return
  fi

  log "publish-input validation passed for stable channel"
}

main "$@"
