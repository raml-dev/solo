# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

deb_artifact_name() {
  local target="$1"
  case "$target" in
    deb-amd64) printf 'solo_%s_amd64.deb\n' "${SOLO_RELEASE_VERSION}" ;;
    deb-arm64) printf 'solo_%s_arm64.deb\n' "${SOLO_RELEASE_VERSION}" ;;
    *) die "cannot infer deb artifact name for target: $target" ;;
  esac
}

rpm_artifact_name() {
  local target="$1"
  case "$target" in
    rpm-amd64) printf 'solo-%s-1.x86_64.rpm\n' "${SOLO_RELEASE_VERSION}" ;;
    rpm-arm64) printf 'solo-%s-1.aarch64.rpm\n' "${SOLO_RELEASE_VERSION}" ;;
    *) die "cannot infer rpm artifact name for target: $target" ;;
  esac
}

arch_artifact_name() {
  local target="$1"
  case "$target" in
    arch-amd64) printf 'solo-%s-1-x86_64.pkg.tar.zst\n' "${SOLO_RELEASE_VERSION}" ;;
    arch-arm64) printf 'solo-%s-1-aarch64.pkg.tar.zst\n' "${SOLO_RELEASE_VERSION}" ;;
    *) die "cannot infer Arch artifact name for target: $target" ;;
  esac
}

linux_raw_target_for_package_target() {
  case "$1" in
    deb-amd64|rpm-amd64|arch-amd64) printf '%s\n' linux-amd64 ;;
    deb-arm64|rpm-arm64|arch-arm64) printf '%s\n' linux-arm64 ;;
    *) die "cannot infer Linux raw target for package target: $1" ;;
  esac
}

nfpm_packager_for_target() {
  case "$1" in
    deb-*) printf '%s\n' deb ;;
    rpm-*) printf '%s\n' rpm ;;
    arch-*) printf '%s\n' archlinux ;;
    *) die "cannot infer nFPM packager for target: $1" ;;
  esac
}

package_arch_for_target() {
  case "$1" in
    *-amd64) printf '%s\n' amd64 ;;
    *-arm64) printf '%s\n' arm64 ;;
    *) die "cannot infer package architecture for target: $1" ;;
  esac
}

nfpm_arch_for_target() {
  case "$1" in
    arch-amd64) printf '%s\n' x86_64 ;;
    arch-arm64) printf '%s\n' aarch64 ;;
    *)
      package_arch_for_target "$1"
      ;;
  esac
}

linux_package_root_dir() {
  local target="$1"
  printf '%s\n' "${RELEASE_TMP_DIR}/packages/${target}/root"
}

linux_nfpm_config_path() {
  local target="$1"
  printf '%s\n' "${RELEASE_TMP_DIR}/packages/${target}/nfpm.yaml"
}

linux_stage_package_root() {
  local package_target="$1"
  local root_dir raw_target raw_artifact

  root_dir="$(linux_package_root_dir "${package_target}")"
  raw_target="$(linux_raw_target_for_package_target "${package_target}")"
  raw_artifact="${RELEASE_DIST_DIR}/$(raw_artifact_name "${raw_target}")"

  [[ -f "${raw_artifact}" ]] || die "required Linux raw artifact is missing: ${raw_artifact}"
  [[ -f "${RELEASE_LINUX_DESKTOP_FILE_SOURCE}" ]] || die "Linux desktop file is missing"
  [[ -f "${RELEASE_LINUX_METAINFO_SOURCE}" ]] || die "Linux metainfo file is missing"
  [[ -f "${RELEASE_LINUX_ICON_SOURCE}" ]] || die "Linux icon source is missing"
  [[ -f "${RELEASE_REPO_ROOT}/LICENSE" ]] || die "LICENSE file is missing"

  rm -rf "$(dirname "${root_dir}")"
  mkdir -p \
    "${root_dir}/usr/bin" \
    "${root_dir}/usr/share/applications" \
    "${root_dir}/usr/share/metainfo" \
    "${root_dir}/usr/share/icons/hicolor/512x512/apps" \
    "${root_dir}/usr/share/doc/solo"

  install -m0755 "${raw_artifact}" "${root_dir}/usr/bin/solo"
  install -m0644 "${RELEASE_LINUX_DESKTOP_FILE_SOURCE}" "${root_dir}/usr/share/applications/${RELEASE_LINUX_DESKTOP_FILE_NAME}"
  install -m0644 "${RELEASE_LINUX_METAINFO_SOURCE}" "${root_dir}/usr/share/metainfo/${RELEASE_LINUX_METAINFO_FILE_NAME}"
  install -m0644 "${RELEASE_LINUX_ICON_SOURCE}" "${root_dir}/usr/share/icons/hicolor/512x512/apps/${RELEASE_LINUX_DESKTOP_ID}.png"
  install -m0644 "${RELEASE_REPO_ROOT}/LICENSE" "${root_dir}/usr/share/doc/solo/LICENSE"

  printf '%s\n' "${root_dir}"
}

