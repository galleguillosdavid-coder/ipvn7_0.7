#!/usr/bin/env bash
# ==============================================================================
# build_hardened.sh - Compilacion Blindada y Ofuscada IPVN7 - Rol N
# Aplica ofuscacion de AST, cifrado estatico de literales y stripping DWARF
# ==============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SRC_DIR="$REPO_ROOT/src"
BIN_DIR="$REPO_ROOT/bin"

mkdir -p "$BIN_DIR"

echo "================================================================"
echo "  IPVN7 - COMPILACION BLINDADA Y OFUSCACION (ROL N)"
echo "================================================================"

GARBLE_CMD="garble"
GARBLE_AVAILABLE=false

if command -v garble >/dev/null 2>&1; then
    GARBLE_AVAILABLE=true
elif [ -f "$HOME/go/bin/garble" ]; then
    GARBLE_CMD="$HOME/go/bin/garble"
    GARBLE_AVAILABLE=true
fi

if [ "$GARBLE_AVAILABLE" = true ]; then
    echo "[+] Motor de ofuscacion detectado: GARBLE activo"
    echo "    Banderas: -literals (strings cifrados), -tiny (poda tipos), -seed=random"
else
    echo "[!] ADVERTENCIA: Garble no detectado en PATH/GOPATH."
    echo "    Ejecutando con Stripping Nativo Go (-trimpath -ldflags='-s -w')."
    echo "    Para ofuscacion AST profunda: go install mvdan.cc/garble@latest"
fi

TARGET_OS="${1:-linux}"
TARGET_ARCH="${2:-amd64}"
LDFLAGS="-s -w"

cd "$SRC_DIR"

OUT_NAME="ipvn7-hardened"
if [ "$TARGET_OS" = "windows" ]; then
    OUT_NAME="ipvn7-hardened.exe"
fi

echo "[-] Compilando blindado $TARGET_OS/$TARGET_ARCH -> $OUT_NAME..."

export GOOS="$TARGET_OS"
export GOARCH="$TARGET_ARCH"
export CGO_ENABLED=0

if [ "$GARBLE_AVAILABLE" = true ]; then
    "$GARBLE_CMD" -literals -tiny -seed=random build -trimpath -ldflags="$LDFLAGS" -o "$BIN_DIR/$OUT_NAME" ./cmd/ipvn7
else
    go build -trimpath -ldflags="$LDFLAGS" -o "$BIN_DIR/$OUT_NAME" ./cmd/ipvn7
fi

echo " [OK] Binario generado en bin/$OUT_NAME"

# Suma SHA256
if command -v sha256sum >/dev/null 2>&1; then
    cd "$BIN_DIR"
    sha256sum "$OUT_NAME" >> SHA256SUMS.txt
    echo " [OK] Manifiesto SHA256 actualizado en bin/SHA256SUMS.txt"
fi

echo "================================================================"
echo "  BLINDAJE Y EMPAQUETADO COMPLETADO EXITOSAMENTE (ROL N)"
echo "================================================================"
