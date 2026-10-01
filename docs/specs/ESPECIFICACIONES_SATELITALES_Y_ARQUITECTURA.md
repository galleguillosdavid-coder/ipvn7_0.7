# ESPECIFICACIONES DE COMPONENTES SATELITALES Y ARQUITECTURA MULTIPLATAFORMA — IPVN7 v0.7.0

> **Estado:** VIGENTE / ESPECIFICACIÓN TÉCNICA  
> **Norma Rectora:** [`docs/FUENTE_DE_VERDAD.md`](../FUENTE_DE_VERDAD.md) (Capítulo 4)  
> **Decisiones Vinculadas:** DEC-112 (Especificaciones Satelitales), DEC-132 (Consolidación Mundial 6-Arch)  

---

## 1. PRINCIPIO DE DESACOPLAMIENTO SATELITAL
En IPVN7 v0.7, el demonio central (`ipvn7`) es un **Núcleo Funcional Universal** ultra-liviano enfocado exclusivamente en transporte, identidad y enrutamiento L0-L1.
Cualquier funcionalidad de nivel superior (interfaces de usuario, inteligencia artificial, agentes MCP, almacenamiento o herramientas de red) se ejecuta como un **Componente Inteligente Satelital** acoplado externamente a través del **Smart Component Gateway** (`/api/v1/*`, WebSockets, IPC y SSE).

---

## 2. COMPONENTES SATELITALES CANÓNICOS

| Componente | Tipo | Interfaz de Bus | Responsabilidad |
|---|---|---|---|
| **WebUI Zen** | Dashboard / Admin | HTTP REST / SSE | Panel visual de control, métricas en tiempo real y botón VPN 1-clic. |
| **SOCKS5 Proxy** | Adaptador Local | TCP Socket `127.0.0.1:10807` | Pasarela para navegación web sin privilegios de kernel. |
| **Agentic Bridge** | Integración AI / MCP | IPC / WebSockets | Pasarela para conectar agentes autónomos y modelos locales. |
| **P2P File Transfer** | Servicio Satelital | Stream Multiplexado L1 | Transferencia de archivos fragmentada en tramas CBOR firmadas. |
| **Mesh Chat** | Comunicación P2P | Datagramas CBOR deterministas | Mensajería efímera de extremo a extremo sin servidores centrales. |

---

## 3. MATRIZ DE CONSOLIDACIÓN MULTIPLATAFORMA (6-ARCH)

El binario y la suite se certifican sobre 6 arquitecturas de hardware y sistemas operativos de primer orden sin dependencias dinámicas CGO:

1. `windows/amd64` (Windows x64 con driver Wintun nativo).
2. `windows/arm64` (Dispositivos Surface / Copilot+ PCs con emulación o Wintun nativo).
3. `linux/amd64` (Servidores en nube y nodos de enrutamiento WAN).
4. `linux/arm64` (SBCs Raspberry Pi 4/5 y servidores ARM Graviton).
5. `darwin/arm64` (Apple Silicon M1/M2/M3/M4).
6. `darwin/amd64` (macOS Intel legacy).

---

## 4. INVARIANTE DE LATENCIA Y PASE DETERMINISTA
* **MTU:** Siempre <= 1280 bytes en toda trama emitida por componentes satelitales.
* **Separación de memoria:** Los componentes satelitales no pueden alocar directamente sobre los ring-buffers del Core; la comunicación se realiza vía búferes preasignados en el bus de gateway.
