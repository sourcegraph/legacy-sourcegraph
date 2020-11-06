#!/usr/bin/env bash

cd "$(dirname "${BASH_SOURCE[0]}")/../.." || exit
set -x

# shellcheck disable=SC1091
source /root/.profile
bash test/setup-deps.sh
bash test/setup-display.sh

# ==========================

CONTAINER=sourcegraph-server

docker_logs() {
  LOGFILE=$(docker inspect ${CONTAINER} --format '{{.LogPath}}')
  cp "$LOGFILE" $CONTAINER.log
  chmod 744 $CONTAINER.log
}

IMAGE=us.gcr.io/sourcegraph-dev/server:$CANDIDATE_VERSION ./dev/run-server-image.sh -d --name $CONTAINER
trap docker_logs exit

sleep 15

go run test/init-server.go

source /root/.profile
pushd client/web || exit
yarn run test:regression:core
yarn run test:regression:integrations
popd || exit

# ==========================

test/cleanup-display.sh
