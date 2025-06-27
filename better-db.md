# Production-Ready Dockerized PostgreSQL Deployment Guide

## Introduction

This guide provides a comprehensive, step-by-step manual for deploying a secure, robust, and production-ready PostgreSQL database using Docker and Docker Compose on a single VPS.

The principles and configurations outlined here are designed for high-security environments, emphasizing best practices for credential management, data persistence, and automated workflows. This document captures a complete refinement process, providing not just the final solution but the reasoning behind each critical decision, making it ideal for internal documentation and onboarding.

---

## Prerequisites

- A Linux-based VPS (this guide is tailored for Hostinger).
- Docker and Docker Compose installed on the VPS. (Hostinger's "Ubuntu 22.04 with Docker" template is recommended).
- A non-root user with `sudo` privileges.
- Basic familiarity with Docker, Makefiles, and shell commands.

---

## Step 1: Core Principles of a Production Setup

Before writing any code, it's crucial to understand the principles that separate a development setup from a production one. A robust architecture, especially for a stateful service like a database, requires deliberate choices about security and persistence.

### Principle 1: Secure Credential Management (Secrets vs. `.env` files)

- **The Problem:** Storing credentials like database passwords in plain-text `.env` files and injecting them as environment variables is a significant security risk. Any process with access to inspect the container can read these variables.
- **The Solution (Docker Secrets):** Docker Secrets are the industry standard. They mount credentials as files into a secure, in-memory filesystem (`/run/secrets/`) inside the container. The application is configured to read from these files. The credentials never exist as inspectable environment variables in the running container.
- **Our Implementation:** We will use the `_FILE` suffix mechanism (`POSTGRES_PASSWORD_FILE`), which is built into the official PostgreSQL image to read credentials directly from Docker Secret files.

### Principle 2: Robust Data Persistence (Named Volumes vs. Bind Mounts)

- **The Problem:** Bind mounts (e.g., mapping `./data` to `/var/lib/postgresql/data`) tie your critical database data to a specific path within your project structure. This is brittle, prone to host permission issues, and can have performance penalties.
- **The Solution (Named Volumes):** A named volume (e.g., `oullin_db_data`) instructs Docker to manage the data in a dedicated, optimized location on the host's filesystem. This decouples the data from the container's lifecycle and the host's file structure, making it portable, secure, and easy to manage with `docker volume` commands.
- **Our Implementation:** We will use a named volume for all PostgreSQL data.

### Principle 3: Image Specificity and Security (`alpine` vs. `latest`)

- **The Problem:** Using the `latest` tag for an image is unpredictable and can lead to breaking changes during deployment. Standard Debian-based images are large and have a wider attack surface due to the number of included packages.
- **The Solution (`postgres:16-alpine`):** We explicitly pin a major version (`16`) for stability and use the `alpine` variant. Alpine Linux is a minimal distribution, resulting in a significantly smaller and more secure image with fewer potential vulnerabilities.

---

## Step 2: Crafting the `docker-compose.yml`

This file is the heart of our deployment. It is designed to be **environment-agnostic**, supporting both local development and CI/CD pipelines through dynamic variable substitution.

```yaml
version: '3.9'

services:
  # ---------------------------------
  #  PostgreSQL Database Service
  # ---------------------------------
  api-db:
    image: postgres:16-alpine
    container_name: oullin_db
    restart: always
    networks:
      - oullin_net
    environment:
      POSTGRES_USER_FILE: /run/secrets/postgres_user
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
      POSTGRES_DB_FILE: /run/secrets/postgres_db
      PGDATA: /var/lib/postgresql/data/pgdata
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    ports:
      - "127.0.0.1:${ENV_DB_PORT:-5432}:5432"
    volumes:
      - oullin_db_data:/var/lib/postgresql/data
      - ./database/infra/ssl/server.crt:/etc/ssl/certs/server.crt:ro
      - ./database/infra/ssl/server.key:/etc/ssl/private/server.key
      - ./database/infra/config/postgresql.conf:/etc/postgresql/postgresql.conf:ro
      - ./database/infra/scripts/healthcheck.sh:/healthcheck.sh:ro
    healthcheck:
      test: ["CMD", "/healthcheck.sh"]
      interval: 10s
      timeout: 5s
      retries: 5
    command: >
      sh -c "chown postgres:postgres /etc/ssl/private/server.key && chmod 600 /etc/ssl/private/server.key && exec docker-entrypoint.sh -c 'config_file=/etc/postgresql/postgresql.conf'"

  # ---------------------------------
  #  Secure Migration Job Service
  # ---------------------------------
  db-migrate:
    image: migrate/migrate:v4.18.3
    container_name: oullin_db_migrate
    networks:
      - oullin_net
    volumes:
      - ./database/migrations:/migrations
      - ./database/infra/scripts/run-migration.sh:/run-migration.sh
    secrets:
      - postgres_user
      - postgres_password
      - postgres_db
    entrypoint: /run-migration.sh
    command: ""
    depends_on:
      api-db:
        condition: service_healthy
    restart: "no"

  # ---------------------------------
  #  Your Main Application Service
  # ---------------------------------
  api:
    # image: your_app_image:latest
    # build: .
    restart: always
    networks:
      - oullin_net
    depends_on:
      # Ensures the app only starts AFTER migrations are successful.
      - db-migrate
    # ... rest of your application config

# ---------------------------------
#  Top-Level Definitions
# ---------------------------------
volumes:
  oullin_db_data:
    driver: local

secrets:
  # This block is now dynamic. It uses shell environment variables if they exist,
  # otherwise it falls back to the default local path for development.
  postgres_user:
    file: ${POSTGRES_USER_SECRET_PATH:-./database/infra/secrets/postgres_user}
  postgres_password:
    file: ${POSTGRES_PASSWORD_SECRET_PATH:-./database/infra/secrets/postgres_password}
  postgres_db:
    file: ${POSTGRES_DB_SECRET_PATH:-./database/infra/secrets/postgres_db}

networks:
  oullin_net:
    driver: bridge

```

---

## Step 3: Creating Helper Scripts

To achieve maximum security and robustness, we delegate complex logic to small, self-contained shell scripts.

### The Migration Script

This script securely reads secrets and executes the `migrate/migrate` tool.

1.  **Create the file:** `./database/infra/scripts/run-migration.sh`
2.  **Add content:**
    ```bash
    #!/bin/sh
    set -e
    
    # Read credentials securely from Docker Secret files
    DB_USER=$(cat /run/secrets/postgres_user)
    DB_PASSWORD=$(cat /run/secrets/postgres_password)
    DB_NAME=$(cat /run/secrets/postgres_db)
    
    # Construct the database URL using the values from the secrets
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@api-db:5432/${DB_NAME}?sslmode=disable"
    
    # Execute the migrate tool, passing along all arguments from the Makefile
    exec migrate -path /migrations -database "${DATABASE_URL}" "$@"
    ```
3.  **Make it executable:**
    ```bash
    chmod +x ./database/infra/scripts/run-migration.sh
    ```

### The Healthcheck Script

This script provides a reliable way for Docker to check if the database is ready, using the application's actual credentials.

1.  **Create the file:** `./database/infra/scripts/healthcheck.sh`
2.  **Add content:**
    ```bash
    #!/bin/sh
    set -e
    
    DB_USER=$(cat /run/secrets/postgres_user)
    DB_NAME=$(cat /run/secrets/postgres_db)
    
    # Explicitly check if the user variable is empty to prevent errors.
    if [ -z "$DB_USER" ]; then
      echo "Healthcheck Error: The postgres_user secret is empty or could not be read." >&2
      exit 1
    fi
    
    # Execute the final check command.
    exec pg_isready -U "$DB_USER" -d "$DB_NAME"
    ```
3.  **Make it executable:**
    ```bash
    chmod +x ./database/infra/scripts/healthcheck.sh
    ```

---

## Step 4: Creating the `Makefile`

A `Makefile` provides a simple, standardized command interface for all project operations.

Create a file named `Makefile` in your project root:

```makefile
# ==============================================================================
# Makefile for Production-Ready Dockerized Services
# ==============================================================================
.PHONY: help
.PHONY: db:up db:down db:logs db:bash
.PHONY: db:migrate db:rollback db:migrate:create db:migrate:force
.PHONY: db:fresh db:delete db:secure db:seed

# --- Docker & Project Variables
DB_DOCKER_SERVICE_NAME := api-db
DB_DOCKER_CONTAINER_NAME := oullin_db
DB_MIGRATE_SERVICE_NAME := db-migrate

# --- Paths
DB_INFRA_ROOT_PATH := ./database/infra
DB_INFRA_SSL_PATH := $(DB_INFRA_ROOT_PATH)/ssl

# --- SSL Certificate Files
DB_INFRA_SERVER_CRT := $(DB_INFRA_SSL_PATH)/server.crt
DB_INFRA_SERVER_CSR := $(DB_INFRA_SSL_PATH)/server.csr
DB_INFRA_SERVER_KEY := $(DB_INFRA_SSL_PATH)/server.key


# ==============================================================================
# CORE LIFECYCLE COMMANDS
# ==============================================================================

help:
	@echo "Available commands:"
	@echo "  db:up              - Start the database service in detached mode."
	@echo "  db:down            - Stop all services."
	@echo "  db:logs            - Tail the logs of the database container."
	@echo "  db:bash            - Get a bash shell inside the database container."
	@echo "  db:fresh           - Recreate the database from scratch (deletes all data)."
	@echo "  db:delete          - Stop services and DELETE all associated data volumes."
	@echo "  db:secure          - Generate new self-signed SSL certificates."
	@echo "  db:seed            - Run the database seeder (example)."
	@echo "  db:migrate         - Apply all available database migrations."
	@echo "  db:rollback        - Roll back the last applied migration."
	@echo "  db:migrate:create  - Create a new migration file. Usage: make db:migrate:create name=your_migration_name"
	@echo "  db:migrate:force   - Force the database to a specific migration version. Usage: make db:migrate:force version=number"

db:up:
	@echo "--> Starting all services in detached mode..."
	docker compose up -d

db:down:
	@echo "--> Stopping all services..."
	docker compose down

db:logs:
	@echo "--> Tailing logs for $(DB_DOCKER_CONTAINER_NAME)..."
	docker logs -f $(DB_DOCKER_CONTAINER_NAME)

db:bash:
	@echo "--> Opening a bash shell in $(DB_DOCKER_CONTAINER_NAME)..."
	docker exec -it $(DB_DOCKER_CONTAINER_NAME) bash

# ==============================================================================
# SECURE MIGRATION COMMANDS
# ==============================================================================

db:migrate:
	@printf "\n--> Applying all available 'up' migrations...\n"
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) up
	@printf "--> Migration finished.\n\n"

db:rollback:
	@printf "\n--> Rolling back the last applied migration...\n"
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) down 1
	@printf "--> Migration rollback finished.\n\n"

db:migrate:create:
	@echo "--> Creating new migration file named: $(name)"
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) create -ext sql -dir /migrations -seq $(name)

db:migrate:force:
	@printf "\n--> Forcing migration to version $(version)...\n"
	@docker-compose run --rm $(DB_MIGRATE_SERVICE_NAME) force $(version)
	@printf "--> Force migration finished.\n\n"


# ==============================================================================
# SETUP & DESTRUCTIVE COMMANDS
# ==============================================================================

db:fresh:
	@echo "--> Recreating database from a fresh state (all data will be lost)..."
	make db:delete
	make db:up

db:delete:
	@echo "--> Stopping services and PERMANENTLY DELETING associated volumes..."
	docker compose down -v --remove-orphans

db:secure:
	@echo "--> Generating new self-signed SSL certificates..."
	rm -f $(DB_INFRA_SERVER_CRT) $(DB_INFRA_SERVER_CSR) $(DB_INFRA_SERVER_KEY)
	openssl genpkey -algorithm RSA -out $(DB_INFRA_SERVER_KEY)
	openssl req -new -key $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CSR) -subj "/CN=oullin-db-ssl"
	openssl x509 -req -days 365 -in $(DB_INFRA_SERVER_CSR) -signkey $(DB_INFRA_SERVER_KEY) -out $(DB_INFRA_SERVER_CRT)
	@echo "--> SSL certificates created. The container will set its own key permissions on startup."

db:seed:
	@echo "--> Running database seeder (example)..."
	# Example: docker-compose run --rm api go run ./seeder/main.go
	@echo "--> Seeder finished."
```

---

## Step 5: Final Setup Workflow

1.  **Create Secret Files:** For local development, create the files in their default locations.
    ```bash
    mkdir -p ./database/infra/secrets
    printf "your_db_user" > ./database/infra/secrets/postgres_user
    printf "your_strong_password" > ./database/infra/secrets/postgres_password
    printf "your_db_name" > ./database/infra/secrets/postgres_db
    ```

2.  **Create Migration Files:** Place your SQL migration files (e.g., `0001_create_users.up.sql`) in the `./database/migrations/` directory.

3.  **Generate SSL Certificates:** Run the Makefile command.
    ```bash
    make db:secure
    ```

4.  **Deploy:** Run the simple `up` command from the Makefile.
    ```bash
    make db:up
    ```
    Docker Compose will automatically start the database, wait for it to be healthy, run the migrations, and then start your application.

---

## Step 6: Preparing for CI/CD with Dynamic Secret Paths

To make this setup truly automated and environment-agnostic, the hardcoded paths to secret files were made dynamic in Step 2. This allows your CI/CD pipeline to inject secrets from a secure vault without modifying the `docker-compose.yml` file.

### How It Works

The `secrets` block in our final `docker-compose.yml` uses this syntax:
`file: ${POSTGRES_USER_SECRET_PATH:-./database/infra/secrets/postgres_user}`

- **In a CI/CD Pipeline:** Your pipeline should fetch secrets from a vault (e.g., AWS Secrets Manager, GitHub Secrets) and save them to a temporary path. Then, it exports an environment variable before running Docker Compose:
  ```bash
  # Example command in a CI/CD script
  export POSTGRES_USER_SECRET_PATH=/tmp/prod_secrets/db_user
  docker compose up -d
  ```
  Docker Compose will use the path specified by the environment variable.

- **In Local Development:** When you run `make db:up` locally, the `POSTGRES_USER_SECRET_PATH` environment variable is not set. Docker Compose automatically falls back to the default value: `./database/infra/secrets/postgres_user`.

This powerful technique ensures a single, clean `docker-compose.yml` can be used across all environments, from a developer's laptop to a fully automated production deployment.

This completes the guide. You now have a fully documented, secure, and automated system for deploying and managing your PostgreSQL database.
