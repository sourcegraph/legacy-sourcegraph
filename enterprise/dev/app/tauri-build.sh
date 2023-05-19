#!/usr/bin/env bash
set -exu

cd "$(dirname "${BASH_SOURCE[0]}")"/../../.. || exit 1

BIN_DIR=".bin"
DIST_DIR="dist"

download_artifacts() {
  mkdir -p .bin
  buildkite-agent artifact download "${BIN_DIR}/sourcegraph-backend-*" .bin/
  chmod -R +x .bin/
}

set_version() {
  local version
  local tauri_conf
  local tmp
  version=$1
  tauri_conf="./src-tauri/tauri.conf.json"
  tmp=$(mktemp)
  echo "--- Updating package version in '${tauri_conf}' to ${version}"
  jq --arg version "${version}" '.package.version = $version' "${tauri_conf}" > "${tmp}"
  mv "${tmp}" ./src-tauri/tauri.conf.json
}

bundle_path() {
  local platform="$(./enterprise/dev/app/detect_platform.sh)"
  echo  "./src-tauri/target/${platform}/release/bundle"
}

upload_dist() {
  local path="$(bundle_path)"
  echo "searching for artefacts in '${path}' and moving them to dist/"
  src=$(find "${path}" -type f \( -name "Sourcegraph*.dmg" -o -name "Sourcegraph*.tar.gz" -o -name "sourcegraph*.deb" -o -name "sourcegraph*.AppImage" -o -name "sourcegraph*.tar.gz" \));

  mkdir -p dist
  for from in ${src}; do
    mv -vf "${from}" "./${DIST_DIR}/"
  done

  echo --- Uploading artifacts from dist
  ls -al ./dist
  buildkite-agent artifact upload "./${DIST_DIR}/*"

}

create_app_archive() {
  local version=$1
  local path="$(bundle_path)"
  local app_path=$(find "${path}" -type d -name "Sourcegraph.app")

  # # we have to handle Sourcegraph.App differently since it is a dir
  if [[ -d  ${app_path} ]]; then
    local arch
    local target
    arch="$(uname -m)"
    if [[ "${arch}" == "arm64" ]]; then
      arch="aarch64"
    fi

    target="Sourcegraph.${version}.${arch}.app.tar.gz"
    pushd .
    cd "${path}/macos/"
    echo "--- :file_cabinet: Creating archive ${target}"
    tar -czvf "${target}" "Sourcegraph.app"
    popd
  fi
}

cleanup_codesigning() {
    if [[ $(security list-keychains -d user | grep -q "my_temporary_keychain") ]]; then
      set +e
      echo "--- :broom: cleaning up keychains"
      security delete-keychain my_temporary_keychain.keychain
      set -e
    fi
}

pre_codesign() {
  local binary_path
  binary_path=$1
  # Tauri won't code sign our sidecar sourcegraph-backend Go binary for us, so we need to do it on
  # our own. https://github.com/tauri-apps/tauri/discussions/2269
  # For details on code signing, see doc/dev/background-information/app/codesigning.md
  trap 'cleanup_codesigning' ERR INT TERM EXIT

  if [[ ${CI} == "true" ]]; then
    local secrets
    echo "--- :aws: Retrieving signing secrets"
    secrets=$(aws secretsmanager get-secret-value --secret-id sourcegraph/mac-codesigning | jq '.SecretString |  fromjson')
    export APPLE_SIGNING_IDENTITY="$(echo "${secrets}" | jq -r '.APPLE_SIGNING_IDENTITY')"
    export APPLE_CERTIFICATE="$(echo "${secrets}" | jq -r '.APPLE_CERTIFICATE')"
    export APPLE_CERTIFICATE_PASSWORD="$(echo "${secrets}" | jq -r  '.APPLE_CERTIFICATE_PASSWORD')"
    export APPLE_ID="$(echo "${secrets}" | jq -r '.APPLE_ID')"
    export APPLE_PASSWORD="$(echo "${secrets}" | jq -r '.APPLE_PASSWORD')"
  fi
  # We expect the same APPLE_ env vars that Tauri does here, see https://tauri.app/v1/guides/distribution/sign-macos
  cleanup_codesigning
  security create-keychain -p my_temporary_keychain_password my_temporary_keychain.keychain
  security set-keychain-settings my_temporary_keychain.keychain
  security unlock-keychain -p my_temporary_keychain_password my_temporary_keychain.keychain
  security list-keychains -d user -s my_temporary_keychain.keychain "$(security list-keychains -d user | sed 's/["]//g')"

  echo "$APPLE_CERTIFICATE" >cert.p12.base64
  base64 -d -i cert.p12.base64 -o cert.p12

  security import ./cert.p12 -k my_temporary_keychain.keychain -P "$APPLE_CERTIFICATE_PASSWORD" -T /usr/bin/codesign
  security set-key-partition-list -S apple-tool:,apple:, -s -k my_temporary_keychain_password -D "$APPLE_SIGNING_IDENTITY" -t private my_temporary_keychain.keychain

  echo "--- :mac::pencil2: binary: ${binary_path}"
  codesign --force -s "$APPLE_SIGNING_IDENTITY" --keychain my_temporary_keychain.keychain --deep "${binary_path}"

  security delete-keychain my_temporary_keychain.keychain
  security list-keychains -d user -s login.keychain
}

secret_value() {
  local name
  local value
  name=$1
  if [[ $(uname -s) == "Darwin" ]]; then
    # host is in aws - probably
    value=$(aws secretsmanager get-secret-value --secret-id "${name}" | jq '.SecretString | fromjson')
  else
    # On Linux we assume we're in GCP thus the secret should be injected as an evironment variable. Please check the instance configuration
    value=""
  fi
  echo "${value}"
}

build() {
  echo --- :magnify_glass: detecting platform
  local platform="$(./enterprise/dev/app/detect_platform.sh)"
  echo "platform is: ${platform}"

  if [[ ${CI} == "true" ]]; then
    local secrets
    echo "--- :aws::gcp::tauri: Retrieving tauri signing secrets"
    secrets=$(secret_value "sourcegraph/tauri-key")
    # if the value is not found in secrets we default to an empty string
    export TAURI_PRIVATE_KEY="${TAURI_PRIVATE_KEY:-"$(echo "${secrets}" | jq -r '.TAURI_PRIVATE_KEY' | base64 -d || echo '')"}"
    export TAURI_KEY_PASSWORD="${TAURI_KEY_PASSWORD:-"$(echo "${secrets}" | jq -r '.TAURI_KEY_PASSWORD' || echo '')"}"
  fi

  echo "--- :tauri: Building Application (${VERSION}) for platform: ${platform}"
  NODE_ENV=production pnpm run build-app-shell
  pnpm tauri build --bundles deb,appimage,app,dmg,updater --target "${platform}"
}

if [[ ${CI:-""} == "true" ]]; then
  download_artifacts
fi

VERSION=$(./enterprise/dev/app/app_version.sh)
set_version "${VERSION}"


if [[ ${CODESIGNING:-"0"} == 1 && $(uname -s) == "Darwin" ]]; then
  # We want any xcode related tools to be picked up first so inject it here in the path
  export PATH="$(xcode-select -p)/usr/bin:$PATH"
  # If on a macOS host, Tauri will invoke the base64 command as part of the code signing process.
  # it expects the macOS base64 command, not the gnutils one provided by homebrew, so we prefer
  # that one here:
  export PATH="/usr/bin/:$PATH"

  echo "--- :tauri::mac: Performing code signing"
  binaries=$(find ${BIN_DIR} -type f -name "*apple*")
  # if the paths contain spaces this for loop will fail, but we're pretty sure the binaries in bin don't contain spaces
  for binary in ${binaries}; do
    pre_codesign "${binary}"
  done
fi

PLATFORM="$(./enterprise/dev/app/detect_platform.sh)"
build "${PLATFORM}"

create_app_archive "${VERSION}"

if [[ ${CI:-""} == "true" ]]; then
  upload_dist
fi
