# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

main() {
  local dist_dir checksum_path
  dist_dir="${RELEASE_DIST_DIR}"
  checksum_path="${dist_dir}/SHA256SUMS"

  require_release_context
  [[ -d "${dist_dir}" ]] || die "dist directory does not exist: ${dist_dir}"

  : > "${checksum_path}"
  while IFS= read -r file; do
    local base hash
    base="$(basename "${file}")"
    [[ "${base}" == "SHA256SUMS" ]] && continue
    hash="$(hash_file "${file}")"
    printf '%s  %s\n' "${hash}" "${base}" >> "${checksum_path}"
  done < <(list_top_level_files)

  log "generated checksums at ${checksum_path}"
}

main "$@"
