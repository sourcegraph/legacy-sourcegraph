#!/bin/sh
set -eux
# Script to generate Go models from TypeSpec, which defines REST APIs
# Ideally, this bash script should be replaced with a `bazel run` command
# but we keep it as a bash script for now to bootstrap our usage of TypeSpec.

# Check if openapi-generator is installed
if ! command -v openapi-generator > /dev/null 2>&1
then
    echo "openapi-generator not found, installing from Homebrew..."
    brew install openapi-generator
fi

# Check if main.tsp exists
MAIN_TSP="internal/openapi/main.tsp"
if [ ! -f "$MAIN_TSP" ]; then
    echo "Error: $MAIN_TSP not found. The working directory must be at the root of the repo."
    exit 1
fi

pnpm install
pnpm -C internal/openapi compile
openapi-generator generate \
  -g go \
  -i internal/openapi/tsp-output/@typespec/openapi3/openapi.yaml \
  -o internal/openapi \
  -c internal/openapi/.openapi-generator/config.yml

echo "internal/openapi/go"
