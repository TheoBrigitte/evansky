#!/bin/bash

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

BUILD_DIR="$SCRIPT_DIR/../build"

rm -rf $BUILD_DIR
