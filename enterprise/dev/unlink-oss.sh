#!/usr/bin/env bash

set -ex
cd $(dirname "${BASH_SOURCE[0]}")/..

go mod edit -dropreplace=github.com/sourcegraph/sourcegraph
