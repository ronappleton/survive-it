#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST_DIR="${DIST_DIR:-${ROOT_DIR}/dist}"
APP_NAME="${APP_NAME:-Survive It}"
PROJECT_NAME="${PROJECT_NAME:-survive-it}"
BUNDLE_ID="${BUNDLE_ID:-com.appengineltd.surviveit}"
ICONSET_DIR="${ICONSET_DIR:-${ROOT_DIR}/assets/images/SurviveIt.iconset}"
BASE_ICON="${BASE_ICON:-${ROOT_DIR}/assets/images/icon.jpg}"
ICON_BASENAME="${ICON_BASENAME:-SurviveIt}"
MACOS_SIGN_IDENTITY="${MACOS_SIGN_IDENTITY:--}"
MACOS_NOTARY_APPLE_ID="${MACOS_NOTARY_APPLE_ID:-}"
MACOS_NOTARY_TEAM_ID="${MACOS_NOTARY_TEAM_ID:-}"
MACOS_NOTARY_APP_PASSWORD="${MACOS_NOTARY_APP_PASSWORD:-}"

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

find_archive_for_arch() {
  local arch="$1"
  local exact_tgz="${DIST_DIR}/${PROJECT_NAME}_${VERSION}_darwin_${arch}.tar.gz"
  local exact_zip="${DIST_DIR}/${PROJECT_NAME}_${VERSION}_darwin_${arch}.zip"
  if [[ -f "${exact_tgz}" ]]; then
    echo "${exact_tgz}"
    return 0
  fi
  if [[ -f "${exact_zip}" ]]; then
    echo "${exact_zip}"
    return 0
  fi
  local found
  found="$(find "${DIST_DIR}" -maxdepth 1 -type f \( -name "*_darwin_${arch}.tar.gz" -o -name "*_darwin_${arch}.zip" \) | head -n 1 || true)"
  if [[ -n "${found}" ]]; then
    echo "${found}"
    return 0
  fi
  return 1
}

extract_binary_for_arch() {
  local arch="$1"
  local archive_path
  archive_path="$(find_archive_for_arch "${arch}" || true)"
  if [[ -z "${archive_path}" ]]; then
    return 1
  fi

  local extract_dir="${DIST_DIR}/macos-app-build/extracted/${arch}"
  rm -rf "${extract_dir}"
  mkdir -p "${extract_dir}"

  case "${archive_path}" in
    *.tar.gz)
      tar -xzf "${archive_path}" -C "${extract_dir}"
      ;;
    *.zip)
      ditto -x -k "${archive_path}" "${extract_dir}"
      ;;
    *)
      return 1
      ;;
  esac

  local found
  found="$(find "${extract_dir}" -type f -name "survive-it" -perm -u+x | head -n 1 || true)"
  if [[ -z "${found}" ]]; then
    found="$(find "${extract_dir}" -type f -name "survive-it" | head -n 1 || true)"
  fi
  if [[ -n "${found}" ]]; then
    echo "${found}"
    return 0
  fi
  return 1
}

