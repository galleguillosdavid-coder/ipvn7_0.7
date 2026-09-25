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
