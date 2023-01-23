#!/bin/bash

cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."
cd "/root" # TODO: Only used in local Docker

set -euf -o pipefail
tmpdir=$(mktemp -d -t melange-bin.XXXXXXXX)
function cleanup() {
  echo "Removing $tmpdir"
  rm -rf "$tmpdir"
}
trap cleanup EXIT

# Install requisite packages
apt update
apt install -y bubblewrap

(
  cd "$tmpdir"
  mkdir bin

  # Install melange
  wget https://github.com/chainguard-dev/melange/releases/download/v0.2.0/melange_0.2.0_linux_amd64.tar.gz
  tar zxf melange_0.2.0_linux_amd64.tar.gz
  mv melange_0.2.0_linux_amd64/melange bin/melange

  # Install apk
  wget https://gitlab.alpinelinux.org/alpine/apk-tools/-/package_files/62/download -O bin/apk
  chmod +x bin/apk
)

export PATH="$tmpdir/bin:$PATH"

if [ $# -eq 0 ]; then
  echo "No arguments supplied - provide the melange YAML file to build"
  exit 0
fi

name=${1%/}

if [ ! -f "wolfi-packages/${name}.yaml" ]; then
  echo "File '$name.yaml' does not exist"
  exit 1
fi

cd "wolfi-packages"

echo " * Building melange package '$name'"
# TODO: Signing key
melange build "$name.yaml" --arch x86_64

# Upload package as build artifact
buildkite-agent upload wolfi-packages/packages/*/*
