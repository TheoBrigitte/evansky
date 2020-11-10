#!/bin/bash
#
# This script requires ARCH, and VERSION environement variables to be set.

set -eu

SCRIPT_DIR="$( cd "$( dirname "$0" )" && pwd )"

PROJECT_NAME=evansky
OS=linux
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

## debian package
# taken from: https://github.com/sharkdp/bat/blob/master/.github/workflows/CICD.yml
echo "> packaging .deb file ... build/${PROJECT_NAME}.deb"
install -Dm755 ${BINARY} ${DPKG_DIR}/usr/bin/${PROJECT_NAME}
# TODO: unsure about bash
install -Dm644 ${ARCHIVE_DIR}/autocomplete/${PROJECT_NAME}.bash ${DPKG_DIR}/usr/share/bash/vendor-completions/${PROJECT_NAME}.bash
install -Dm644 ${ARCHIVE_DIR}/autocomplete/${PROJECT_NAME}.zsh ${DPKG_DIR}/usr/share/zsh/vendor-completions/_${PROJECT_NAME}
install -Dm644 ${ARCHIVE_DIR}/autocomplete/${PROJECT_NAME}.fish ${DPKG_DIR}/usr/share/zsh/vendor_completions.d/${PROJECT_NAME}.fish
install -Dm644 ${PROJECT_ROOT}/README.md ${DPKG_DIR}/usr/share/doc/${PROJECT_NAME}/README.md
install -Dm644 ${PROJECT_ROOT}/LICENSE ${DPKG_DIR}/usr/share/doc/${PROJECT_NAME}/LICENSE
install -Dm644 ${PROJECT_ROOT}/CHANGELOG.md ${DPKG_DIR}/usr/share/doc/${PROJECT_NAME}/changelog
gzip -f -n --best ${DPKG_DIR}/usr/share/doc/${PROJECT_NAME}/changelog
cat > ${DPKG_DIR}/usr/share/doc/${PROJECT_NAME}/copyright <<EOF
Format: http://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: ${PROJECT_NAME}
Source: https://github.com/TheoBrigitte/evansky

Files: *
Copyright: 2020 Théo Brigitte <theo.brigitte@gmail.com>
License: MIT
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:

 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
EOF
chmod 644 ${DPKG_DIR}/usr/share/doc/${PROJECT_NAME}/copyright
mkdir -p "${DPKG_DIR}/DEBIAN"
cat > "${DPKG_DIR}/DEBIAN/control" <<EOF
Package: ${PROJECT_NAME}
Version: ${VERSION}
Section: utils
Priority: optional
Maintainer: Théo Brigitte
Homepage: https://github.com/TheoBrigitte/evansky
Architecture: ${ARCH}
Provides: ${PROJECT_NAME}
Description: evansky media renamer
 Rename media files in order to be detected by media server like Jellyfin.
EOF

fakeroot dpkg-deb --build ${DPKG_DIR} ${BUILD_DIR}/${PROJECT_NAME}_${VERSION}_${ARCH}.deb
