## VPS & Docker App Deployment: A Complete Guide

This document is a comprehensive, step-by-step tutorial for deploying and managing a Docker-based application on a new 
VPS. It covers the entire process from initial server access and troubleshooting to creating a resilient, auto-restarting 
service that is straightforward to manage.

This guide consolidates the solutions to a series of real-world challenges, providing a proven roadmap for a robust production setup.

---

## Table of Contents
1.  [**Chapter 1: Initial VPS & GitHub SSH Setup**](#chapter-1-initial-vps--github-ssh-setup)
    - Solving `Permission denied (publickey)`
2.  [**Chapter 2: Solving Docker & Makefile Permissions**](#chapter-2-solving-docker--makefile-permissions)
    - 2.1. Deleting Docker-Owned Volumes
    - 2.2. Fixing SSL Certificate Permissions
    - 2.3. Correcting Private Key Ownership
3.  [**Chapter 3: Process Resiliency with Supervisor**](#chapter-3-process-resiliency-with-supervisor)
    - 3.1. Installing Supervisor
    - 3.2. Creating the Service Configuration
    - 3.3. Activating the Service
4.  [**Chapter 4: Convenient Management with a Makefile**](#chapter-4-convenient-management-with-a-makefile)
    - 4.1. Creating Robust Log Commands
    - 4.2. The Final Management Makefile

---

## Chapter 1: Initial VPS & GitHub SSH Setup

When you first try to clone a private repository from a new VPS, you will almost certainly encounter a permission error. 
This is because your VPS has no "ID" that GitHub recognizes.

**The Problem: `Permission denied (publickey)`**
```
git@github.com: Permission denied (publickey).
fatal: Could not read from remote repository.
```

**The Solution: Create an SSH key on the VPS and add it to your GitHub account.**

### Step 1.1: Generate a New SSH Key
Log into your VPS and run the following command. Replace the email with your own.
```bash
# Using ED25519 is modern and secure
ssh-keygen -t ed25519 -C "your_email@example.com"
```
Press `Enter` through all the prompts to accept the default file location and to create a key without a passphrase.

### Step 1.2: Display and Copy the Public Key
Use the `cat` command to display your new public key.
```bash
cat ~/.ssh/id_ed25519.pub
```
Copy the entire output, which starts with `ssh-ed25519...`.

### Step 1.3: Add the Key to GitHub
1.  Go to your GitHub **Settings**.
2.  Navigate to **SSH and GPG keys**.
3.  Click **New SSH key**.
4.  Give it a recognizable title (e.g., "Hostinger VPS").
5.  Paste the key you copied into the "Key" field and save.

### Step 1.4: Test the Connection
Verify that the setup works.
```bash
ssh -T git@github.com
```
A success message confirms you can now interact with your repositories.

---

## Chapter 2: Solving Docker & Makefile Permissions

When using Docker with bind mounts, where you link a host directory to a container directory, you will encounter complex 
permission issues. This section solves them in order.

### 2.1. Deleting Docker-Owned Volumes

**The Problem:** A `Makefile` script fails when trying to delete a data directory that was used by a Docker container.
```
rm: cannot remove '/home/gocanto/Sites/oullin/api/database/infra/data': Permission denied
```
**The Reason:** The Docker daemon's user (often `root`) takes ownership of the host directory (`.../data`) so it can write 
to it from inside the container. Your host user (`gocanto`) no longer has permission to delete it.

**The Solution:** Use `sudo` in your `Makefile` for the deletion command.

**Example `Makefile` Target:**
```makefile
db\:delete:
    docker compose down $(DB_DOCKER_SERVICE_NAME) --remove-orphans && \
    sudo rm -rf $(DB_INFRA_DATA_PATH) && \
    docker ps
```

### 2.2. Fixing SSL Certificate Permissions

**The Problem:** The Postgres container fails to start, citing a permissions error on the SSL certificate.
```
FATAL: could not load server certificate file "/etc/ssl/certs/server.crt": Permission denied
```
**The Reason:** The public certificate file (`.crt`) on the host has permissions that are too strict (e.g., `600`). 
The `postgres` user inside the container is not the owner and therefore cannot read it.

**The Solution:** Adjust the `Makefile` command that sets permissions. A public certificate must be world-readable (`644`), 
while a private key must remain secret (`600`).

**Example `Makefile` Target:**
```makefile
db\:chmod:
    sudo chmod 600 $(DB_INFRA_SERVER_KEY) && sudo chmod 644 $(DB_INFRA_SERVER_CRT)
```

### 2.3. Correcting Private Key Ownership

**The Problem:** After fixing the certificate permission, Postgres still fails with a new error related to the private key.
```
FATAL: private key file "/etc/ssl/private/server.key" must be owned by the database user or root
```
**The Reason:** This is a security feature. Postgres will not start if the private key file is owned by an untrusted 
user *inside the container*. Your mounted file retains its host ownership (`gocanto`), which Postgres doesn't trust.

**The Solution:** Override the container's startup command in your `docker-compose.yml` to change the key's ownership 
before starting Postgres.

**Example `docker-compose.yml` Service Definition:**
```yaml
services:
  api-db:
    image: postgres:latest
    environment:
      # Your ENV variables here
    volumes:
      - ./database/infra/ssl/server.crt:/etc/ssl/certs/server.crt
      - ./database/infra/ssl/server.key:/etc/ssl/private/server.key
      - ./database/infra/data:/var/lib/postgresql/data
      # ... other volumes
    # This command fixes ownership, then runs the normal Postgres entrypoint
    command: >
      sh -c "chown postgres:postgres /etc/ssl/private/server.key && chmod 600 /etc/ssl/private/server.key && exec docker-entrypoint.sh postgres"
```

---

## Chapter 3: Process Resiliency with Supervisor

To ensure your application automatically restarts after a crash or server reboot, you need a process manager like `supervisor`.

### 3.1. Installing Supervisor
```bash
sudo apt-get update
sudo apt-get install supervisor
```

### 3.2. Creating the Service Configuration
Create a new configuration file for your service. We use a clean, comment-free version to avoid potential syntax errors 
from IDE plugins (like GoLand's Nginx plugin) that can misread `.conf` files.

**File: `/etc/supervisor/conf.d/oullin-api.conf`**
```ini
[program:oullin-api]
command=/usr/bin/docker compose up
directory=/home/gocanto/Sites/oullin/api
user=gocanto
autostart=true
autorestart=true
startsecs=10
startretries=3
stopsignal=INT
stopwaitsecs=30
killasgroup=true
stdout_logfile=/var/log/supervisor/oullin-api.log
stderr_logfile=/var/log/supervisor/oullin-api.err.log
stdout_logfile_maxbytes=10MB
stdout_logfile_backups=5
```

### 3.3. Activating the Service
Tell Supervisor to load and run your new configuration.

**Step 1: Reread config files**
```bash
sudo supervisorctl reread
```

**Step 2: Apply changes and start the service**
```bash
sudo supervisorctl update
```

**Step 3: Check the status**
```bash
sudo supervisorctl status
```

---

## Chapter 4: Convenient Management with a Makefile

To maintain a simple and consistent workflow, we can wrap the Supervisor commands in a `Makefile`.

### 4.1. Creating Robust Log Commands

**The Problem:** The standard `supervisorctl tail` command can crash with an XML-RPC error on high-volume logs.
```
error: <class 'xml.parsers.expat.ExpatError'>, not well-formed (invalid token)
```
**The Solution:** Instead of using `supervisorctl tail`, we use the more robust Linux command `tail -f` to read the log files directly.

### 4.2. The Final Management Makefile
Place this `Makefile` in your project directory for easy access to all essential commands for managing your now-supervised service.

**File: `Makefile`**
```makefile
# =================================================================
# Makefile for managing services via Supervisor
# =================================================================

# --- Variables ---
# The name of the program as defined in your .conf file.
SERVICE_NAME := oullin-api


# --- Self-documentation ---
# This target will display the available commands.
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Service Management Commands:"
	@echo "  status       Check the status of the $(SERVICE_NAME) service."
	@echo "  start        Start the service."
	@echo "  stop         Stop the service."
	@echo "  restart      Restart the service."
	@echo ""
	@echo "Logging Commands:"
	@echo "  logs         View the live stdout logs for the service."
	@echo "  logs-err     View the live stderr logs for the service."


# --- Process Control ---
.PHONY: status start stop restart

status:
	@echo "--> Checking status for [$(SERVICE_NAME)]..."
	@sudo supervisorctl status $(SERVICE_NAME)

start:
	@echo "--> Starting [$(SERVICE_NAME)]..."
	@sudo supervisorctl start $(SERVICE_NAME)

stop:
	@echo "--> Stopping [$(SERVICE_NAME)]..."
	@sudo supervisorctl stop $(SERVICE_NAME)

restart:
	@echo "--> Restarting [$(SERVICE_NAME)]..."
	@sudo supervisorctl restart $(SERVICE_NAME)


# --- Logging ---
# Using `tail -f` directly on the log files is more robust than `supervisorctl tail`.
.PHONY: logs logs-err

logs:
	@echo "--> Tailing stdout logs for [$(SERVICE_NAME)]... (Press Ctrl+C to exit)"
	@sudo tail -f /var/log/supervisor/$(SERVICE_NAME).log

logs-err:
	@echo "--> Tailing stderr logs for [$(SERVICE_NAME)]... (Press Ctrl+C to exit)"
	@sudo tail -f /var/log/supervisor/$(SERVICE_NAME).err.log
