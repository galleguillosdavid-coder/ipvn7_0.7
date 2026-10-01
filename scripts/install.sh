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
    elif [ -f "./bin/ipvn7-${GOOS}-${GOARCH}" ]; then
        log_info "Instalando binario de arquitectura detectada..."
        cp -f "./bin/ipvn7-${GOOS}-${GOARCH}" "$INSTALL_DIR/ipvn7"
        chmod +x "$INSTALL_DIR/ipvn7"
    elif command -v go >/dev/null 2>&1 && [ -d "./src" ]; then
        log_info "Compilando con Go local..."
        (cd ./src && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$INSTALL_DIR/ipvn7" ./cmd/ipvn7)
        chmod +x "$INSTALL_DIR/ipvn7"
    else
        log_info "Descargando binario estatico ipvn7-${GOOS}-${GOARCH}..."
        DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/ipvn7-${GOOS}-${GOARCH}"
        if command -v curl >/dev/null 2>&1; then
            curl -fsSL -o "$INSTALL_DIR/ipvn7" "$DOWNLOAD_URL" || true
        elif command -v wget >/dev/null 2>&1; then
            wget -qO "$INSTALL_DIR/ipvn7" "$DOWNLOAD_URL" || true
        fi
        
        if [ ! -s "$INSTALL_DIR/ipvn7" ] && command -v go >/dev/null 2>&1; then
            log_info "Compilando con Go global..."
            go install "github.com/${REPO}/src/cmd/ipvn7@latest" || true
            if [ -f "$HOME/go/bin/ipvn7" ]; then
                cp -f "$HOME/go/bin/ipvn7" "$INSTALL_DIR/ipvn7"
            fi
        fi

        if [ ! -s "$INSTALL_DIR/ipvn7" ]; then
            log_error "No se pudo obtener el binario precompilado ni compilar con Go."
            log_warn "Puedes ejecutarlo via Docker: docker run -d --net=host ghcr.io/${REPO}:latest"
            exit 1
        fi
        chmod +x "$INSTALL_DIR/ipvn7"
    fi

    if [ "$IS_ROOT" -eq 1 ] && [ "$GOOS" = "linux" ] && command -v setcap >/dev/null 2>&1; then
        setcap 'cap_net_admin,cap_net_bind_service=+ep' "$INSTALL_DIR/ipvn7" || log_warn "setcap falló"
    fi
    log_info "Binario instalado en $INSTALL_DIR/ipvn7"
}

configure_service() {
    if [ "$GOOS" = "linux" ] && command -v systemctl >/dev/null 2>&1 && [ -d /run/systemd/system ]; then
        mkdir -p "$SYSTEMD_DIR"
        SERVICE_ARGS="-port 7777 -web-port 7070 -keystore $CONFIG_DIR/node_identity.key"
        if [ "$IS_ROOT" -eq 1 ]; then
            SERVICE_ARGS="$SERVICE_ARGS -tun"
        else
            SERVICE_ARGS="$SERVICE_ARGS -socks5 10807"
        fi

        cat <<EOF > "$SYSTEMD_DIR/$SERVICE_NAME.service"
[Unit]
Description=IPVN7 Sovereign Network OS Daemon
After=network.target network-online.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/ipvn7 $SERVICE_ARGS
Restart=always
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=default.target
EOF
        if [ "$IS_ROOT" -eq 1 ]; then
            systemctl daemon-reload
            systemctl enable "$SERVICE_NAME" || true
            systemctl start "$SERVICE_NAME" || true
            log_info "Servicio de sistema iniciado: $SERVICE_NAME"
        else
            systemctl --user daemon-reload
            systemctl --user enable "$SERVICE_NAME" || true
            systemctl --user start "$SERVICE_NAME" || true
            log_info "Servicio de usuario iniciado: $SERVICE_NAME"
        fi
    else
        log_info "Iniciando daemon ipvn7 en segundo plano..."
        nohup "$INSTALL_DIR/ipvn7" -port 7777 -web-port 7070 -socks5 10807 >/dev/null 2>&1 &
    fi
}

create_desktop_entry() {
    DESKTOP_DIR="/usr/share/applications"
    if [ "$IS_ROOT" -ne 1 ]; then
        DESKTOP_DIR="$HOME/.local/share/applications"
    fi
    if [ -d "$DESKTOP_DIR" ] || [ -n "$DISPLAY" ] || [ -n "$WAYLAND_DISPLAY" ]; then
        mkdir -p "$DESKTOP_DIR"
        cat <<EOF > "$DESKTOP_DIR/ipvn7.desktop"
[Desktop Entry]
Name=VPN I7
Comment=Red Soberana Cuántica
Exec=xdg-open http://localhost:7070
Terminal=false
Type=Application
Categories=Network;Security;
Icon=network-vpn
EOF
        log_info "Acceso directo de escritorio creado: $DESKTOP_DIR/ipvn7.desktop"
    fi
}

open_ui_if_graphical() {
    URL="http://localhost:7070"
    if [ -n "$DISPLAY" ] || [ -n "$WAYLAND_DISPLAY" ] || [ -n "$WSL_DISTRO_NAME" ]; then
        if command -v msedge.exe >/dev/null 2>&1; then
            (sleep 1 && msedge.exe --app="$URL" --window-size=440,680 >/dev/null 2>&1) &
        elif command -v cmd.exe >/dev/null 2>&1; then
            (sleep 1 && cmd.exe /c start "$URL" >/dev/null 2>&1) &
        elif command -v xdg-open >/dev/null 2>&1; then
            (sleep 1 && xdg-open "$URL" >/dev/null 2>&1) &
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
    create_desktop_entry
    open_ui_if_graphical
    log_info "=================================================="
    log_info "  INSTALACIÓN COMPLETADA EXITOSAMENTE             "
    log_info "  Dashboard: http://localhost:7070                "
    log_info "  Puerto Mesh UDP: 7777                           "
    log_info "=================================================="
}

main "$@"
