#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

# Remove bazelisk from path
PATH=$(echo "${PATH}" | awk -v RS=: -v ORS=: '/bazelisk/ {next} {print}')
export PATH

cd "${BUILD_WORKSPACE_DIRECTORY}"

aspectRC="$(mktemp -t "aspect-generated.bazelrc.XXXXXX")"
rosetta bazelrc > "$aspectRC"

bazel --bazelrc="$aspectRC" run //:gazelle

if [ "${CI:-}" ]; then
  git ls-files --exclude-standard --others | xargs git add --intent-to-add || true

  diff_file=$(mktemp)
  trap 'rm -f "${diff_file}"' EXIT

  EXIT_CODE=0
  git diff --color=never --output="${diff_file}" --exit-code || EXIT_CODE=$?

  # if we have a diff, BUILD files were updated so we notify people
  if [[ $EXIT_CODE -ne 0 ]]; then
    cat "${diff_file}"
    exit 1
  fi
fi
