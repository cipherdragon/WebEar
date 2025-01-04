#!/bin/bash
set -e

BINARY_NAME="webear"
BINARY_PATH="/usr/local/bin/${BINARY_NAME}"
SERVICE_NAME="${BINARY_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
CONFIG_DIR="/etc/${BINARY_NAME}"

if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root"
    exit 1
fi

# Stop and disable service
systemctl stop "${SERVICE_NAME}" || true
systemctl disable "${SERVICE_NAME}" || true

# Remove files
rm -f "${SERVICE_PATH}"
rm -f "${BINARY_PATH}"
rm -rf "${CONFIG_DIR}"

# Reload systemd
systemctl daemon-reload

echo "Uninstallation complete"