#!/bin/bash
set -e  # Exit on any error

# Configuration
BINARY_NAME="webear"
BINARY_PATH="/usr/local/bin/${BINARY_NAME}"
SERVICE_NAME="${BINARY_NAME}.service"
SERVICE_PATH="/etc/systemd/system/${SERVICE_NAME}"
CONFIG_DIR="/etc/${BINARY_NAME}"

# Must run as root
if [ "$EUID" -ne 0 ]; then 
    echo "Please run as root"
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo_step() {
    echo -e "${GREEN}==>${NC} $1"
}

echo_error() {
    echo -e "${RED}Error:${NC} $1"
}

# Create directories
echo_step "Creating necessary directories..."
mkdir -p "${CONFIG_DIR}"

# Copy binary
echo_step "Installing binary..."
if [ ! -f "./webear" ]; then
    echo_error "webear binary not found in current directory"
    exit 1
fi
cp ./webear "${BINARY_PATH}"
chmod 755 "${BINARY_PATH}"
chown root:root "${BINARY_PATH}"

# Create systemd service file
echo_step "Creating systemd service..."
cat > "${SERVICE_PATH}" << EOL
[Unit]
Description=WebEar Service
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
User=root
Group=root
ExecStart=${BINARY_PATH}
WorkingDirectory=${CONFIG_DIR}

# Environment setup
Environment=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
Environment=USER=root
Environment=HOME=/root

# Logging
StandardOutput=journal+console
StandardError=journal+console

# Restart configuration
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOL

# Set correct SELinux context
echo_step "Configuring SELinux..."
if command -v semanage &> /dev/null; then
    semanage fcontext -a -t bin_t "${BINARY_PATH}"
    restorecon -v "${BINARY_PATH}"
else
    echo_error "SELinux tools not found. Please install policycoreutils-python-utils"
    exit 1
fi

# Reload systemd and enable service
echo_step "Configuring systemd service..."
systemctl daemon-reload
systemctl enable "${SERVICE_NAME}"

# Start the service
echo_step "Starting service..."
systemctl start "${SERVICE_NAME}"

# Check service status
echo_step "Checking service status..."
if systemctl is-active --quiet "${SERVICE_NAME}"; then
    echo -e "${GREEN}Installation successful!${NC}"
    echo "Service is running."
    echo "You can check the status with: systemctl status ${SERVICE_NAME}"
    echo "View logs with: journalctl -u ${SERVICE_NAME} -f"
else
    echo_error "Service failed to start. Check logs with: journalctl -u ${SERVICE_NAME}"
    exit 1
fi

# Print helpful information
echo
echo "Installation Summary:"
echo "-------------------"
echo "Binary location: ${BINARY_PATH}"
echo "Config directory: ${CONFIG_DIR}"
echo "Service file: ${SERVICE_PATH}"
echo
echo "Useful commands:"
echo "  - View logs: journalctl -u ${SERVICE_NAME} -f"
echo "  - Check status: systemctl status ${SERVICE_NAME}"
echo "  - Restart service: systemctl restart ${SERVICE_NAME}"