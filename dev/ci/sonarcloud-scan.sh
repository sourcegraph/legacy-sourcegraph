#!/usr/bin/env bash

set -e
export SONAR_SCANNER_VERSION=4.7.0.2747
export SONAR_SCANNER_HOME=/tmp/sonar/sonar-scanner-$SONAR_SCANNER_VERSION-linux
export SONAR_SCANNER_OPTS="-server"

echo "--- :arrow_down: downloading Sonarcloud binary"
echo ""
curl --fail --create-dirs -sSLo /tmp/sonar-scanner.zip https://binaries.sonarsource.com/Distribution/sonar-scanner-cli/sonar-scanner-cli-$SONAR_SCANNER_VERSION-linux.zip
unzip -o /tmp/sonar-scanner.zip -d /tmp/sonar/

echo "--- :lock: running Sonarcloud scan"
echo ""
$SONAR_SCANNER_HOME/bin/sonar-scanner \
  -Dsonar.organization=sourcegraph \
  -Dsonar.projectKey=sourcegraph_sourcegraph \
  -Dsonar.sources=/root/buildkite/build/sourcegraph/ \
  -Dsonar.host.url=https://sonarcloud.io
