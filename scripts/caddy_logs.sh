#!/bin/bash

# A script to create a shared log directory and assign ownership to the current user.
# This avoids Docker permission errors related to user home directories.

# --- Configuration ---
LOG_DIR="/var/log/oullin/caddy"
# Uses the environment variable $USER to get the currently logged-in user.
CURRENT_USER=$USER

# --- Script ---
echo "Setting up log directory: $LOG_DIR"

# Step 1: Create the directory. The -p flag prevents errors if it already exists.
# The command is run with sudo to get root privileges.
echo "--> Creating directory..."
sudo mkdir -p "$LOG_DIR"

# Step 2: Change the directory owner to the current user.
echo "--> Assigning ownership to user '$CURRENT_USER'..."
sudo chown -R "$CURRENT_USER":"$CURRENT_USER" "$LOG_DIR"

echo ""
echo "✅ Setup complete."
echo "The directory '$LOG_DIR' is now ready and owned by '$CURRENT_USER'."
