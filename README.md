[![Build Status](https://github.com/TrianaLab/remake/actions/workflows/ci.yml/badge.svg)](https://github.com/TrianaLab/remake/actions)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/TrianaLab/remake)](https://pkg.go.dev/github.com/TrianaLab/remake)
[![Go Report Card](https://goreportcard.com/badge/github.com/TrianaLab/remake)](https://goreportcard.com/report/github.com/TrianaLab/remake)
[![codecov](https://codecov.io/gh/TrianaLab/remake/graph/badge.svg?token=DI2AL1DL9T)](https://codecov.io/gh/TrianaLab/remake)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# Remake CLI 🚀

Remake is a powerful CLI tool that packages, distributes, and runs Makefiles as OCI artifacts. It enables centralized management of build scripts, seamless integration into CI/CD pipelines, and consistent execution across environments.

## 🗂️ Catalog

This repository includes a **Makefile catalog** under `catalog/`, providing reusable OCI-published Makefiles organized by category:

* **Runtimes** (`catalog/runtimes/`):

  * `make-os.mk`: Detect host OS, distribution, version, and architecture.
  * `make-podman.mk`: Install and start Podman runtime from upstream binaries.
* **Services** (`catalog/services/`):

  * `make-redis.mk`: Spin up a Redis container via Podman with configurable password, port, and data volume.

Each Makefile declares its own `VERSION := x.y.z` and bootstraps any lower-level dependencies automatically (OS detection, Podman). Simply invoke the highest-level Makefile:

```bash
remake run -f oci://ghcr.io/TrianaLab/make-os:latest detect
# Or if you feel lazy today, you can just run:
remake run -f trianalab/make-os detect
```

## 🌟 Benefits

* **CI/CD Friendly**: Easily integrate Makefile-based workflows into modern CI/CD systems. Keep your build logic versioned and reproducible across pipelines.
* **Centralized Makefiles**: Store your Makefiles in container registries for a single source of truth. Share and reuse build definitions across teams.
* **Caching & Performance**: Local cache reduces redundant downloads, speeding up builds and reducing registry load.
* **Versioning & Rollback**: Tag your build scripts via OCI tags. Roll back to previous versions with zero hassle.
* **Secure Distribution**: Leverage OCI registry authentication and transport security for safe artifact delivery.

## 🛠️ Installation

### Via Installer Script

```bash
curl -fsSL https://raw.githubusercontent.com/TrianaLab/remake/main/scripts/get-remake.sh | bash
```

This command downloads and runs an installation script that:

1. Fetches the latest `remake` binary.
2. Places it in under `/usr/local/bin`.

### Go Installation

Make sure your Go `bin` directory is in your `PATH`:

```bash
go install github.com/TrianaLab/remake@latest
```

Alternatively, clone and build from source:

```bash
git clone https://github.com/TrianaLab/remake.git
cd remake
make install
```

## 📚 Usage

All commands share the same global options and configuration (default: `~/.remake/config.yaml`).

### 🔒 Login

Authenticate with an OCI registry (default: `ghcr.io`).

```bash
remake login [registry] -u <username> -p <password>
```

* `registry`: Optional OCI host (e.g., `docker.io`).
* Prompts for missing credentials interactively.

### 📦 Push

Upload a local Makefile to an OCI registry, tagging it as an artifact.

```bash
remake push <registry/repo:tag> [-f <path>]
```

* `<registry/repo:tag>`: e.g., `ghcr.io/myorg/myrepo:1.0.0`.
* `-f`: Path to Makefile (default: `makefile`).

### 📥 Pull

Download and display a Makefile artifact.

```bash
remake pull <registry/repo:tag> [--no-cache]
```

* `--no-cache`: Force re-download, bypassing local cache.

### 🏃 Run

Execute targets from a local or remote Makefile artifact.

```bash
remake run [targets...] [-f <path|registry/repo:tag>] [--make-flag <flag>] [--no-cache]
```

* `targets`: One or more Makefile targets.
* `-f`: Specify Makefile path or OCI reference.
* `--make-flag`: Pass flags to the `make` command (can be repeated).

### ⚙️ Config

Print the current configuration (registry, cache directory, credentials).

```bash
remake config
```

Redirect to file to export settings:

```bash
remake config > config.yaml
```

### 📄 Version

Show the installed Remake CLI version.

```bash
remake version
```

## 🤝 Contributing

1. Fork the repository.
2. Create a new branch: `git checkout -b feature/<YourFeature>`.
3. Add or update Makefile artifacts under `catalog/`, setting `VERSION := x.y.z`.
4. Ensure new/updated `*.mk` includes `install`, `status`/`run` targets.
5. Commit and push—CI will publish artifacts automatically to GHCR.
6. Open a Pull Request describing your changes.

## 📜 License

MIT © TrianaLab. See the [LICENSE](LICENSE) file for details.


## 📝 NOTICE

See [NOTICE.md](NOTICE.md) for additional legal and licensing details.