linux_render_nfpm_config() {
  local package_target="$1"
  local root_dir config_path arch packager

  root_dir="$(linux_package_root_dir "${package_target}")"
  config_path="$(linux_nfpm_config_path "${package_target}")"
  arch="$(nfpm_arch_for_target "${package_target}")"
  packager="$(nfpm_packager_for_target "${package_target}")"

  mkdir -p "$(dirname "${config_path}")"
  sed \
    -e "s|@PACKAGE_ROOT@|${root_dir}|g" \
    -e "s|@VERSION@|${SOLO_RELEASE_VERSION}|g" \
    -e "s|@ARCH@|${arch}|g" \
    -e "s|@PACKAGER@|${packager}|g" \
    "${RELEASE_LINUX_NFPM_TEMPLATE_PATH}" > "${config_path}"

  printf '%s\n' "${config_path}"
}

build_deb_package() {
  local target="$1"
  local config_path artifact_path

  linux_stage_package_root "${target}" >/dev/null
  config_path="$(linux_render_nfpm_config "${target}")"
  artifact_path="${RELEASE_DIST_DIR}/$(deb_artifact_name "${target}")"

  nfpm package \
    --config "${config_path}" \
    --packager deb \
    --target "${artifact_path}"

  [[ -f "${artifact_path}" ]] || die "expected deb artifact is missing: ${artifact_path}"
  printf '%s\n' "${artifact_path}"
}

build_rpm_package() {
  local target="$1"
  local config_path artifact_path

  linux_stage_package_root "${target}" >/dev/null
  config_path="$(linux_render_nfpm_config "${target}")"
  artifact_path="${RELEASE_DIST_DIR}/$(rpm_artifact_name "${target}")"

  nfpm package \
    --config "${config_path}" \
    --packager rpm \
    --target "${artifact_path}"

  [[ -f "${artifact_path}" ]] || die "expected rpm artifact is missing: ${artifact_path}"
  printf '%s\n' "${artifact_path}"
}

build_arch_package() {
  local target="$1"
  local config_path artifact_path

  linux_stage_package_root "${target}" >/dev/null
  config_path="$(linux_render_nfpm_config "${target}")"
  artifact_path="${RELEASE_DIST_DIR}/$(arch_artifact_name "${target}")"

  nfpm package \
    --config "${config_path}" \
    --packager archlinux \
    --target "${artifact_path}"

  [[ -f "${artifact_path}" ]] || die "expected Arch package artifact is missing: ${artifact_path}"
  printf '%s\n' "${artifact_path}"
}

main() {
  local target artifact_path
  local package_targets=(
    deb-amd64
    deb-arm64
    rpm-amd64
    rpm-arm64
    arch-amd64
    arch-arm64
  )

  for target in "${package_targets[@]}"; do
    case "${target}" in
      deb-*)
        artifact_path="$(build_deb_package "${target}")"
        log "built Debian package artifact at ${artifact_path}"
        ;;
      rpm-*)
        artifact_path="$(build_rpm_package "${target}")"
        log "built RPM package artifact at ${artifact_path}"
        ;;
      arch-*)
        artifact_path="$(build_arch_package "${target}")"
        log "built Arch package artifact at ${artifact_path}"
        ;;
      *)
        die "unsupported Linux package target: ${target}"
        ;;
    esac
  done
}

main "$@"
