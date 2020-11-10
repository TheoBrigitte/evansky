#!/bin/bash
#
# This script requires VERSION environement variables to be set.

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

echo $SCRIPT_DIR

echo "> preparing build image ... builder"
docker build -t builder -f $SCRIPT_DIR/../Dockerfile.builder $SCRIPT_DIR
docker run --rm \
	-v $SCRIPT_DIR/..:/app \
	-w /app \
	-e VERSION=$VERSION \
	-e GOCACHE=/tmp/go-build/ \
	--user $(id -u):$(id -g) \
	-ti \
	builder \
	./script/release.sh
