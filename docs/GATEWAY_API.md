# ipvn7 — Smart Component Gateway API

**Versión:** 0.7.0  
**Puerto Base:** `http://127.0.0.1:7070`  
**Módulos:** [`pkg/core/gateway.go`](../pkg/core/gateway.go), [`pkg/core/server.go`](../pkg/core/server.go)

---

## 1. Arquitectura del Bus
El Smart Component Gateway desacopla el núcleo universal (L0-L2) de servicios satelitales (agentes IA, bots, UIs, puentes IoT) mediante HTTP REST, SSE y WebSockets:
```text
[Componente Satelital (Py/JS/Rust/Go)] ──► REST/SSE/WS ──► [Gateway :7070] ──► [Núcleo ipvn7]
```

---

## 2. Especificación de Endpoints

### 2.1 `GET /api/v1/status`
Telemetría global, estado del nodo y lista de componentes activos.
```bash
curl -s http://127.0.0.1:7070/api/v1/status
```

### 2.2 `POST /api/v1/components/register`
Registra un nuevo componente satelital en el bus.
```bash
curl -X POST http://127.0.0.1:7070/api/v1/components/register \
  -H "Content-Type: application/json" \
  -d '{
    "id": "agente_auditor",
    "name": "Agente Centinela",
    "version": "1.0.0",
    "capabilities": ["telemetry:audit", "chat:notify"],
    "endpoint": "http://127.0.0.1:9090/webhook",
    "transport": "http"
  }'
```
* **Transportes soportados:** `"http"`, `"websocket"`, `"ipc"`, `"inproc"`.

### 2.3 `POST /api/v1/components/unregister`
Desacopla limpiamente un componente.
```bash
curl -X POST http://127.0.0.1:7070/api/v1/components/unregister \
  -H "Content-Type: application/json" \
  -d '{"id": "agente_auditor"}'
```

### 2.4 `POST /api/v1/send`
Transmite un datagrama E2EE encapsulado en la trama fija de 1280B hacia la malla.
```bash
curl -X POST http://127.0.0.1:7070/api/v1/send \
  -H "Content-Type: application/json" \
  -d '{
    "target_did": "did:ipvn7:bda3fed80fbbb6e52ed2bd34f2dc87a63436f301f07e7748d72c5b4a1b5c0c24",
    "protocol": "agent:task_dispatch",
    "priority": 1,
    "payload": [72, 111, 108, 97]
  }'
```
* **Prioridades:** `0` (Bulk), `1` (Interactive / Streaming), `2` (Control crítico).

### 2.5 `GET /api/v1/events` (Server-Sent Events)
Flujo continuo en tiempo real de eventos de red:
```bash
curl -N -H "Accept: text/event-stream" http://127.0.0.1:7070/api/v1/events
```
* **Tipos de Eventos:** `PEER_JOINED`, `PEER_LEFT`, `PACKET_RX`, `PACKET_TX`, `PACKET_DROP`, `ZTNA_ALERT`, `COMPONENT_BOUND`, `COMPONENT_UNBOUND`, `TOPOLOGY_CHANGE`.

---

## 3. Integración en Python (Ejemplo Mínimo)

```python
import requests

GATEWAY = "http://127.0.0.1:7070"

# Enviar datagrama soberano
requests.post(f"{GATEWAY}/api/v1/send", json={
    "target_did": "did:ipvn7:destinatario",
    "protocol": "p2p:ping",
    "priority": 2,
    "payload": list(b"PING")
})
```

---

## 4. Endpoint Nativo Model Context Protocol (`/mcp`) para Agentes de IA

IPvN7 expone un endpoint nativo compatible con la especificación **Model Context Protocol (MCP)** mediante JSON-RPC 2.0 sobre HTTP, permitiendo que agentes de IA (Claude, Cursor, Gemini o agentes locales) inspeccionen y controlen la red soberana:

