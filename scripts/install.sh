#!/bin/bash
set -e

INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
BINARY_NAME="netbridge"

echo "Installing NetBridge..."

if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root"
    exit 1
fi

if [ ! -f "build/${BINARY_NAME}" ]; then
    echo "Binary not found. Building..."
    bash scripts/build.sh
fi

cp "build/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

mkdir -p /etc/netbridge/profiles
mkdir -p /etc/netbridge/sessions
mkdir -p /etc/netbridge/cache
mkdir -p /etc/netbridge/logs
mkdir -p /etc/netbridge/state
mkdir -p /etc/netbridge/routes

cp systemd/netbridge.service "${SERVICE_DIR}/"
systemctl daemon-reload
systemctl enable netbridge

echo "NetBridge installed successfully!"
echo "Run 'netbridge' to start."
