#!/bin/bash
#
# This script requires OS, ARCH, and VERSION environement variables to be set.

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

test -t 1 && USE_TTY="-it" || USE_TTY=""

docker run --rm \
	-v $SCRIPT_DIR/..:/app \
	-w /app \
	-e OS=$OS \
	-e ARCH=$ARCH \
	-e VERSION=$VERSION \
	-e GOCACHE=/tmp/go-build/ \
	--user $(id -u):$(id -g) \
	${USE_TTY} \
	builder \
	$@
