#!/usr/bin/env bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")

parallel_run() {
    ../ci/parallel_run.sh "$@"
}

go version
go env

CHECKS=(
    ./yarn-deduplicate.sh \
    ./docsite.sh \
    ./gofmt.sh \
    ./template-inlines.sh \
    ./go-enterprise-import.sh \
    ./go-dbconn-import.sh \
    ./go-generate.sh \
    ./go-lint.sh \
    ./todo-security.sh \
    ./no-localhost-guard.sh \
    ./bash-syntax.sh \
    ./check-owners.sh
)

echo "--- 🚨 Buildkite's timing information is misleading! Only consider the job timing that's printed after 'done'"

parallel_run {} ::: "${CHECKS[@]}"

# TODO(sqs): Reenable this check when about.sourcegraph.com is reliable. Most failures come from its
# downtime, not from broken URLs.
#
# ./broken-urls.bash
