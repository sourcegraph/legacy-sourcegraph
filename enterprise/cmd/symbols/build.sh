#!/usr/bin/env bash

# This script builds the symbols docker image.

cd "$(dirname "${BASH_SOURCE[0]}")/../../.."
set -eu

echo "--- docker build"
docker build -f cmd/symbols/Dockerfile -t "$IMAGE" "$(pwd)" \
  --progress=plain \
  --build-arg COMMIT_SHA \
  --build-arg DATE \
  --build-arg VERSION \
  --build-arg exe=enterprise-symbols \
  --build-arg pkg=github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols
