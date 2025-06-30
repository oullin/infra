.PHONY: watch format

ROOT_NETWORK          := oullin_infra
ROOT_PATH             := $(shell pwd)

.PHONY: fresh audit watch format

format:
	gofmt -w -s .

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(ROOT_PATH) && air
