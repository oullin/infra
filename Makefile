.PHONY: watch format build

include .env

ROOT_NETWORK := oullin_infra
ROOT_PATH := $(shell pwd)
BINARY_FILE_NAME := deployment

format:
	gofmt -w -s .

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(ROOT_PATH) && air

build:
	docker compose build $(BINARY_FILE_NAME)
	docker create --name infra_extractor oullin/infra-builder
	docker cp infra_extractor:/home/$(DOCKER_INFRA_USER)/bin/$(BINARY_FILE_NAME) ./bin/$(BINARY_FILE_NAME)
	docker rm infra_extractor

