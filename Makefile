.PHONY: watch format build

ROOT_NETWORK          := oullin_infra
ROOT_PATH             := $(shell pwd)

format:
	gofmt -w -s .

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(ROOT_PATH) && air

build:
	docker compose build deployment
	docker create --name infra_extractor oullin/infra-builder
	docker cp infra_extractor:/home/$(DOCKER_INFRA_USER)/bin/infra ./infra
	docker rm infra_extractor
