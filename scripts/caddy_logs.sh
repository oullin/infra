#!/bin/bash
set -euo pipefail

# A script to create a shared log Caddy directory and assign ownership to the current user.
# This avoids Docker permission errors related to user home directories.

# --- Configuration ---
# Example:
#   ls -ld /var/log/oullin
#   ls -ld /var/log/oullin/caddy

LOG_DIR="/var/log/oullin/caddy"

# Uses the environment variable $USER to get the currently logged-in user.
CURRENT_USER=$USER
if [[ -z "$CURRENT_USER" ]]; then
    echo "Unable to determine invoking user (both \$SUDO_USER and \$USER are empty)." >&2
    exit 1
fi

# --- Script ---
echo "Setting up log directory: $LOG_DIR"

# --- Step 1: Create the directory. The -p flag prevents errors if it already exists.
#     The command is ran with sudo to get root privileges.
echo "--> Creating directory..."
sudo mkdir -p "$LOG_DIR"

# Step 2: Change the directory owner to the current user.
echo "--> Assigning ownership to user '$CURRENT_USER'..."
sudo chown -R "$CURRENT_USER":"$CURRENT_USER" "$LOG_DIR"

echo ""
echo "✅ Setup complete."
echo "The directory '$LOG_DIR' is now ready and owned by '$CURRENT_USER'."
