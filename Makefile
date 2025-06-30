ROOT_NETWORK          := oullin_infra
ROOT_PATH             := $(shell pwd)

watch:
	# --- Works with (air).
	#     https://github.com/air-verse/air
	cd $(ROOT_PATH) && air
