#!/bin/bash
#
# This script requires OS, ARCH, and VERSION environement variables to be set.

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

BUILD_DIR="$SCRIPT_DIR/../build"

docker build -t builder -f $SCRIPT_DIR/../Dockerfile.builder $SCRIPT_DIR
docker run --rm \
	-v $SCRIPT_DIR/..:/app \
	-w /app \
	-e OS=$OS \
	-e ARCH=$ARCH \
	-e VERSION=$VERSION \
	-e GOCACHE=/tmp/go-build/ \
	--user $(id -u):$(id -g) \
	-ti \
	builder \
	./script/build.sh
