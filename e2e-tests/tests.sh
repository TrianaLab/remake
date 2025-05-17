#!/usr/bin/env bash
set -euo pipefail

ROOT=$(cd "$(dirname "$0")" && pwd)
export GITHUB_USER=edu-diaz
export GITHUB_TOKEN=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

cd ..
make install
cd e2e-tests

echo "==> 1. Login"
remake login -u "$GITHUB_USER" -p "$GITHUB_TOKEN"

echo "==> 2. Publish ci.mk to GHCR"
remake publish ${GITHUB_USER}/ci.mk:v0.1.0 -f "$ROOT/fixtures/ci.mk"

echo "==> 3. Pull ci.mk (shorthand, latest)"
remake pull ${GITHUB_USER}/ci.mk:v0.1.0 -o "$ROOT/pulled-ci.mk"

echo "==> 4. Run remote module"
cat > "$ROOT/Makefile.remote" <<EOF
include oci://ghcr.io/${GITHUB_USER}/ci.mk:v0.1.0

.PHONY: test
test: ci
EOF
remake run test -f "$ROOT/Makefile.remote"

echo "==> 5. Run HTTP module"
# asume servidor en http://localhost:8000/http.mk
cat > "$ROOT/Makefile.http" <<EOF
include http://localhost:8000/http.mk

.PHONY: test
test: http
EOF
remake run test -f "$ROOT/Makefile.http"

echo "==> 6. Run local module"
cat > "$ROOT/Makefile.local" <<EOF
include fixtures/local.mk

.PHONY: test
test: hello
EOF
remake run test -f "$ROOT/Makefile.local"

echo "==> 7. Cleanup"
rm -f "$ROOT"/pulled-*.mk "$ROOT"/Makefile.* .remake