resolve_binary_for_arch() {
  local arch="$1"
  local found
  found="$(find_binary_for_arch "${arch}" || true)"
  if [[ -n "${found}" ]]; then
    echo "${found}"
    return 0
  fi
  found="$(extract_binary_for_arch "${arch}" || true)"
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
  local icon_basename="$6"
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
  <key>CFBundleIconFile</key>
  <string>${icon_basename}</string>
  <key>CFBundleIconName</key>
  <string>${icon_basename}</string>
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

ensure_png_from_source() {
  local src="$1"
  local dst="$2"
  mkdir -p "$(dirname "${dst}")"
  sips -s format png "${src}" --out "${dst}" >/dev/null
}

build_icns() {
  local resources_dir="$1"
  local stage_dir="$2"
  local iconset_out="${stage_dir}/${ICON_BASENAME}.iconset"
  local icns_out="${resources_dir}/${ICON_BASENAME}.icns"
  mkdir -p "${iconset_out}"

  local names=(
    "icon_16x16"
    "icon_16x16@2x"
    "icon_32x32"
    "icon_32x32@2x"
    "icon_128x128"
    "icon_128x128@2x"
    "icon_256x256"
    "icon_256x256@2x"
    "icon_512x512"
    "icon_512x512@2x"
  )
  local generated_any=0

  for name in "${names[@]}"; do
    local src=""
    for ext in png PNG jpg JPG jpeg JPEG; do
      if [[ -f "${ICONSET_DIR}/${name}.${ext}" ]]; then
        src="${ICONSET_DIR}/${name}.${ext}"
        break
      fi
    done
    local dst="${iconset_out}/${name}.png"
    if [[ -n "${src}" ]]; then
      ensure_png_from_source "${src}" "${dst}"
      generated_any=1
      continue
    fi
    if [[ -f "${BASE_ICON}" ]]; then
      local size_part="${name#icon_}"
      local px="${size_part%x*}"
      px="${px%@2x}"
      local target_size="${px}"
      if [[ "${name}" == *"@2x" ]]; then
        target_size=$((px * 2))
      fi
      sips -s format png -z "${target_size}" "${target_size}" "${BASE_ICON}" --out "${dst}" >/dev/null
      generated_any=1
    fi
  done

  if [[ "${generated_any}" -eq 0 ]]; then
    return 1
  fi

  iconutil -c icns "${iconset_out}" -o "${icns_out}"
  [[ -f "${icns_out}" ]]
}

codesign_app() {
  local app_dir="$1"
  local identity="${MACOS_SIGN_IDENTITY}"
  if [[ "${identity}" == "-" ]]; then
    codesign --force --deep --sign - "${app_dir}"
  else
    codesign --force --deep --options runtime --timestamp --sign "${identity}" "${app_dir}"
  fi
  codesign --verify --deep --strict "${app_dir}"
}

should_notarize() {
  [[ -n "${MACOS_NOTARY_APPLE_ID}" && -n "${MACOS_NOTARY_TEAM_ID}" && -n "${MACOS_NOTARY_APP_PASSWORD}" ]]
}

notarize_and_staple() {
  local app_dir="$1"
  local stage_dir="$2"
  local submit_zip="${stage_dir}/notary-submit.zip"

  if ! should_notarize; then
    echo "Notarization credentials not set; skipping notarization." >&2
    return 0
  fi
  if [[ "${MACOS_SIGN_IDENTITY}" == "-" ]]; then
    echo "Notarization requested, but app is ad-hoc signed. Set MACOS_SIGN_IDENTITY to a Developer ID identity." >&2
    return 1
  fi

  rm -f "${submit_zip}"
  ditto -c -k --sequesterRsrc --keepParent "${app_dir}" "${submit_zip}"
  xcrun notarytool submit "${submit_zip}" \
    --apple-id "${MACOS_NOTARY_APPLE_ID}" \
    --team-id "${MACOS_NOTARY_TEAM_ID}" \
    --password "${MACOS_NOTARY_APP_PASSWORD}" \
    --wait
  xcrun stapler staple -v "${app_dir}"
  rm -f "${submit_zip}"
}

mkdir -p "${DIST_DIR}/macos-app-build"
declare -a ARTIFACTS=()

for arch in amd64 arm64; do
  binary_path="$(resolve_binary_for_arch "${arch}" || true)"
  if [[ -z "${binary_path}" ]]; then
    echo "Missing darwin binary/archive for ${arch} under ${DIST_DIR}" >&2
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

  if ! build_icns "${resources_dir}" "${stage_dir}"; then
    echo "Warning: could not build ${ICON_BASENAME}.icns from assets; app icon may be generic." >&2
  fi

  make_info_plist "${contents_dir}/Info.plist" "${VERSION}" "${APP_NAME}" "${executable_name}" "${BUNDLE_ID}" "${ICON_BASENAME}"

  codesign_app "${app_dir}"
  notarize_and_staple "${app_dir}" "${stage_dir}"

  rm -f "${artifact_path}"
  ditto -c -k --sequesterRsrc --keepParent "${app_dir}" "${artifact_path}"

  ARTIFACTS+=("${artifact_path}")
  echo "Created ${artifact_path}"
done

manifest="${DIST_DIR}/macos-app-artifacts.txt"
printf "%s\n" "${ARTIFACTS[@]}" > "${manifest}"
echo "Wrote artifact list to ${manifest}"

if ! should_notarize; then
  echo "Warning: app bundles were not notarized. Gatekeeper may report them as damaged when downloaded from the internet." >&2
fi
