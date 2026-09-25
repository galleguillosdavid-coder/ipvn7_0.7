#!/usr/bin/env bash
# ==============================================================================
# start_vpn_i7.sh - Lanzador Soberano 1-Clic "VPN I7" con Autodetección Zero-Admin
# ==============================================================================

set -e

PORT=${1:-7777}
WEB_PORT=${2:-7070}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo ""
echo "================================================================"
echo "         IPVN7 - INICIANDO NODO SOBERANO ZERO-FRICTION"
echo "================================================================"

# 1. Asegurar existencia de binario
BIN_PATH="$SCRIPT_DIR/ipvn7"
if [ ! -f "$BIN_PATH" ]; then
    echo "[BUILD] Binario no detectado. Compilando ipvn7..."
    (cd "$SCRIPT_DIR" && go build -o ipvn7 ./cmd/ipvn7)
    echo "[BUILD] Compilacion exitosa."
fi

# 2. Detección de Privilegios Root
CMD_ARGS=("-port" "$PORT" "-web-port" "$WEB_PORT")

if [ "$(id -u)" -eq 0 ]; then
    echo "[MODO] Root detectado: Activando interfaz TUN Nativa de Kernel..."
    CMD_ARGS+=("-tun-native")
else
    echo "[MODO] Usuario estandar detectado: Activando Userspace FastPath (Zero-Privilegios)..."
    echo "       (SOCKS5 Proxy operativo en localhost:10807 sin requerir permisos root)"
fi

# 3. Lanzamiento del Navegador
URL="http://localhost:$WEB_PORT"
echo "[UI] Abriendo panel de control en: $URL"
if command -v xdg-open > /dev/null 2>&1; then
    xdg-open "$URL" > /dev/null 2>&1 &
elif command -v open > /dev/null 2>&1; then
    open "$URL" > /dev/null 2>&1 &
fi

echo "[READY] Ejecutando Núcleo Universal ipvn7..."
echo "================================================================"

# 4. Ejecución del binario
exec "$BIN_PATH" "${CMD_ARGS[@]}"
