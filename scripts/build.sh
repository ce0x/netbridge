#!/bin/bash
set -e

BINARY_NAME="netbridge"
BUILD_DIR="build"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS="-s -w -X main.version=${VERSION} -X main.buildTime=${BUILD_TIME}"

echo "Building ${BINARY_NAME}..."
echo "Version: ${VERSION}"

mkdir -p ${BUILD_DIR}

CGO_ENABLED=0 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/${BINARY_NAME} ./cmd/

echo "Build complete: ${BUILD_DIR}/${BINARY_NAME}"
