#!/bin/bash

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

echo "> preparing build image ... builder"
docker build -t builder -f $SCRIPT_DIR/../Dockerfile.builder $SCRIPT_DIR
