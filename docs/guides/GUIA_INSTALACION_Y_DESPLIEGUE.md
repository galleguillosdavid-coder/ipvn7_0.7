# GUÍA UNIVERSAL DE INSTALACIÓN Y DESPLIEGUE — IPVN7 v0.7.0

> **Estado:** VIGENTE / PRODUCCIÓN  
> **Plataformas Soportadas:** Windows (x64/ARM64), Linux (amd64/arm64), macOS (Apple Silicon / Intel), Docker  
> **Requisitos:** Go 1.22+ (para compilación local) o binario precompilado oficial  

---

## 1. COMPILACIÓN LOCAL RÁPIDA (0 DEPENDENCIAS EXTERNAS)

Desde la raíz del repositorio o dentro de `src/`:

```bash
# Compilar el binario universal soberano (Windows)
go build -o ./bin/ipvn7.exe ./cmd/ipvn7

# Compilar para Linux (CGO_ENABLED=0 estricto)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/ipvn7 ./cmd/ipvn7

# Compilar para macOS (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/ipvn7-darwin-arm64 ./cmd/ipvn7

# Compilar la herramienta de consola CLI
go build -o ./bin/ipvn7-cli.exe ./cmd/ipvn7-cli
```

---

## 2. EJECUCIÓN EN WINDOWS

### Modo Automático (Doble Clic o Terminal)
```powershell
.\bin\ipvn7.exe --port 7777 --web-port 7070
```

1. **Si se otorgan privilegios de Administrador (UAC):** El nodo inicializa el adaptador de kernel **Wintun L3**, enrutando tráfico TCP/UDP nativo a velocidad de línea.
2. **Si se rechaza el UAC:** El nodo **no falla ni se interrumpe**; conmuta automáticamente a **Modo Usuario (Proxy HTTP CONNECT + SOCKS5 en 127.0.0.1:10807)**, permitiendo navegación segura sin privilegios.
3. **Panel WebUI Local:** Disponible de forma segura en: `http://127.0.0.1:7070`.

---

## 3. DESPLIEGUE EN LINUX (TUN & SYSTEMD)

1. **Crear interfaz TUN y otorgar capacidades sin root:**
```bash
sudo setcap cap_net_admin=eip ./bin/ipvn7
./bin/ipvn7 --port 7777 --web-port 7070
```

2. **Servicio Systemd de fondo (`/etc/systemd/system/ipvn7.service`):**
```ini
[Unit]
Description=IPVN7 Universal Sovereign Core Node
After=network.target

[Service]
Type=simple
User=root
ExecStart=/opt/ipvn7/bin/ipvn7 --port 7777 --web-port 7070
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

---

## 4. DESPLIEGUE EN DOCKER (MULTI-STAGE ULTRA-LIVIANO)

El proyecto incluye un contenedor minimalista multi-stage sin dependencias innecesarias:

```dockerfile
# Construir imagen ligera
docker build -t ipvn7-node:v0.7.0 -f docker/Dockerfile .

# Ejecutar con paso de tráfico de red
docker run -d \
  --name ipvn7 \
  --cap-add=NET_ADMIN \
  --device /dev/net/tun \
  -p 7777:7777/udp \
  -p 7070:7070 \
  ipvn7-node:v0.7.0
```

---

## 5. DIAGNÓSTICO EN FRÍO
Para verificar que el nodo funcione correctamente antes de conectarlo a la malla:
```bash
./bin/ipvn7.exe --diagnostics
```
Valida la generación de identidad Ed25519 (`did:ipvn7:...`), derivación IPv6 soberana (`fd07::/64`), asignación de puertos y disponibilidad de adaptadores TUN/Wintun.
