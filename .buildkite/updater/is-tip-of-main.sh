#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.."
set -euxo pipefail

# There are two GITHUB tokens required on our pipeline,
# one for ghe.sgdev.org and this. This line ensures for
# this job we're using the correct token
export GITHUB_TOKEN="${BUILDKITE_GITHUBDOTCOM_TOKEN}"

COMMIT="${BUILDKITE_COMMIT}"

API_SLUG="repos/sourcegraph/sourcegraph/commits"
function get_branch_tip() {
  local ref="$1"

  # https://docs.github.com/en/rest/reference/repos#list-commits
  gh api "${API_SLUG}?sha=${ref}&per_page=1" --jq '.[].sha'
}

REF="main"
tip_of_main="$(get_branch_tip ${REF})"

[[ "$tip_of_main" == "$COMMIT" ]]
