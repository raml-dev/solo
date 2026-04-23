# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

platform_for_target() {
  case "$1" in
    darwin-amd64) printf '%s\n' darwin/amd64 ;;
    darwin-arm64) printf '%s\n' darwin/arm64 ;;
    *) die "cannot infer macOS platform for target: $1" ;;
  esac
}

macos_info_plist_path() {
  local bundle_path="${1}"
  printf '%s\n' "${bundle_path}/Contents/Info.plist"
}

macos_read_plist_value() {
  local plist_path="$1"
  local key="$2"
  /usr/bin/plutil -extract "${key}" raw -o - "${plist_path}"
}

macos_verify_bundle_metadata() {
  local bundle_path="$1"
  local plist_path executable_name actual_executable

  plist_path="$(macos_info_plist_path "${bundle_path}")"
  [[ -f "${plist_path}" ]] || die "macOS bundle Info.plist is missing: ${plist_path}"

  executable_name="${RELEASE_MACOS_EXECUTABLE_NAME}"
  actual_executable="$(macos_read_plist_value "${plist_path}" "CFBundleExecutable")"
  [[ "${actual_executable}" == "${executable_name}" ]] || die "macOS bundle executable mismatch: expected ${executable_name}, got ${actual_executable}"
  [[ -x "${bundle_path}/Contents/MacOS/${executable_name}" ]] || die "macOS bundle executable is missing: ${bundle_path}/Contents/MacOS/${executable_name}"
}

macos_sign_path() {
  local path="$1"
  log "codesigning ${path}"
  codesign --force --sign - "${path}"
}

macos_sign_nested_bundle_content() {
  local bundle_path="$1"
  local contents_path
  contents_path="${bundle_path}/Contents"

  [[ -d "${contents_path}" ]] || die "macOS bundle contents directory is missing: ${contents_path}"

  while IFS= read -r -d '' path; do
    macos_sign_path "${path}"
  done < <(find "${contents_path}" -depth \( -name '*.app' -o -name '*.appex' -o -name '*.framework' -o -name '*.xpc' \) -print0)

  while IFS= read -r -d '' path; do
    macos_sign_path "${path}"
  done < <(find "${contents_path}" -type f \( -name '*.dylib' -o -name '*.so' -o -path "${contents_path}/MacOS/*" \) -print0)
}

macos_verify_bundle_signature() {
  local bundle_path="$1"
  codesign --verify --deep --strict --verbose=2 "${bundle_path}"
  codesign -dv --verbose=4 "${bundle_path}" || true
}

macos_prepare_dmg_staging_dir() {
  local bundle_path="$1"
  local staging_dir

  mkdir -p "${RELEASE_TMP_DIR}"
  staging_dir="$(mktemp -d "${RELEASE_TMP_DIR}/darwin-stage.XXXXXX")"
  ditto "${bundle_path}" "${staging_dir}/${RELEASE_MACOS_BUNDLE_NAME}"
  ln -s /Applications "${staging_dir}/Applications"
  printf '%s\n' "${staging_dir}"
}

macos_create_dmg() {
  local staging_dir="$1"
  local dmg_path="$2"

  rm -f "${dmg_path}"
  hdiutil create \
    -volname "${RELEASE_MACOS_VOLUME_NAME}" \
    -srcfolder "${staging_dir}" \
    -ov \
    -format UDZO \
    "${dmg_path}"
}

macos_sign_dmg() {
  local dmg_path="$1"
  macos_sign_path "${dmg_path}"
}

macos_verify_dmg_signature() {
  local dmg_path="$1"
  codesign --verify --verbose=2 "${dmg_path}"
  codesign -dv --verbose=4 "${dmg_path}" || true
}

macos_verify_dmg_contents() {
  local dmg_path="$1"
  local mount_dir entry_count

  mkdir -p "${RELEASE_TMP_DIR}"
  mount_dir="$(mktemp -d "${RELEASE_TMP_DIR}/darwin-mount.XXXXXX")"
  hdiutil attach -readonly -noautoopen -mountpoint "${mount_dir}" "${dmg_path}" >/dev/null

  [[ -d "${mount_dir}/${RELEASE_MACOS_BUNDLE_NAME}" ]] || die "mounted DMG is missing ${RELEASE_MACOS_BUNDLE_NAME}"
  [[ -L "${mount_dir}/Applications" ]] || die "mounted DMG is missing Applications symlink"
  [[ "$(readlink "${mount_dir}/Applications")" == "/Applications" ]] || die "mounted DMG Applications symlink does not target /Applications"

  entry_count="$(find "${mount_dir}" -mindepth 1 -maxdepth 1 ! -name '.DS_Store' | wc -l | tr -d ' ')"
  [[ "${entry_count}" == "2" ]] || die "mounted DMG contains unexpected top-level entries"

  hdiutil detach "${mount_dir}" >/dev/null
  rmdir "${mount_dir}"
}

main() {
  local target="${1:-}"
  local artifact_name platform bundle_path dmg_path staging_dir
  local -a wails_args

  require_release_context
  [[ -n "${target}" ]] || die "missing macOS raw build target"
  validate_raw_target "${target}"

  case "${target}" in
    darwin-amd64|darwin-arm64) ;;
    *) die "target ${target} does not belong to the macOS raw build family" ;;
  esac

  artifact_name="$(raw_artifact_name "${target}")"
  platform="$(platform_for_target "${target}")"
  bundle_path="${RELEASE_MACOS_BUNDLE_PATH}"
  dmg_path="${RELEASE_DIST_DIR}/${artifact_name}"
  wails_args=(
    build
    -clean
    -platform "${platform}"
    -o "${RELEASE_MACOS_EXECUTABLE_NAME}"
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

  [[ -d "${bundle_path}" ]] || die "expected app bundle is missing: ${bundle_path}"
  macos_verify_bundle_metadata "${bundle_path}"
  macos_sign_nested_bundle_content "${bundle_path}"
  macos_sign_path "${bundle_path}"
  macos_verify_bundle_signature "${bundle_path}"

  staging_dir="$(macos_prepare_dmg_staging_dir "${bundle_path}")"
  macos_create_dmg "${staging_dir}" "${dmg_path}"
  macos_sign_dmg "${dmg_path}"
  macos_verify_dmg_signature "${dmg_path}"
  macos_verify_dmg_contents "${dmg_path}"
  rm -rf "${staging_dir}"

  [[ -f "${dmg_path}" ]] || die "expected DMG output is missing: ${dmg_path}"
  log "built raw artifact ${artifact_name}"
}

main "$@"
