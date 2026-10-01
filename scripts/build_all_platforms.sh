#!/usr/bin/env bash
# ==============================================================================
# build_all_platforms.sh - Universal Multiplatform Builder IPVN7
# Genera binarios estáticos independientes (CGO_ENABLED=0) para Windows, Linux y macOS
# ==============================================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC_DIR="$REPO_ROOT/src"
BIN_DIR="$REPO_ROOT/bin"

mkdir -p "$BIN_DIR"

echo "================================================================"
echo "  IPVN7 - COMPILACIÓN MULTIPLATAFORMA UNIVERSAL (ROL K)"
echo "================================================================"

TARGETS=(
    "windows/amd64/ipvn7-windows-amd64.exe"
    "windows/arm64/ipvn7-windows-arm64.exe"
    "linux/amd64/ipvn7-linux-amd64"
    "linux/arm64/ipvn7-linux-arm64"
    "darwin/amd64/ipvn7-darwin-amd64"
    "darwin/arm64/ipvn7-darwin-arm64"
)

LDFLAGS="-s -w -X main.BuildVersion=v0.7.0"

cd "$SRC_DIR"
for item in "${TARGETS[@]}"; do
    IFS='/' read -r target_os target_arch out_name <<< "$item"
    echo -n "[-] Compilando para $target_os/$target_arch -> $out_name..."
    CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" \
        go build -trimpath -ldflags="$LDFLAGS" -o "$BIN_DIR/$out_name" ./cmd/ipvn7
    echo " [OK]"
done

# Copiar binario local linux por conveniencia si aplica
if [[ "$(uname -s)" == "Linux" && -f "$BIN_DIR/ipvn7-linux-amd64" ]]; then
    cp -f "$BIN_DIR/ipvn7-linux-amd64" "$BIN_DIR/ipvn7"
fi

cd "$BIN_DIR"
echo "[-] Calculando sumas SHA256..."
if command -v sha256sum >/dev/null 2>&1; then
    sha256sum ipvn7-* > SHA256SUMS.txt
elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 ipvn7-* > SHA256SUMS.txt
fi

echo "================================================================"
echo "  COMPILACIÓN COMPLETADA EXITOSAMENTE (0 DEPENDENCIAS)"
echo "================================================================"
