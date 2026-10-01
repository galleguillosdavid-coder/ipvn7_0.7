# Especificaciones Canónicas de Componentes Satelitales (v0.7.0)

Este documento condensa en especificaciones de lenguaje natural de alta densidad los subsistemas satelitales avanzados de ipvn7. Sustituye las implementaciones históricas archivadas, preservando sus contratos, protocolos y directivas para cuando se requiera su reimplementación sobre el Núcleo Funcional Universal (Pure Go, Zero-Copy, $\le 400$ líneas).

---

## 1. L3: Gobernanza y Senado de Agentes (Democracia Líquida)
* **Propósito:** Toma de decisiones descentralizada y votación ponderada sobre cambios de enrutamiento y parámetros de red entre agentes y nodos soberanos.
* **Modelo y Contratos:**
  - **Identidad:** Nodos votantes identificados por su DID Ed25519 (`did:ipvn7:<pubkey>`).
  - **Mecanismo:** Delegación de voto transitiva (Democracia Líquida) basada en reputación de enrutamiento (Tit-for-Tat / Web-of-Trust).
  - **Propuestas (`Proposal`):** Objeto CBOR con `ID`, `AutorDID`, `Categoria` (routing, security, params), `Payload`, `Quorum` y `Deadline`.
  - **Validación Constitucional:** Todo voto y propuesta debe satisfacer los axiomas inmutables (Axioma III $\le 400$ líneas, Zero-Copy 0 B/op, Privacidad Default-Deny).
* **Integración Satelital:** Se conecta externamente a través de `/api/v1/events` y `/api/v1/send` del Smart Component Gateway.

---

## 2. L3: Orquestación de Intenciones y Servidor MCP
* **Propósito:** Permitir que usuarios humanos o modelos de IA controlen la red mediante lenguaje natural local con 0 tokens de API comercial.
* **Estándar:** Model Context Protocol (MCP) sobre `stdio` o HTTP/SSE (JSON-RPC 2.0).
* **Herramientas Expuestas (`Tools`):**
  - `network_status()`: Estado de interfaz Wintun L3, gateway proxy y telemetría.
  - `list_peers()`: Tabla de 12 anillos de Kleinberg con latencias RTT.
  - `route_traffic(dest, mode)`: Conmutación entre túnel completo o túnel dividido.
  - `firewall_rule(did, action)`: Modificación de listas ZTNA de control de acceso.
* **Regla Antihumo:** Procesamiento local determinista por coincidencia de patrones y FSM; cero llamadas a APIs externas privativas.

---

## 3. L4: DFS Distribuido (Almacén Inmutable DAG)
* **Propósito:** Almacenamiento distribuido resistente a censura basado en contenido (Content-Addressed Storage - CAS).
* **Estructura de Datos:**
  - **Bloque:** MTU canónico de 1280 bytes, identificado por su hash BLAKE3 o SHA-256 (`cid:blake3:<hash>`).
  - **Árbol Merkle-DAG:** Archivos mayores fragmentados en hojas de 1200 bytes con nodo raíz firmado digitalmente por el DID del creador.
  - **Estrategia DTN (Store-and-Forward):** Nodos móviles transportan fragmentos y los replican vorazmente al cruzarse con pares conocidos en el anillo Kleinberg.

---

## 4. Dispositivos IoT, Domótica y Puentes Satelitales
* **Propósito:** Integrar impresoras, televisores inteligentes, vehículos y sensores en la malla sin modificar su firmware original.
* **Puentes Específicos:**
  - **Shadow DIDs:** Dispositivos LAN detectados pasivamente por mDNS/SSDP/ARP reciben un DID virtual (`did:ipvn7:shadow:<mac>`) en la subred `10.7.0.0/16`.
  - **Puente MQTT / CoAP:** Proxy que traduce tópicos MQTT v5 y recursos CoAP (RFC 7252) a datagramas canónicos CBOR autenticados por Noise XX.
  - **Wake-on-LAN (WoL) P2P:** Envío autenticado del paquete mágico (`0xFF*6 + MAC*16`) a través de un nodo satélite local para encendido remoto seguro.

---

## 5. Mensajería E2EE y Chat P2P Soberano
* **Propósito:** Comunicación humana cifrada de extremo a extremo sin intermediarios, servidores de retransmisión central ni metadatos expuestos.
* **Criptografía:** Double Ratchet post-cuántico (X-Wing KEM + ML-KEM-768 + Ed25519) con Perfect Forward Secrecy (PFS).
* **Enrutamiento:** Envío directo P2P por sockets UDP; si hay CGNAT simétrico extremo, relay ciego *Consume-and-Burn* sin persistencia en disco.
