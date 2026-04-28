# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

platform_for_target() {
  case "$1" in
    linux-amd64) printf '%s\n' linux/amd64 ;;
    linux-arm64) printf '%s\n' linux/arm64 ;;
    *) die "cannot infer Linux platform for target: $1" ;;
  esac
}

main() {
  local target="${1:-}"
  local artifact_name platform output_path
  local -a wails_args

  require_release_context
  [[ -n "${target}" ]] || die "missing Linux raw build target"
  validate_raw_target "${target}"

  case "${target}" in
    linux-amd64|linux-arm64) ;;
    *) die "target ${target} does not belong to the Linux raw build family" ;;
  esac

  artifact_name="$(raw_artifact_name "${target}")"
  platform="$(platform_for_target "${target}")"
  output_path="${RELEASE_REPO_ROOT}/build/bin/${artifact_name}"
  wails_args=(
    build
    -clean
    -platform "${platform}"
    -o "${artifact_name}"
    -nopackage
  )

  if [[ -n "${SOLO_BUILD_TAGS:-}" ]]; then
    wails_args+=(-tags "${SOLO_BUILD_TAGS}")
  fi

  mkdir -p "${RELEASE_DIST_DIR}"
  log "building ${target} -> ${artifact_name}"

  (
    cd "${RELEASE_REPO_ROOT}"
    mise exec -- wails "${wails_args[@]}"
  )

  [[ -f "${output_path}" ]] || die "expected build output is missing: ${output_path}"
  cp "${output_path}" "${RELEASE_DIST_DIR}/${artifact_name}"
  log "built raw artifact ${artifact_name}"
}

main "$@"
