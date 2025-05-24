# make-redis.mk — Idempotent Redis via Podman
# Requires: remake run -f oci://ghcr.io/TrianaLab/make-os:0.1.0 detect
#           remake run -f oci://ghcr.io/TrianaLab/make-podman:0.1.0

VERSION        := 0.1.0
OCI_PODMAN     := oci://ghcr.io/TrianaLab/make-podman:0.1.0
HOST_INFO      := $(shell remake run -f $(OCI_PODMAN) detect)
HOST_OS        := $(word 1,$(HOST_INFO))
DISTRO         := $(word 2,$(HOST_INFO))
OS_VER         := $(word 3,$(HOST_INFO))
ARCH           := $(word 4,$(HOST_INFO))

# Configurable defaults
REDIS_VERSION  ?= 7.0.0
REDIS_PASSWORD ?= redispass
REDIS_PORT     ?= 6379
DATA_DIR       ?= ./data/redis
CONTAINER_NAME ?= redis-server

.PHONY: install run stop status

install:
	@echo "🔍 Ensuring Podman runtime"
	@remake run -f $(OCI_PODMAN) ensure-podman

run: install
	@echo "🚀 Starting Redis $(REDIS_VERSION)"
	@mkdir -p $(DATA_DIR)
	@if podman ps --format '{{.Names}}' | grep -q '^$(CONTAINER_NAME)$$'; then \
		echo "✔ Container $(CONTAINER_NAME) already running"; \
	else \
		podman run -d \
		  --name $(CONTAINER_NAME) \
		  -e REDIS_PASSWORD=$(REDIS_PASSWORD) \
		  -p $(REDIS_PORT):6379 \
		  -v $(realpath $(DATA_DIR)):/data \
		  docker.io/library/redis:$(REDIS_VERSION) \
		  redis-server --requirepass $(REDIS_PASSWORD); \
		echo "✔ Redis started at localhost:$(REDIS_PORT)"; \
	fi

stop:
	@echo "🛑 Stopping Redis"
	@if podman ps --format '{{.Names}}' | grep -q '^$(CONTAINER_NAME)$$'; then \
		podman stop $(CONTAINER_NAME) && podman rm $(CONTAINER_NAME); \
		echo "✔ Container stopped and removed"; \
	else \
		echo "✔ No running container named $(CONTAINER_NAME)"; \
	fi

status:
	@podman ps --filter "name=$(CONTAINER_NAME)" --format "{{.Names}}	{{.Status}}" || echo "Redis not running"