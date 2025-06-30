# Makefile for building and managing the Go application.

# --- Configuration ---
# Include variables from .env file for configuration. This makes variables
# like DOCKER_INFRA_USER available to Make commands.
include .env
export

# Define primary variables. '?=' sets a default value only if the variable is not already set.
# This ensures that the binary name is consistent across all commands.
DOCKER_BINARY_NAME ?= deployment
BINARY_NAME        := $(DOCKER_BINARY_NAME)
DOCKER_IMAGE_NAME  ?= oullin/infra-builder
SERVICE_NAME       ?= deployment
DOCKER_INFRA_USER  ?= infrauser
DOCKER_INFRA_GROUP ?= infragroup

DOCKER_EXTRACTOR_NAME ?= oullin_infra_extractor


# --- Phony Targets ---
# Ensures these targets run even if files with the same name exist.
.PHONY: fresh build-local build run format watch clean clean-extractor

fresh:
	docker compose down --remove-orphans && \
	docker container prune -f && \
	docker image prune -f && \
	docker volume prune -f && \
	docker network prune -f && \
	docker system prune -a --volumes -f && \
	docker ps -aq | xargs --no-run-if-empty docker stop && \
	docker ps -aq | xargs --no-run-if-empty docker rm && \
	docker ps

# --- Build for the local host machine's OS and architecture.
#     Depends on clean-extractor to prevent container name conflicts.
#     Use a robust Debian-based builder for cross-compilation to macOS.
#     The default builder in the Dockerfile remains Alpine for Linux builds.
build-local: clean-extractor
	echo "\nBuilding for the local machine: ($$(go env GOOS)/$$(go env GOARCH))...\n"
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
	echo "\nLocal binary created at: ./bin/$(BINARY_NAME)\n"

# --- Build specifically for the Linux production environment.
#     Depends on clean-extractor to prevent container name conflicts.
build: clean-extractor
	@echo "\mBuilding for Linux (amd64)...\n"
	docker compose build $(SERVICE_NAME)
	docker create --name=$(DOCKER_EXTRACTOR_NAME) $(DOCKER_IMAGE_NAME)
	docker cp $(DOCKER_EXTRACTOR_NAME):/home/$(DOCKER_INFRA_USER)/bin/$(BINARY_NAME) ./bin/$(BINARY_NAME)-linux-amd64
	docker rm -f $(DOCKER_EXTRACTOR_NAME)
	@echo "\nLinux binary for production created at: ./bin/$(BINARY_NAME)"

run: build
	./bin/$(BINARY_NAME)

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
