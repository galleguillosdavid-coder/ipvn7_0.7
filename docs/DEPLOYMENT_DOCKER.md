# Despliegue Perimetral en Contenedores: IPVN7 Network OS

Guía de despliegue soberano en contenedores mínimos basados en `scratch` sin privilegios de root (`UID 10007`).

---

## 1. Características del Contenedor Canónico

* **Superficie de ataque nula:** Imagen pura `scratch` sin shell (`/bin/sh`), utilitarios ni librerías dinámicas del sistema operativo anfitrión.
* **Cero privilegios:** Ejecución bajo UID `10007:10007`, operando con `ModeUserspaceProxy` (SOCKS5 `:10807` y HTTP CONNECT `:10808`) sin requerir permisos de kernel `CAP_NET_ADMIN`.
* **Binario estático:** Compilado con `CGO_ENABLED=0` y flags `-trimpath -ldflags="-s -w -extldflags '-static'"`.

---

## 2. Construcción de la Imagen

```bash
docker build -t ipvn7-core:latest .
```

---

## 3. Ejecución con Persistencia de Identidad Criptográfica

Para preservar la identidad criptográfica Ed25519 (`did:ipvn7:<pubkey>`) entre reinicios del contenedor:

```bash
docker run -d \
  --name ipvn7-node \
  --restart unless-stopped \
  -p 7777:7777/udp \
  -p 7070:7070 \
  -p 10807:10807 \
  -p 10808:10808 \
  -v $(pwd)/keystore:/keystore \
  ipvn7-core:latest
```

---

## 4. Mapeo de Puertos y Protocolos

| Puerto | Protocolo | Capa | Propósito |
|:---:|:---:|:---:|---|
| **7777** | UDP | L1 | Transporte de malla P2P (Kleinberg 12 anillos concéntricos) |
| **7070** | TCP | Core | API REST, SSE y Dashboard de control Web |
| **10807** | TCP | L1 | Proxy SOCKS5 para navegación e infraestructura |
| **10808** | TCP | L1 | Proxy HTTP CONNECT para túneles y camuflaje TLS 1.3 |

---

## 5. Verificación de Salud en Vivo

```bash
curl -s http://localhost:7070/api/v1/status | jq .
```
Retorna el estado de identidad soberana, prefijos IPv6/IPv4 y métricas de telemetría sin requerir herramientas instaladas dentro del contenedor.
