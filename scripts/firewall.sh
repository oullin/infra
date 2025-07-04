#!/bin/bash

# Exit on error, undefined variable, or pipeline failure
set -euo pipefail

# Abort if not running as root
if [ "$(id -u)" -ne 0 ]; then
  echo "This script must be run as root. Please use 'sudo'." >&2
  exit 1
fi

# --- Set default policies
ufw default deny incoming
ufw default allow outgoing

# --- Allow essential services
ufw allow OpenSSH
ufw allow http
ufw allow https
ufw allow 443/udp

# --- Enable the firewall and automatically answer "yes" to the prompt
ufw --force enable

# --- Display the final UFW status
ufw status verbose
