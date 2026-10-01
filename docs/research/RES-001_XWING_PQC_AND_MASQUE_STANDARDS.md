# RES-001: Vigilancia Tecnológica, Estandarización IETF PQC (X-Wing) y Transporte MASQUE

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE PARA ROADMAP FASE 38
* **Áreas de Impacto:** Capa L1 (Criptografía Híbrida PQC), Capa L2 (Transporte Zero-Admin WAN), Resiliencia P2P.

---

## 1. Benchmarking de Criptografía Post-Cuántica: WireGuard vs ipvn7

### Estado del Arte Global (Septiembre 2026):
* **WireGuard Oficial:** No cuenta con soporte nativo de PQC en su handshake Noise. Para mitigar ataques "harvest now, decrypt later", la industria ha recurrido al uso de TLS 1.3 externo para inyectar una clave pre-compartida (PSK) en WireGuard.
* **Estándar IETF Emergente:**
  * **RFC 10024 (Publicado Agosto 2026):** Acuerdo de claves híbrido en TLS 1.3 (`X25519MLKEM768`).
  * **`draft-connolly-cfrg-xwing-kem` (X-Wing KEM):** Construcción formal del Crypto Forum Research Group (CFRG) del IETF que combina formalmente `X25519` con `ML-KEM-768` (NIST FIPS 203).

### Adopción en ipvn7:
* El módulo `pkg/l1/hybrid_kem.go` ya implementa FIPS 203 ML-KEM y FIPS 204 ML-DSA con encapsulamiento híbrido.
* **Plan de Acción:** Alinear la función KDF y serialización con el borrador canónico de **X-Wing KEM**, asegurando interoperabilidad estándar y compatibilidad nativa sin depender de suites propietarias.

---

## 2. Perforación NAT y Retransmisión: Tailscale (DERP) vs Malla Soberana ipvn7

| Dimensión | Tailscale (STUN + DERP + DISCO) | ipvn7 (STUN + Kleinberg + Ed25519) |
|---|---|---|
| **P2P Directo** | Intenta STUN + ICE hole-punching | Perforación UDP bidireccional STUN RFC 5389 |
| **Relay de Respaldo** | Servidores centrales operados por Tailscale (DERP) | Relay efímero distribuido sobre los 12 anillos de Kleinberg |
| **Dependencia Central** | **ALTA:** Si los servidores de Tailscale caen, no hay red | **CERO:** Todo nodo de la red puede actuar como retransmisor firmado |
| **Privilegios de SO** | Requiere demonio root/admin para TUN | TUN automático con admin; SOCKS5 userspace sin admin |

---

## 3. Transporte Zero-Admin: Evasión mediante MASQUE (RFC 9298)

* **Problema:** En redes corporativas restrictivas o campus universitarios, el tráfico UDP crudo en puertos no estándar suele ser bloqueado por firewalls de inspección profunda (DPI).
* **Solución Estándar:** **RFC 9298 (CONNECT-UDP / MASQUE)**. Permite transportar datagramas UDP sobre túneles seguros HTTP/3 (QUIC) o HTTP/2 (TLS 1.3) en el puerto 443 estándar.
* **Plan de Acción para ipvn7:**
  1. Extender `pkg/l2/tunnel_tls.go` con soporte para el encabezado `CONNECT-UDP` de RFC 9298.
  2. Proveer fallback automático e invisible: si UDP directo en el puerto 7777 falla por CGNAT simétrico extremo, canalizar la trama 1280B vía MASQUE sobre el puerto 443 sin requerir privilegios de administrador.
