#!/usr/bin/env bash
# ==============================================================================
# start_vpn_i7.sh - Lanzador Soberano 1-Clic "VPN I7" con Autodetección Zero-Admin
# ==============================================================================

set -e

PORT=${1:-7777}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo ""
echo "================================================================"
echo "         IPVN7 - INICIANDO NODO SOBERANO ZERO-FRICTION"
echo "================================================================"

# 1. Asegurar existencia de binario
BIN_PATH="$REPO_ROOT/bin/ipvn7"
if [ ! -f "$BIN_PATH" ]; then
    echo "[BUILD] Binario no detectado. Compilando ipvn7..."
    (cd "$REPO_ROOT/src" && go build -o "$BIN_PATH" ./cmd/ipvn7)
    echo "[BUILD] Compilacion exitosa."
fi

# 2. Detección de Privilegios Root
CMD_ARGS=("-port" "$PORT")

if [ "$(id -u)" -eq 0 ]; then
    echo "[MODO] Root detectado: Activando interfaz TUN Nativa de Kernel..."
    CMD_ARGS+=("-tun")
else
    echo "[MODO] Usuario estandar detectado: Activando Userspace FastPath (Zero-Privilegios)..."
    echo "       (SOCKS5 Proxy operativo en 127.0.0.1:10807 sin requerir permisos root)"
    CMD_ARGS+=("-socks5" "10807")
fi

echo "[READY] Ejecutando Núcleo Universal ipvn7..."
echo "================================================================"

# 3. Ejecución del binario
exec "$BIN_PATH" "${CMD_ARGS[@]}"
