# Makefile for building and managing the application.

# --- Configuration ---
# Include variables from .env file for configuration. This makes variables
# like DOCKER_INFRA_USER available to Make commands.
-include .env
export

# --- Colours
NC     := \033[0m
BOLD   := \033[1m
CYAN   := \033[36m
WHITE  := \033[37m
GREEN  := \033[0;32m
BLUE   := \033[0;34m
RED    := \033[0;31m
YELLOW := \033[1;33m
# -----------

# Define primary variables.
# This ensures that the binary name is consistent across all commands.
ROOT_PATH             ?= $(shell pwd)
DOCKER_BINARY_NAME    ?= deployment
BINARY_NAME           ?= $(DOCKER_BINARY_NAME)
DOCKER_IMAGE_NAME     ?= oullin/infra-builder
SERVICE_NAME          ?= deployment
DOCKER_INFRA_USER     ?= infrauser
DOCKER_INFRA_GROUP    ?= infragroup
DOCKER_EXTRACTOR_NAME ?= oullin_infra_extractor
API_SUPERVISOR_NAME   ?= oullin-sup

# --- Phony Targets ---
# Ensures these targets run even if files with the same name exist.
.PHONY: fresh build-local build run format watch clean clean-extractor build-test sup-api-status sup-api-restart
.PHONY: ufw-setup ufw-status

fresh:
	make clean && make clean-extractor && \
	docker compose down --remove-orphans && \
	docker container prune -f && \
	docker image prune -f && \
	docker volume prune -f && \
	docker network prune -f && \
	docker system prune -a --volumes -f && \
	docker ps -aq | xargs --no-run-if-empty docker stop && \
	docker ps -aq | xargs --no-run-if-empty docker rm

# --- Build for the local host machine's OS and architecture.
#     Depends on clean-extractor to prevent container name conflicts.
#     Use a robust Debian-based builder for cross-compilation to macOS.
#     The default builder in the Dockerfile remains Alpine for Linux builds.
build-local: clean-extractor
	printf "\nBuilding for the local machine: ($$(go env GOOS)/$$(go env GOARCH))...\n"
	cp $(ROOT_PATH)/.env $(ROOT_PATH)/bin/.env
	docker build \
		--build-arg BUILDER_IMAGE=golang:1.24-bookworm \
		--build-arg BINARY_NAME=$(BINARY_NAME) \
		--build-arg INFRA_USER=$(DOCKER_INFRA_USER) \
		--build-arg INFRA_GROUP=$(DOCKER_INFRA_GROUP) \
		--build-arg GOOS=$(shell go env GOOS) \
		--build-arg GOARCH=$(shell go env GOARCH) \
		-t $(DOCKER_IMAGE_NAME)-local -f ./docker/Dockerfile .
	docker create --name=$(DOCKER_EXTRACTOR_NAME) $(DOCKER_IMAGE_NAME)-local
	docker cp $(DOCKER_EXTRACTOR_NAME):/home/$(DOCKER_INFRA_USER)/bin/$(BINARY_NAME) ./bin/$(BINARY_NAME)
	docker rm -f $(DOCKER_EXTRACTOR_NAME)
	chmod +x ./bin/$(BINARY_NAME)
	printf "\n\e[32mBinary created at:\e[0m \e[1;36m./bin/$(BINARY_NAME)\e[0m\n"

# --- Build specifically for the Linux production environment.
#     Depends on clean-extractor to prevent container name conflicts.
build: clean-extractor
	cp $(ROOT_PATH)/.env $(ROOT_PATH)/bin/.env
	docker compose build $(SERVICE_NAME)
	docker create --name=$(DOCKER_EXTRACTOR_NAME) $(DOCKER_IMAGE_NAME)
	docker cp $(DOCKER_EXTRACTOR_NAME):/home/$(DOCKER_INFRA_USER)/bin/$(BINARY_NAME) ./bin/$(BINARY_NAME)
	docker rm -f $(DOCKER_EXTRACTOR_NAME)
	printf "\n\e[32mBinary created at:\e[0m \e[1;36m./bin/$(BINARY_NAME)\e[0m\n"

run: build
	./bin/$(BINARY_NAME)

# --- Supervisors
# These targets manage the application's supervisor process on the remote server.
#
# IMPORTANT: Both 'sup-api-status' and 'sup-api-restart' invoke 'sudo supervisorctl'.
# For automated deployments (e.g., via GitHub Actions), the SSH user provided
# in the secrets must be configured for passwordless sudo for these specific commands.
#
# If this is not configured, any remote deployment will hang or fail.
# This is typically done by adding a rule to the /etc/sudoers file using 'visudo'.

sup-api-status:
	@sudo supervisorctl status $(API_SUPERVISOR_NAME)

sup-api-restart:
	@sudo supervisorctl restart $(API_SUPERVISOR_NAME)

# --- Firewall (UFW)
ufw-setup:
	@chmod +x $(ROOT_PATH)/scripts/firewall.sh
	$(ROOT_PATH)/scripts/firewall.sh
	@printf "$(GREEN)Firewall properly activated.$(NC)\n"

ufw-status:
	@sudo ufw status verbose

# --- Miscellanious
format:
	gofmt -w -s .

# Watch for file changes and live-reload using 'air'
# https://github.com/air-verse/air
watch:
	air

clean:
	@find ./bin -mindepth 1 ! -name '.gitkeep' -delete

clean-extractor:
	@docker rm -f $(DOCKER_EXTRACTOR_NAME) || true

# ---- Test Commands
build-test:
	POSTGRES_USER_SECRET_PATH="$(POSTGRES_USER_SECRET_PATH)" \
	POSTGRES_PASSWORD_SECRET_PATH="$(POSTGRES_PASSWORD_SECRET_PATH)" \
	POSTGRES_DB_SECRET_PATH="$(POSTGRES_DB_SECRET_PATH)" \
	ENV_DB_USER_NAME="$(ENV_DB_USER_NAME)" \
	ENV_DB_USER_PASSWORD="$(ENV_DB_USER_PASSWORD)" \
	ENV_DB_DATABASE_NAME="$(ENV_DB_DATABASE_NAME)" \
	echo "Done ..."
