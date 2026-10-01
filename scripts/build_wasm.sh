#!/usr/bin/env bash
# ==============================================================================
# build_wasm.sh - Compilador reproducible de WebAssembly para IPVN7 en POSIX
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$REPO_ROOT"

echo "--- Compilando IPVN7 WebAssembly (GOOS=js GOARCH=wasm) ---"
GOOS=js GOARCH=wasm go build -o web/ipvn7.wasm ./cmd/ipvn7-wasm

GOROOT_DIR="$(go env GOROOT)"
if [ -f "$GOROOT_DIR/misc/wasm/wasm_exec.js" ]; then
    cp "$GOROOT_DIR/misc/wasm/wasm_exec.js" web/wasm_exec.js
elif [ -f "$GOROOT_DIR/lib/wasm/wasm_exec.js" ]; then
    cp "$GOROOT_DIR/lib/wasm/wasm_exec.js" web/wasm_exec.js
fi

echo "  [OK] web/ipvn7.wasm y wasm_exec.js listos para distribución universal."
