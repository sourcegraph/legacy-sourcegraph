#!/usr/bin/env bash

# We want to build multiple go binaries, so we use a custom build step on CI.
cd "$(dirname "${BASH_SOURCE[0]}")"/../..
set -ex

OUTPUT=$(mktemp -d -t sgdockerbuild_XXXXXXX)
cleanup() {
  rm -rf "$OUTPUT"
}
trap cleanup EXIT

# Environment for building linux binaries
export GO111MODULE=on
export GOARCH=amd64
export GOOS=linux
export CGO_ENABLED=0

echo "--- go build"
pkg="github.com/sourcegraph/sourcegraph/cmd/frontend"
go build -trimpath -ldflags "-X github.com/sourcegraph/sourcegraph/internal/version.version=$VERSION  -X github.com/sourcegraph/sourcegraph/internal/version.timestamp=$(date +%s)" -buildmode exe -tags dist -o "$OUTPUT/$(basename $pkg)" "$pkg"

# Enable image build caching via CACHE=true
BUILD_CACHE="--no-cache"
if [[ "$CACHE" == "true" ]]; then
  BUILD_CACHE=""
fi

echo "--- docker build $IMAGE"
docker pull us.gcr.io/sourcegraph-dev/frontend:insiders || true
docker build --cache-from us.gcr.io/sourcegraph-dev/frotend:insiders ${BUILD_CACHE} -f cmd/frontend/Dockerfile -t "$IMAGE" "$OUTPUT" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION
