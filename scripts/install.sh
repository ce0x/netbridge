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

echo ""
echo "NetBridge binary installed successfully!"
echo ""

# Interactive core installation
echo "Which core backends would you like to install?"
echo "  1) xray (VLESS/VMess/Trojan/Shadowsocks) [Recommended]"
echo "  2) sing-box (alternative proxy platform)"
echo "  3) wireguard-tools (WireGuard VPN)"
echo "  4) openvpn (OpenVPN)"
echo "  5) All of the above"
echo "  6) Skip (install later with 'netbridge core install')"
echo ""
read -p "Enter choice [1-6, default=5]: " choice

case "${choice}" in
    1)
        echo "Installing xray..."
        ${INSTALL_DIR}/${BINARY_NAME} core install xray
        ;;
    2)
        echo "Installing sing-box..."
        ${INSTALL_DIR}/${BINARY_NAME} core install sing-box
        ;;
    3)
        echo "Installing wireguard-tools..."
        ${INSTALL_DIR}/${BINARY_NAME} core install wireguard-tools
        ;;
    4)
        echo "Installing openvpn..."
        ${INSTALL_DIR}/${BINARY_NAME} core install openvpn
        ;;
    5|"")
        echo "Installing all cores..."
        ${INSTALL_DIR}/${BINARY_NAME} core install --all
        ;;
    6)
        echo "Skipping core installation. Run 'netbridge core install --all' later."
        ;;
    *)
        echo "Invalid choice. Skipping core installation."
        ;;
esac

echo ""
echo "Installation complete!"
echo "Run 'netbridge --help' to see available commands."
echo "Run 'netbridge core status' to check installed cores."
