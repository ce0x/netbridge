#!/bin/bash
set -e

VERSION=${1:-"v0.1.0"}
PLATFORMS="linux/amd64 linux/arm64"

echo "Building release ${VERSION}..."

for platform in ${PLATFORMS}; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    output="build/netbridge-${VERSION}-${GOOS}-${GOARCH}"
    
    echo "  Building ${GOOS}/${GOARCH}..."
    CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
        -ldflags "-s -w -X main.version=${VERSION}" \
        -o ${output} ./cmd/
    
    tar -czf "${output}.tar.gz" -C build $(basename ${output})
done

echo "Release ${VERSION} built successfully!"
ls -la build/*.tar.gz
