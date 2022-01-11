#!/usr/bin/env bash

set -euo pipefail

function usage {
  cat <<EOF
Usage: go-test.sh [only|exclude package-path-1 package-path-2 ...]

Run go tests, optionally restricting which ones based on the only and exclude coommands.

EOF
}

function go_test() {
  local test_packages="$1"
  local tmpfile=$(mktemp)
  trap "rm $tmpfile" EXIT

  # shellcheck disable=SC2086
  go test \
    -timeout 10m \
    -coverprofile=coverage.txt \
    -covermode=atomic \
    -race \
    -v \
    $TEST_PACKAGES \
    | tee "$tmpfile"

  local xml=$(cat "$tmpfile" | go-junit-report)
  # escape xml output properly for JSON
  local quoted_xml="$( jq --null-input --compact-output --arg str "$xml" '$str' )"

  local data=$(cat <<EOF
{
  "format": "junit",
  "run_env": {
    "CI": "buildkite",
    "key": "$BUILDKITE_BUILD_ID",
    "job_id": "$BUILDKITE_JOB_ID",
    "branch": "$BUILDKITE_BRANCH_NAME",
    "commit_sha": "$BUILDKITE_COMMIT",
    "message": "$BUILDKITE_MESSAGE",
    "url": "$BUILDKITE_BUILD_URL"
  },
  "data": $quoted_xml
}
EOF
  )

  curl --request POST \
  --url https://analytics-api.buildkite.com/v1/uploads \
  --header "Authorization: Token token=\"$BUILDKITE_ANALYTICS_BACKEND_TEST_SUITE_API_KEY\";" \
  --header 'Content-Type: application/json' \
  --data "$xml"
}

if [ "$1" == "-h" ]; then
  usage
  exit 1
fi

if [ -n "$1" ]; then
  FILTER_ACTION=$1
  shift
  FILTER_TARGETS=$*
fi

# Display to the user what kind of filtering is happening here
if [ -n "$FILTER_ACTION" ]; then
  echo -e "--- :information_source: \033[0;34mFiltering go tests: $FILTER_ACTION $FILTER_TARGETS\033[0m"
fi

# Buildkite analytics
# TODO is that the best way to handle this?
go install github.com/jstemmer/go-junit-report@latest
asdf reshim golang

# TODO move to manifest
BUILDKITE_ANALYTICS_BACKEND_TEST_SUITE_API_KEY=$(gcloud secrets versions access latest --secret="BUILDKITE_ANALYTICS_BACKEND_TEST_SUITE_API_KEY" --project="sourcegraph-ci" --quiet)

# For searcher
echo "--- comby install"
./dev/comby-install-or-upgrade.sh

# For code insights test
./dev/codeinsights-db.sh &
export CODEINSIGHTS_PGDATASOURCE=postgres://postgres:password@127.0.0.1:5435/postgres
export DB_STARTUP_TIMEOUT=360s # codeinsights-db needs more time to start in some instances.

# We have multiple go.mod files and go list doesn't recurse into them.
find . -name go.mod -exec dirname '{}' \; | while read -r d; do
  pushd "$d" >/dev/null

  # Separate out time for go mod from go test
  echo "--- $d go mod download"
  go mod download

  patterns="${FILTER_TARGETS[*]// /\\|}" # replace spaces with \| to have multiple patterns being matched
  case "$FILTER_ACTION" in
    exclude)
      TEST_PACKAGES=$(go list ./... | { grep -v "$patterns" || true; }) # -v to reject
      if [ -n "$TEST_PACKAGES" ]; then
        echo "--- $d go test"
        go_test "$TEST_PACKAGES"
      else
        echo "--- $d go test (skipping)"
      fi
      ;;
    only)
      TEST_PACKAGES=$(go list ./... | { grep "$patterns" || true; }) # select only what we need
      if [ -n "$TEST_PACKAGES" ]; then
        echo "--- $d go test"
        go_test "$TEST_PACKAGES"
      else
        echo "--- $d go test (skipping)"
      fi
      ;;
    *)
      TEST_PACKAGES="./..."
      echo "--- $d go test"
      go_test "$TEST_PACKAGES"
      ;;
  esac

  popd >/dev/null
done
