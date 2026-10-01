# Despliegue e Instalación Universal de IPvN7 Network OS

**Versión:** 0.7.0  
**Objetivo:** Instalación en cualquier dispositivo del planeta (Windows, Linux, macOS, Android, Docker) y fuera de él (Satélites LEO, Cubesats, Redes Interplanetarias Desconectadas) en **1 comando o 1 clic**, sin dependencias externas (`CGO_ENABLED=0`).

---

## 🌍 1. Instalación Rápida en la Tierra (1 Comando)

### A. En Cualquier PC con Windows (Windows 10, 11 o Server)
Abre PowerShell y ejecuta:
```powershell
irm https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/main/scripts/install.ps1 | iex
```
* **Qué hace:** Crea el directorio en `%LOCALAPPDATA%\IPVN7`, instala `ipvn7.exe` y `wintun.dll`, abre los puertos en Windows Defender Firewall y crea el acceso directo con el panel visual de 3 estados en el Escritorio.
* **Sin permisos de Administrador:** Opera automáticamente en modo usuario con el Gateway SOCKS5/HTTP Connect en `127.0.0.1:10807`.

---

### B. En Cualquier Servidor, PC o Laptop con Linux (Ubuntu, Debian, Fedora, Arch)
Abre una terminal y ejecuta:
```bash
curl -fsSL https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/main/scripts/install.sh | bash
```
* **Qué hace:** Detecta si eres `root` o usuario estándar. Si es root, activa la interfaz TUN de kernel y registra el servicio `systemd` para arranque automático. Si es usuario común, configura el servicio de usuario `systemctl --user` y el proxy SOCKS5 en `127.0.0.1:10807`.
* **Panel de Control:** Abre `http://localhost:7070` en tu navegador.

---

### C. En Cualquier Mac con macOS (Apple Silicon M1/M2/M3/M4 o Intel)
Abre la Terminal y ejecuta:
```bash
curl -fsSL https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/main/scripts/install.sh | bash
```
* **Qué hace:** Instala el binario estático compilado en `~/.local/bin/ipvn7`, genera las llaves Ed25519 y levanta el panel soberano en `http://localhost:7070`.

---

### D. En Raspberry Pi, Orange Pi o Servidores ARM (Edge IoT)
```bash
curl -fsSL https://raw.githubusercontent.com/galleguillosdavid-coder/ipvn7_0.7/main/scripts/install.sh | bash
```
* Compatible nativo con arquitecturas `arm64` (aarch64) y `armv7` (32-bit).

---

### E. En Cualquier Nube, VPS o Clúster Docker (AWS, GCP, DigitalOcean, Hetzner)
Ejecuta el contenedor reproducible de tamaño mínimo (<10 MB):
```bash
docker run -d \
  --name ipvn7 \
  --restart always \
  --network host \
  -v ipvn7_keystore:/keystore \
  -v ipvn7_data:/data \
  ghcr.io/galleguillosdavid-coder/ipvn7:latest
```

---

## 🛰️ 2. Despliegue Fuera del Planeta (Satélites LEO, Cubesats y Nodos Interplanetarios)

IPvN7 está diseñado desde su capa criptográfica L0 para operar en **entornos de aislamiento orbital y redes con alta latencia y partición física (Delay-Tolerant Networking / DTN)**:

### Principios Físicos de Operación en el Espacio:
1. **Cero Dependencia de DNS Terrestre o Autoridades Centrales:**
   - La identidad del satélite no depende de una IP fija ni de dominios `.com` / ICANN.
   - Cada nodo deriva su identidad criptográfica autónomamente mediante llaves Ed25519: `did:ipvn7:<pubkey>`.
2. **Cifrado Post-Cuántico Blindado contra Radiación y Criptoanálisis:**
   - Cada datagrama L0 viaja encapsulado con **ML-KEM-768 (FIPS 203)** y tramas fijas deterministas CBOR RFC 8949 (1280 bytes).
3. **Descubrimiento Autónomo por Radio / Óptico (Sin Internet Terrestre):**
   - El nodo emite balizas efímeras *Consume-and-Burn* (EBRA) sobre el espectro de radiofrecuencia (UHF/VHF/S-Band) o enlaces ópticos inter-satelitales (ISL).
   - Cuando dos satélites entran en rango de visibilidad orbital, establecen el túnel P2P en microsegundos sin requerir servidores intermediarios.

### Comando de Arranque para Nodos Orbitales / Aislados:
En el sistema embebido del satélite (Linux RTOS o micro-controlador con soporte Go/WASM):
```bash
./ipvn7 --port 7777 --keystore /satellite/secure_key.dat --web-port 7070
```
Si se conoce la órbita de un satélite hermano o estación terrena:
```bash
./ipvn7 --port 7777 --peer 10.7.0.2:7777 --keystore /satellite/secure_key.dat
```

---

## 📋 Resumen de Puertos Canónicos de IPvN7

| Puerto | Protocolo | Capa | Propósito |
| :---: | :---: | :---: | :--- |
| **7777** | UDP | L1 | Transporte físico de malla P2P (Enrutamiento Kleinberg) |
| **7070** | TCP | Core | Panel visual interactivo y API REST / MCP / A2A |
| **10807** | TCP | Core | Gateway Universal SOCKS5 + HTTP CONNECT (Zero-Admin) |
