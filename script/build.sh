#!/bin/bash
#
# This script requires OS, ARCH, and VERSION environement variables to be set.

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

PROJECT_NAME=evansky
PKG_NAME=${PROJECT_NAME}-${VERSION}-${OS}-${ARCH}

PROJECT_ROOT=$SCRIPT_DIR/..
BUILD_DIR=$PROJECT_ROOT/build
BIN_DIR=$BUILD_DIR/bin
ARCHIVE_DIR=$BUILD_DIR/${PKG_NAME}
DPKG_DIR=${BUILD_DIR}/dpkg

mkdir -p $BUILD_DIR $BIN_DIR $ARCHIVE_DIR

## Build binary
# -s: strip symbols
# -w: strip debug symbols
BINARY_NAME=${PROJECT_NAME}_${VERSION}_${OS}_${ARCH}
BINARY=${BIN_DIR}/${BINARY_NAME}
BINARY_RUNNER=${BIN_DIR}/${PROJECT_NAME}
echo "> compiling binary ... build/$BINARY_NAME"
GOOS=$OS GOARCH=$ARCH CGO_ENABLED=0 go build -v -ldflags "-s -w" -o $BINARY $PROJECT_ROOT
CGO_ENABLED=0 go build -v -ldflags "-s -w" -o $BINARY_RUNNER $PROJECT_ROOT

## base archive
echo "> packaging archive ... build/${PKG_NAME}.tar.gz"
cp $BINARY $ARCHIVE_DIR/${PROJECT_NAME}
cp $PROJECT_ROOT/README.md $PROJECT_ROOT/LICENSE $PROJECT_ROOT/CHANGELOG.md $ARCHIVE_DIR
mkdir -p $ARCHIVE_DIR/autocomplete
$BINARY_RUNNER completion bash > $ARCHIVE_DIR/autocomplete/${PROJECT_NAME}.bash
$BINARY_RUNNER completion zsh  > $ARCHIVE_DIR/autocomplete/${PROJECT_NAME}.zsh
$BINARY_RUNNER completion fish > $ARCHIVE_DIR/autocomplete/${PROJECT_NAME}.fish
pushd $BUILD_DIR > /dev/null
tar czf ${PKG_NAME}.tar.gz ${PKG_NAME}/*
popd > /dev/null
