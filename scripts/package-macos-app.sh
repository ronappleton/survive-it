#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-${ROOT_DIR}/dist}"
APP_NAME="${APP_NAME:-Survive It}"
PROJECT_NAME="${PROJECT_NAME:-survive-it}"
BUNDLE_ID="${BUNDLE_ID:-com.appengineltd.surviveit}"

RAW_VERSION="${1:-}"
if [[ -z "${RAW_VERSION}" ]]; then
  RAW_VERSION="$(git -C "${ROOT_DIR}" describe --tags --always 2>/dev/null || echo dev)"
fi
VERSION="${RAW_VERSION#v}"

if [[ ! -d "${DIST_DIR}" ]]; then
  echo "dist directory not found: ${DIST_DIR}" >&2
  echo "Run GoReleaser first (e.g. goreleaser release --clean)." >&2
  exit 1
fi

find_binary_for_arch() {
  local arch="$1"
  local direct="${DIST_DIR}/survive-it-gui-darwin_darwin_${arch}/survive-it"
  if [[ -f "${direct}" ]]; then
    echo "${direct}"
    return 0
  fi
  local found
  found="$(find "${DIST_DIR}" -type f -path "*darwin_${arch}/survive-it" | head -n 1 || true)"
  if [[ -n "${found}" ]]; then
    echo "${found}"
    return 0
  fi
  return 1
}

make_info_plist() {
  local plist_path="$1"
  local short_version="$2"
  local bundle_name="$3"
  local executable_name="$4"
  local bundle_id="$5"
  cat > "${plist_path}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDevelopmentRegion</key>
  <string>en</string>
  <key>CFBundleDisplayName</key>
  <string>${bundle_name}</string>
  <key>CFBundleExecutable</key>
  <string>${executable_name}</string>
  <key>CFBundleIdentifier</key>
  <string>${bundle_id}</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundleName</key>
  <string>${bundle_name}</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>CFBundleShortVersionString</key>
  <string>${short_version}</string>
  <key>CFBundleVersion</key>
  <string>${short_version}</string>
  <key>LSMinimumSystemVersion</key>
  <string>11.0</string>
  <key>NSHighResolutionCapable</key>
  <true/>
</dict>
</plist>
EOF
}

mkdir -p "${DIST_DIR}/macos-app-build"
declare -a ARTIFACTS=()

for arch in amd64 arm64; do
  binary_path="$(find_binary_for_arch "${arch}" || true)"
  if [[ -z "${binary_path}" ]]; then
    echo "Missing darwin binary for ${arch} under ${DIST_DIR}" >&2
    exit 1
  fi

  stage_dir="${DIST_DIR}/macos-app-build/${arch}"
  app_dir="${stage_dir}/${APP_NAME}.app"
  contents_dir="${app_dir}/Contents"
  macos_dir="${contents_dir}/MacOS"
  resources_dir="${contents_dir}/Resources"
  executable_name="survive-it"
  artifact_path="${DIST_DIR}/${PROJECT_NAME}_${VERSION}_darwin_${arch}.app.zip"

  rm -rf "${stage_dir}"
  mkdir -p "${macos_dir}" "${resources_dir}"

  cp "${binary_path}" "${macos_dir}/${executable_name}"
  chmod +x "${macos_dir}/${executable_name}"

  if [[ -d "${ROOT_DIR}/assets" ]]; then
    mkdir -p "${resources_dir}/assets"
    rsync -a "${ROOT_DIR}/assets/" "${resources_dir}/assets/"
  fi

  make_info_plist "${contents_dir}/Info.plist" "${VERSION}" "${APP_NAME}" "${executable_name}" "${BUNDLE_ID}"

  rm -f "${artifact_path}"
  ditto -c -k --sequesterRsrc --keepParent "${app_dir}" "${artifact_path}"

  ARTIFACTS+=("${artifact_path}")
  echo "Created ${artifact_path}"
done

manifest="${DIST_DIR}/macos-app-artifacts.txt"
printf "%s\n" "${ARTIFACTS[@]}" > "${manifest}"
echo "Wrote artifact list to ${manifest}"
