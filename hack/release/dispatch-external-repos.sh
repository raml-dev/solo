# Copyright 2026-present raml-dev
# SPDX-License-Identifier: AGPL-3.0-only

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=hack/release/lib.sh
source "${SCRIPT_DIR}/lib.sh"

write_dispatch_payload() {
  local payload_path="$1"
  local event_type="$2"
  local object_key="$3"
  local object_name="$4"
  local html_release_url
  local release_app_name
  local release_description
  local release_license

  html_release_url="${GITHUB_SERVER_URL:-https://github.com}/${GITHUB_REPOSITORY}/releases/tag/${GITHUB_REF_NAME}"
  release_app_name="${RELEASE_APP_NAME}"
  release_description="${RELEASE_DESCRIPTION}"
  release_license="${RELEASE_LICENSE}"

  jq -n \
    --arg event_type "${event_type}" \
    --arg object_key "${object_key}" \
    --arg object_name "${object_name}" \
    --arg source_repository "${GITHUB_REPOSITORY}" \
    --arg version "${GITHUB_REF_NAME}" \
    --arg api_release_url "${GITHUB_API_URL}/repos/${GITHUB_REPOSITORY}/releases/tags/${GITHUB_REF_NAME}" \
    --arg html_release_url "${html_release_url}" \
    --arg app_name "${release_app_name}" \
    --arg description "${release_description}" \
    --arg license "${release_license}" \
    '
    {
      event_type: $event_type,
      client_payload: (
        {
          release: {
            version: $version,
            api_url: $api_release_url,
            html_url: $html_release_url
          }
        }
        + {
            ($object_key): {
              name: $object_name,
              source_repository: $source_repository,
              app_name: $app_name,
              description: $description,
              license: $license
            }
          }
      )
    }
    ' > "${payload_path}"
}

dispatch_linux_packages_repo() {
  dispatch_repository_release_event \
    "${LINUX_PACKAGE_REPOS_REPOSITORY:-raml-dev/linux-packages-repos}" \
    "linux-package-release" \
    "package" \
    "${RELEASE_PACKAGE_NAME}"
}

dispatch_homebrew_tap() {
  dispatch_repository_release_event \
    "${HOMEBREW_TAP_REPOSITORY:-raml-dev/homebrew-tap}" \
    "homebrew-cask-release" \
    "cask" \
    "${RELEASE_PACKAGE_NAME}"
}

dispatch_scoop_bucket() {
  dispatch_repository_release_event \
    "${SCOOP_BUCKET_REPOSITORY:-raml-dev/scoop-bucket}" \
    "scoop-manifest-release" \
    "manifest" \
    "${RELEASE_PACKAGE_NAME}"
}

dispatch_repository_release_event() {
  local target_repo="$1"
  local event_type="$2"
  local object_key="$3"
  local object_name="$4"
  local api_url payload

  api_url="${GITHUB_API_URL}/repos/${target_repo}/dispatches"
  payload="$(mktemp)"

  write_dispatch_payload "${payload}" "${event_type}" "${object_key}" "${object_name}"

  curl -fsSL \
    -X POST \
    -H "Authorization: Bearer ${EXTERNAL_REPOS_DISPATCH_TOKEN}" \
    -H "Accept: application/vnd.github+json" \
    "${api_url}" \
    --data-binary "@${payload}" >/dev/null

  rm -f "${payload}"
  log "dispatched ${GITHUB_REF_NAME} to ${target_repo}"
}

main() {
  require_env "GITHUB_REF_NAME"
  require_env "GITHUB_API_URL"
  require_env "GITHUB_REPOSITORY"
  require_env "EXTERNAL_REPOS_DISPATCH_TOKEN"
  require_env "RELEASE_PACKAGE_NAME"
  require_env "RELEASE_APP_NAME"
  require_env "RELEASE_DESCRIPTION"
  require_env "RELEASE_LICENSE"

  log "dispatching stable release to linux-packages-repos"
  dispatch_linux_packages_repo

  log "dispatching stable release to homebrew-tap"
  dispatch_homebrew_tap

  log "dispatching stable release to scoop-bucket"
  dispatch_scoop_bucket
}

main "$@"
