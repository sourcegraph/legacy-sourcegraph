#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

echo "--- bazel build"
./dev/ci/bazel.sh build //enterprise/cmd/symbols \
  --stamp \
  --workspace_status_command=./dev/bazel_stamp_vars.sh \
  --platforms @zig_sdk//platform:linux_amd64 \
  --extra_toolchains @zig_sdk//toolchain:linux_amd64_musl

out=$(bazel cquery //enterprise/cmd/symbols --output=files)
cp "$out" "$OUTPUT"
cp cmd/symbols/ctags-install-alpine.sh "$OUTPUT"

docker build -f cmd/symbols/Dockerfile.bazel -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
