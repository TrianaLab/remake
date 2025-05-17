# E2E Tests for Remake CLI

This directory contains end-to-end (E2E) tests to validate the full functionality of the `remake` CLI, including:

- **login** (OCI registry authentication)
- **publish** (push Makefile as OCI artifact)
- **pull** (fetch remote Makefiles)
- **run** (execute Makefiles with local, HTTP, and OCI includes)
- **version** (CLI version check)

## Directory Structure

```
e2e-tests/
├── fixtures/
│   ├── ci.mk        # Reusable CI module
│   ├── local.mk     # Simple local Makefile
│   └── http.mk      # HTTP-served Makefile
├── tests.sh         # Bash script to run all scenarios
├── Makefile         # Orchestrates individual E2E steps
└── README.md        # This documentation
```

## Prerequisites

- `remake` installed and on your `$PATH` (see project root `make install`).
- **Python 3** (for serving HTTP fixtures).
- GitHub PAT with **write:packages** scope.

Set environment variables:
```bash
export GITHUB_USER=<your-username>
export GITHUB_TOKEN=<your-pat>
```

## Usage

1. **Serve HTTP fixtures** (in a separate shell):
   ```bash
   cd fixtures
   python3 -m http.server 8000
   ```

2. **Run all E2E tests**:
   - Via script:
     ```bash
     cd e2e-tests
     ./tests.sh
     ```
   - Or via Make:
     ```bash
     cd e2e-tests
     make all GITHUB_USER=$GITHUB_USER GITHUB_TOKEN=$GITHUB_TOKEN
     ```

Each step will print status and errors if any test fails.

## Individual Steps

- `make login` / `tests.sh` → authenticate to `ghcr.io`
- `make publish` → publish `ci.mk` to OCI
- `make pull` → pull module into cache
- `make remote` → run test against OCI module
- `make http` → run test against HTTP module
- `make local` → run test against local Makefile
- `make clean` → clean generated artifacts

---

Happy testing!  🚀