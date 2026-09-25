#!/bin/sh
# ==============================================================================
# IPVN7 Universal Installer (Linux / macOS)
# Instala y configura el Sistema Operativo de Redes Soberano IPVN7
# Soporta ejecución como Root (Kernel TUN) o Usuario Estándar (Zero-Admin Userspace)
# ==============================================================================
set -e

REPO="galleguillosdavid-coder/ipvn7_0.7"
SERVICE_NAME="ipvn7"

log_info() { printf "\033[1;32m[IPVN7 INFO]\033[0m %s\n" "$1"; }
log_warn() { printf "\033[1;33m[IPVN7 WARN]\033[0m %s\n" "$1"; }
log_error() { printf "\033[1;31m[IPVN7 ERROR]\033[0m %s\n" "$1" >&2; }

detect_privileges() {
    if [ "$(id -u)" -eq 0 ]; then
        IS_ROOT=1
        INSTALL_DIR="/usr/local/bin"
        CONFIG_DIR="/etc/ipvn7"
        SYSTEMD_DIR="/etc/systemd/system"
        log_info "Modo de Privilegios: ROOT (Kernel TUN / System Service)"
    else
        IS_ROOT=0
        INSTALL_DIR="$HOME/.local/bin"
        CONFIG_DIR="$HOME/.config/ipvn7"
        SYSTEMD_DIR="$HOME/.config/systemd/user"
        log_info "Modo de Privilegios: USUARIO ESTÁNDAR (Zero-Admin Userspace / SOCKS5)"
    fi
}

detect_platform() {
    OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
    ARCH="$(uname -m)"
    case "$ARCH" in
        x86_64|amd64) GOARCH="amd64" ;;
        aarch64|arm64) GOARCH="arm64" ;;
        armv7l|armhf)  GOARCH="arm" ;;
        *) log_error "Arquitectura no soportada: $ARCH"; exit 1 ;;
    esac
    case "$OS" in
        linux)  GOOS="linux" ;;
        darwin) GOOS="darwin" ;;
        *) log_error "Sistema Operativo no soportado: $OS"; exit 1 ;;
    esac
    log_info "Plataforma detectada: ${GOOS}/${GOARCH}"
}

install_binary() {
    mkdir -p "$INSTALL_DIR"
    mkdir -p "$CONFIG_DIR"

    if [ -f "./ipvn7" ]; then
        log_info "Instalando binario local compilado..."
        cp -f "./ipvn7" "$INSTALL_DIR/ipvn7"
        chmod +x "$INSTALL_DIR/ipvn7"
    elif [ -f "./ipvn7.exe" ] && [ "$GOOS" = "linux" ]; then
        log_info "Compilando binario nativo Linux..."
        if command -v go >/dev/null 2>&1; then
            go build -trimpath -ldflags="-s -w" -o "$INSTALL_DIR/ipvn7" ./cmd/ipvn7
            chmod +x "$INSTALL_DIR/ipvn7"
        else
            log_error "Binario Linux no encontrado y 'go' no está disponible."
            exit 1
        fi
    else
        if ! command -v ipvn7 >/dev/null 2>&1; then
            if command -v go >/dev/null 2>&1; then
                log_info "Compilando con Go..."
                go build -trimpath -ldflags="-s -w" -o "$INSTALL_DIR/ipvn7" ./cmd/ipvn7
                chmod +x "$INSTALL_DIR/ipvn7"
            else
                log_error "No se encontró el binario ni el compilador Go."
                exit 1
            fi
        fi
    fi

    if [ "$IS_ROOT" -eq 1 ] && [ "$GOOS" = "linux" ] && command -v setcap >/dev/null 2>&1; then
        setcap 'cap_net_admin,cap_net_bind_service=+ep' "$INSTALL_DIR/ipvn7" || log_warn "setcap falló"
    fi
    log_info "Binario instalado en $INSTALL_DIR/ipvn7"
}

configure_service() {
    if [ "$GOOS" = "linux" ] && command -v systemctl >/dev/null 2>&1; then
        mkdir -p "$SYSTEMD_DIR"
        cat <<EOF > "$SYSTEMD_DIR/$SERVICE_NAME.service"
[Unit]
Description=IPVN7 Sovereign Network OS Daemon
After=network.target network-online.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/ipvn7 -port 7070 -web-port 7070 -udp 7777 -data $CONFIG_DIR -vpn
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=default.target
EOF
        if [ "$IS_ROOT" -eq 1 ]; then
            systemctl daemon-reload
            systemctl enable "$SERVICE_NAME" || true
            log_info "Servicio de sistema habilitado: systemctl start $SERVICE_NAME"
        else
            systemctl --user daemon-reload
            systemctl --user enable "$SERVICE_NAME" || true
            log_info "Servicio de usuario habilitado: systemctl --user start $SERVICE_NAME"
        fi
    fi
}

main() {
    log_info "=================================================="
    log_info "  INSTALADOR UNIVERSAL IPVN7 NETWORK OS (30s)    "
    log_info "=================================================="
    detect_privileges
    detect_platform
    install_binary
    configure_service
    log_info "=================================================="
    log_info "  INSTALACIÓN COMPLETADA EXITOSAMENTE             "
    log_info "  Dashboard: http://localhost:7070                "
    log_info "  Puerto Mesh UDP: 7777                           "
    log_info "=================================================="
}

main "$@"
