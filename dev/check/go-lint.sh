#!/bin/bash
set -e
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

export GOBIN="$PWD/.bin"
export PATH=$GOBIN:$PATH
export GO111MODULE=on

go install github.com/golangci/golangci-lint/cmd/golangci-lint

echo go install...
go install -buildmode=archive ./...

golangci-lint run -v
