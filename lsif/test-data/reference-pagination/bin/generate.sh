#!/usr/bin/env bash
#
# This script runs in the parent directory.
# This script is invoked by ./../../generate.sh.
# Do NOT run directly.

DIR=./repos REPO=a ./bin/generate-a.sh

for i in `seq 1 10`; do
    DIR=./repos REPO="b${i}" DEP=`pwd`/repos/a ./bin/generate-b.sh
done
