#!/bin/bash

# Define variables
SERVICE_NAME="solo_client.service"
SERVICE_PATH="/etc/systemd/system/$SERVICE_NAME"
DAEMON_BINARY="/usr/local/sbin/solo"

# Check if the daemon binary exists
if [ ! -f "$DAEMON_BINARY" ]; then
	echo "Daemon binary not found at $DAEMON_BINARY. Please check the path."
	exit 1
fi

# Copy the service file to the systemd directory
echo "Installing the systemd service..."
sudo cp $SERVICE_NAME $SERVICE_PATH

# Reload systemd to recognize the new service
echo "Reloading systemd daemon..."
sudo systemctl daemon-reload

# Enable the service to start on boot
echo "Enabling the service..."
sudo systemctl enable $SERVICE_NAME

# Start the service immediately
echo "Starting the service..."
sudo systemctl start $SERVICE_NAME

# Show the status of the service
sudo systemctl status $SERVICE_NAME

