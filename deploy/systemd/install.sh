#!/bin/bash
# Installation script for systemd service files
# Run with: sudo bash deploy/systemd/install.sh

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVICE_DIR="/etc/systemd/system"
INSTALL_DIR="/opt/job-scheduler"

echo "Installing Job Scheduler systemd services..."

# Copy service files
echo "Copying service files to $SERVICE_DIR..."
cp "$SCRIPT_DIR/job-scheduler-api.service" "$SERVICE_DIR/"
cp "$SCRIPT_DIR/job-scheduler-scheduler.service" "$SERVICE_DIR/"
cp "$SCRIPT_DIR/job-scheduler-consumer@.service" "$SERVICE_DIR/"

# Reload systemd
echo "Reloading systemd daemon..."
systemctl daemon-reload

echo ""
echo "Service files installed successfully!"
echo ""
echo "Next steps:"
echo "1. Copy application binary to $INSTALL_DIR/job-scheduler"
echo "2. Create environment file: $INSTALL_DIR/.env"
echo "3. Enable and start services:"
echo ""
echo "   # Enable services"
echo "   sudo systemctl enable job-scheduler-api"
echo "   sudo systemctl enable job-scheduler-scheduler"
echo "   sudo systemctl enable job-scheduler-consumer@1"
echo ""
echo "   # Start services"
echo "   sudo systemctl start job-scheduler-api"
echo "   sudo systemctl start job-scheduler-scheduler"
echo "   sudo systemctl start job-scheduler-consumer@1"
echo ""
echo "   # Check status"
echo "   sudo systemctl status job-scheduler-api"
echo "   sudo systemctl status job-scheduler-scheduler"
echo "   sudo systemctl status job-scheduler-consumer@1"
echo ""
echo "   # To start additional consumers (horizontal scaling)"
echo "   sudo systemctl enable job-scheduler-consumer@2"
echo "   sudo systemctl start job-scheduler-consumer@2"
echo ""
