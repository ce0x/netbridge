#!/bin/bash
set -e

echo "Setting up NetBridge development environment..."

if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed"
    echo "Install from https://golang.org/dl/"
    exit 1
fi

go version

echo "Downloading dependencies..."
go mod download

echo "Running tests..."
go test ./...

echo "Development environment ready!"
