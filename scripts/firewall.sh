#!/bin/bash

set -e

# --- Set default policies
sudo ufw default deny incoming
sudo ufw default allow outgoing

# --- Allow essential services
sudo ufw allow OpenSSH
sudo ufw allow http
sudo ufw allow https
sudo ufw allow 443/udp

# --- Enable the firewall and automatically answer "yes" to the prompt
yes | sudo ufw enable

# --- Display the final UFW status
sudo ufw status verbose
