#!/bin/bash

# ==============================================================================
# Oullin Production Deployment Script
#
# This script is designed to be run on the production VPS. It prepares the
# environment by reading credentials from a secure, non-repository location
# and exporting their paths as environment variables. It then launches the
# Docker Compose services.
#
# This allows the docker-compose.yml file to remain generic and environment-
# agnostic, perfect for CI/CD workflows.
#
# Author: Gus
# ==============================================================================

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
# The absolute path to the directory where secrets are securely stored on the VPS.
# This directory should be outside the project repository and have strict permissions.
SECRETS_DIR="/home/gocanto/.oullin/secrets"
API_DIR="/home/gocanto/Sites/oullin/api"

# --- Pre-flight Checks ---
echo "--> [1/3] Verifying secret files..."

if [ ! -d "$SECRETS_DIR" ]; then
    echo "Error: Secrets directory not found at $SECRETS_DIR"
    exit 1
fi

# Define the full paths to the individual secret files
USER_SECRET_FILE="$SECRETS_DIR/postgres_user"
PASSWORD_SECRET_FILE="$SECRETS_DIR/postgres_password"
DB_SECRET_FILE="$SECRETS_DIR/postgres_db"

# Check that each required secret file exists before proceeding
# Check that each required secret file exists individually for clearer error reporting.
if [ ! -f "$USER_SECRET_FILE" ]; then
    echo "Error: User secret file not found at: $USER_SECRET_FILE"
    exit 1
fi

if [ ! -f "$PASSWORD_SECRET_FILE" ]; then
    echo "Error: Password secret file not found at: $PASSWORD_SECRET_FILE"
    exit 1
fi

if [ ! -f "$DB_SECRET_FILE" ]; then
    echo "Error: Database name secret file not found at: $DB_SECRET_FILE"
    exit 1
fi

echo "--> Secret files verified successfully."

# --- Environment Preparation ---
# Export the variables that our dynamic docker-compose.yml file will use.
# These exports will only be available for the duration of this script.
echo "--> [2/3] Exporting secret paths as environment variables..."

export POSTGRES_USER_SECRET_PATH="$USER_SECRET_FILE"
export POSTGRES_PASSWORD_SECRET_PATH="$PASSWORD_SECRET_FILE"
export POSTGRES_DB_SECRET_PATH="$DB_SECRET_FILE"

echo "--> Environment variables are set."


# --- Deployment Execution ---
echo "--> [3/3] Launching Docker Compose services..."

cd $API_DIR || exit 1

make build:prod

echo ""
echo "--> Deployment initiated successfully!"

