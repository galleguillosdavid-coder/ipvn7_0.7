# RES-003: Análisis Comparativo de Criptografía Post-Cuántica: WireGuard PSK vs ipvn7 Native X-Wing

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE DE SUPERIORIDAD CRIPTOGRÁFICA
* **Fuente Primaria:** IETF Datatracker, WireGuard Technical Whitepaper, NIST FIPS 203, Thom Wiggers / Rebar research.
* **Áreas de Impacto:** Capa L0/L1, Handshake Cuántico-Resistente, Resistencia sin Servidores Centrales.

---

## 1. El Dilema de la Industria: ¿Por qué WireGuard no tiene PQC Nativo?

A fecha de septiembre de 2026, WireGuard oficial **carece de handshake post-cuántico nativo**:
1. **Problema de Fragmentación UDP:** El handshake Noise de WireGuard está diseñado para caber rígidamente en un único paquete UDP diminuto. Las claves y criptogramas de ML-KEM-768 (1088 bytes) o ML-DSA (3293 bytes) exceden el presupuesto de paquete de WireGuard, forzando fragmentación IP que es descartada por la mayoría de firewalls y routers de Internet.
2. **Dependencia de Claves Pre-compartidas (PSK):** Para mitigar ataques "harvest now, decrypt later", proveedores de VPN comerciales que usan WireGuard han tenido que recurrir a inyectar una PSK clásica/PQC rotada fuera de banda mediante un canal de control TLS 1.3 centralizado.
3. **Pérdida de Soberanía:** Esta solución destruye la descentralización: si el servidor central de control cae o está bloqueado, no hay intercambio seguro de PSK.

---

## 2. Comparativa Técnica Factual: WireGuard vs ipvn7

| Dimensión | WireGuard (Estado del Arte 2026) | ipvn7 Network OS |
|---|---|---|
| **Handshake PQC** | **Ninguno en core.** Depende de canal de control TLS 1.3 externo para rotar PSK. | **Nativo en banda.** X-Wing KEM (X25519 + ML-KEM-768 FIPS 203) en cada datagrama. |
| **Tamaño de Trama y MTU** | Colapsa ante claves >1KB por falta de cuantización fija. | Trama canónica determinista de **1280B** con slots cuantizados para wire Sphinx. |
| **Arquitectura de Claves** | Claves estáticas Noise (`Curve25519`). | Claves soberanas desacopladas (`did:ipvn7:<pubkey>`) con rotación proactiva cada 1 GB/1h. |
| **Dependencia Central** | Requiere orquestador central para distribuir PSKs cuánticas. | **Cero servidores centrales.** Peering P2P soberano directo con validación local. |
| **Sobrecarga de Memoria** | Driver de kernel o userspace con alocaciones variables. | Invariante Zero-Copy: **0 B/op y 0 allocs/op** (`35.55 ns/op`). |

---

## 3. Decisión Estratégica para ipvn7 (DEC-095)

* **Conclusión Técnica:** La limitación de WireGuard valida la visión de ipvn7 de diseñar la trama fija de 1280B desde el día cero. Mientras WireGuard requiere muletas centralizadas para ser cuántico-resistente, ipvn7 ejecuta X-Wing KEM e identidad descentralizada de forma nativa e invisible.
* **Plan de Acción:**
  1. Mantener el soporte dual: modo compacto (128B) para enrutamiento cebolla Sphinx multirruta y modo full X-Wing (1120B) para túneles punto a punto directos.
  2. Publicar este análisis formal en la documentación de comparativa global (`docs/08_COMPARATIVA_INFRAESTRUCTURA_GLOBAL.md`).
