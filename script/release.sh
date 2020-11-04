#!/bin/bash

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

BUILD_DIR="$SCRIPT_DIR/../build"

docker build -t builder -f $SCRIPT_DIR/../Dockerfile.builder $SCRIPT_DIR
docker run --rm \
	-v $SCRIPT_DIR/..:/evansky \
	-w /evansky/script \
	-e OS=$OS \
	-e ARCH=$ARCH \
	-e VERSION=$VERSION \
	-e GOCACHE=/tmp/go-build/ \
	--user $(id -u):$(id -g) \
	builder \
	./build.sh