### 4.1 Herramientas Disponibles (`tools/list`)
* `get_network_status`: Consulta el estado físico de la red, túnel VPN, pares y dispositivos puenteados.
* `toggle_vpn`: Conecta o desconecta el túnel soberano (parámetro: `{"action": "connect" | "disconnect"}`).
* `list_peers`: Lista los pares físicos conectados en la malla soberana con sus DIDs.

### 4.2 Ejemplo de Invocación JSON-RPC
```bash
curl -X POST http://127.0.0.1:7070/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "tools/call",
    "params": {
      "name": "get_network_status",
      "arguments": {}
    }
  }'
```
Respuesta:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [
      {
        "type": "text",
        "text": "VPN State: connected | Peers: 1 | IPv4: 10.7.0.2 | PQC: ML-KEM-768"
      }
    ]
  }
}
```

---

## 5. Protocolo Agent2Agent (A2A) y Manifiesto Canónico Agent Card

IPvN7 implementa la especificación **Agent2Agent (A2A v1.0.0)** de la Agentic AI Foundation (AAIF) para la interoperabilidad horizontal ("Este-Oeste") entre agentes autónomos:

### 5.1 Manifiesto Canónico (`GET /.well-known/agent-card.json`)
Expone la tarjeta de capacidades del agente soberano de red:
```bash
curl -s http://127.0.0.1:7070/.well-known/agent-card.json
```
Respuesta:
```json
{
  "name": "ipvn7-network-os-agent",
  "version": "0.7.0",
  "did": "did:ipvn7:01a4e...",
  "protocols": ["a2a/1.0", "mcp/2024-11-05"],
  "endpoints": {
    "a2a": "/a2a",
    "mcp": "/mcp",
    "card": "/.well-known/agent-card.json"
  },
  "capabilities": ["p2p_routing", "nat_traversal", "pqc_encryption", "ztna_shield"],
  "security": {
    "pqc": "ML-KEM-768",
    "signature": "Ed25519",
    "ztna": "default-deny"
  }
}
```

### 5.2 Handshake y Descubrimiento Este-Oeste (`POST /a2a`)
```bash
curl -X POST http://127.0.0.1:7070/a2a \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "handshake",
    "params": {}
  }'
```
Respuesta:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "status": "accepted",
    "agent": "ipvn7-network-os-agent",
    "did": "did:ipvn7:01a4e...",
    "pqc": "ML-KEM-768"
  }
}
```

---

## 6. Cadenas de Delegación APC y Control Adaptativo (DEC-125)

IPvN7 resuelve la "brecha de gobernanza" de MCP y A2A implementando la **Agentic Principal Chain (APC)** conforme al IETF `draft-das-agentic-adaptive-authorization-00`:

### 6.1 Niveles de Control Graduado
- **Ordinary:** Consultas de estado, telemetría y lectura de pares (aprobadas automáticamente si el DID está autenticado).
- **Escalated:** Conexión/desconexión de VPN, modificación de rutas o puentes LAN (requiere firma Ed25519 de la raíz o token APC con atenuación válida).
- **Quarantine:** Solicitudes de agentes no verificados aisladas en proxy de userspace.

### 6.2 Validación de Cadena en A2A (`POST /a2a`)
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tasks/send",
  "params": {
    "task": "distributed_inference",
    "chain": {
      "root_did": "did:ipvn7:01a4e...",
      "links": [
        {
          "issuer_did": "did:ipvn7:01a4e...",
          "subject_did": "did:ipvn7:b2f81...",
          "scope": "task:delegate",
          "tier": "ordinary",
          "expires_at": 1790000000,
          "signature": "3a0b81..."
        }
      ]
    }
  }
}
```
Respuesta:
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "result": {
    "status": "routed",
    "tier": "ordinary",
    "transport": "p2p_sovereign_datagram"
  }
}
```
Si la cadena está rota, expirada o intenta escalar privilegios (`ordinary -> escalated`), el endpoint responde de inmediato con error `-32003: Delegation denied`.

