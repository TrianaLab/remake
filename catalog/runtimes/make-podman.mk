# make-podman.mk — Idempotent Podman installer + service manager
# Requires: remake run -f oci://ghcr.io/TrianaLab/make-os:0.1.0 detect

VERSION := 0.1.0

OCI_OS    := oci://ghcr.io/TrianaLab/make-os:0.1.0
HOST_INFO := $(shell remake run -f $(OCI_OS) detect)
HOST_OS   := $(word 1,$(HOST_INFO))
DISTRO    := $(word 2,$(HOST_INFO))
OS_VER    := $(word 3,$(HOST_INFO))
ARCH      := $(word 4,$(HOST_INFO))

.PHONY: install ensure-podman ensure-service status

install: ensure-podman ensure-service
	@echo "✅ Podman ready on $(HOST_OS)/$(DISTRO) $(OS_VER) ($(ARCH))"

ensure-podman:
	@echo "🔍 Checking Podman..."
	@if ! command -v podman >/dev/null 2>&1; then \
		echo "➕ Installing Podman on $(HOST_OS)/$(DISTRO)..."; \
		if [ "$(HOST_OS)" = "linux" ]; then \
			if command -v apt-get >/dev/null 2>&1; then \
				sudo apt-get update && sudo apt-get install -y podman; \
			elif command -v dnf >/dev/null 2>&1; then \
				sudo dnf install -y podman; \
			else \
				echo "Unsupported distro: $(DISTRO)" >&2 && exit 1; \
			fi; \
		elif [ "$(HOST_OS)" = "macos" ]; then \
			brew install podman; \
		elif [ "$(HOST_OS)" = "windows" ]; then \
			powershell -NoProfile -Command "choco install podman -y"; \
		else \
			echo "Unsupported OS: $(HOST_OS)" >&2 && exit 1; \
		fi; \
	else \
		echo "✔ Podman present: $$(podman --version)"; \
	fi

ensure-service:
	@echo "🔍 Checking Podman service..."
	@if [ "$(HOST_OS)" = "linux" ]; then \
		if ! systemctl is-active --quiet podman.socket; then \
			echo "➕ Starting podman.socket"; \
			sudo systemctl start podman.socket; \
		else \
			echo "✔ podman.socket running"; \
		fi; \
	elif [ "$(HOST_OS)" = "macos" ]; then \
		if ! podman system service --time=0 &>/dev/null; then \
			echo "➕ Starting Podman machine"; \
			podman machine start; \
		else \
			echo "✔ Podman machine running"; \
		fi; \
	elif [ "$(HOST_OS)" = "windows" ]; then \
		if ! powershell -NoProfile -Command "(Get-Service podman).Status -eq 'Running'"; then \
			echo "➕ Starting Podman service"; \
			powershell -NoProfile -Command "Start-Service podman"; \
		else \
			echo "✔ Podman service running"; \
		fi; \
	fi

status:
	@podman --version
	@podman info | grep "host"
