# make-os.mk — Detect host OS, distribution, version, and architecture
# Supports: Linux (any distro), macOS, Windows (via PowerShell)

VERSION := 0.1.0

.PHONY: detect

detect:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -Command "\
		$ver = (Get-CimInstance Win32_OperatingSystem).Version; \
		$arch = if ($env:PROCESSOR_ARCHITECTURE -match 'ARM') { 'arm64' } else { 'amd64' }; \
		Write-Output \"windows windows $ver $arch\"\
	"
else
	@sh -c '\
		UNAME_S="$$(uname -s)"; \
		UNAME_M="$$(uname -m)"; \
		if [ "$$UNAME_S" = "Darwin" ]; then \
			OS_NAME=macos; \
			DISTRO=macos; \
			VERSION="$$(sw_vers -productVersion)"; \
		else \
			OS_NAME=linux; \
			. /etc/os-release; \
			DISTRO="$$ID"; \
			VERSION="$$VERSION_ID"; \
		fi; \
		if echo $$UNAME_M | grep -qE "arm|aarch64"; then \
			ARCH=arm64; \
		else \
			ARCH=amd64; \
		fi; \
		echo "$$OS_NAME $$DISTRO $$VERSION $$ARCH"; \
	'
endif
