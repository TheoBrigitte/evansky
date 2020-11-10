#!/bin/bash
#
# This script requires VERSION environement variables to be set.

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

OS=linux  ARCH=amd64   $SCRIPT_DIR/build.sh $SCRIPT_DIR/package_debian.sh
OS=linux  ARCH=386     $SCRIPT_DIR/build.sh $SCRIPT_DIR/package_debian.sh
OS=linux  ARCH=arm64   $SCRIPT_DIR/build.sh $SCRIPT_DIR/package_debian.sh
OS=linux  ARCH=arm     $SCRIPT_DIR/build.sh $SCRIPT_DIR/package_debian.sh
OS=darwin ARCH=amd64   $SCRIPT_DIR/build.sh
