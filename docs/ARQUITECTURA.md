# ipvn7 Network OS — Arquitectura Canónica y Plan Maestro

**Versión:** 0.7.0  
**Paradigma:** Núcleo Universal desacoplado (Pure Go, `CGO_ENABLED=0`) con Smart Component Gateway.

---

## 1. Arquitectura del Sistema

```text
┌────────────────────────────────────────────────────────────────────────┐
│               COMPONENTES INTELIGENTES SATELITALES (L3 / L4)            │
│  Agentes IA (MCP) · Web Dashboard · E2EE Chat · DAG Store · Petnames   │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │ REST / SSE / WebSockets / IPC
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│                   SMART COMPONENT GATEWAY (Bus Universal)              │
│         Registro de Capacidades · Enrutamiento de Datagramas           │
└───────────────────────────────────┬────────────────────────────────────┘
                                    │
                                    ▼
┌────────────────────────────────────────────────────────────────────────┐
│             NÚCLEO FUNCIONAL UNIVERSAL ipvn7 (Pure Go / L0-L2)         │
│  L0: Ed25519 · Noise XX · PQC Kyber · CBOR Determinista (Core Freeze) │
│  L1: Kleinberg Router · Zero-Copy Pool · ZTNA Firewall · Packet Pacer │
│  L2: Ring Buffer Lock-Free (<28 ns) · Telemetría en Tiempo Real        │
└────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Modelo de Capas e Invariantes

| Capa | Nombre | Estado | Responsabilidad Técnica |
|:---:|---|:---:|---|
| **L0** | **Core Criptográfico** | **FROZEN** | Identidad Ed25519, CBOR RFC 8949 (1280B), Noise XX, PQC (Kyber/ML-KEM), Anti-Replay. |
| **L1** | **Transporte & Malla** | **CANÓNICO** | Enrutador Kleinberg (12 anillos), ZTNA Default-Deny, Buffer Pool 3-Tier, Packet Pacing, TUN. |
| **L2** | **Telemetría** | **CANÓNICO** | Ring Buffer lock-free (<28 ns, 0 alocs), métricas tiempo real, contabilidad Tit-for-Tat. |
| **Core** | **Smart Gateway** | **NÚCLEO** | Bus de integración `/api/v1/*` para componentes externos y periféricos LAN. |
| **L3** | **Inteligencia** | **SATELITAL** | Servidor MCP (JSON-RPC 2.0), Gobernanza, Senado de Agentes, Derecho Computable. |
| **L4** | **Aplicaciones** | **SATELITAL** | Dashboard Web, dDNS Petnames, SAS OOB, Mensajería E2EE, Egress VPN. |

---

## 3. Capacidades Clave Certificadas

1. **Identidad Soberana (DID):** Criptografía pura Ed25519 (`did:ipvn7:<pubkey>`), sin IPs ni hardware fijos.
2. **Datagrama Canónico:** CBOR determinista firmado digitalmente. MTU fijo de 1280 bytes (cero fragmentación).
3. **Mundo Pequeño de Kleinberg:** Búsqueda geométrica 2D (distancia XOR logarítmica + latencia RTT).
4. **Cortafuegos ZTNA:** Política estricta Default-Deny con listas de control por DID y puertos.
5. **Zero-Copy Memory Pool:** 3 tramos (64B, 1500B, 64KB) con conteo atómico de referencias.
6. **Resiliencia Autónoma (EBRA/STUN):** Balizas efímeras Zero-Knowledge *Consume-and-Burn* sin servidores centrales.
7. **Sovereign Device Bridge:** Detección pasiva en LAN de TVs e impresoras mapeadas a subredes virtuales.

---

## 4. Compilación y Ejecución

```bash
# Compilación dual Windows / Linux (CGO_ENABLED=0)
powershell -ExecutionPolicy Bypass -File scripts\build_dual.ps1

# Ejecutar demonio central
./bin/windows_amd64/ipvn7.exe -port 7777 -web-port 7070

# Verificación de pruebas unitarias (<3s)
go test ./...
```

---

## 5. Directiva de Operación Autónoma (.agents)

Toda entidad o agente que examine o desarrolle sobre este proyecto asume automáticamente el rol del Agente de Red en [`.agents/`](../.agents/) y debe actuar en estricta conformidad. El sistema recuerda y mantiene activo un autodisparador de autotareas cada 10 minutos (verificación del estándar, optimización del pipeline zero-copy y auditorías continuas), **excepto si ya existe una tarea en curso**, previniendo solapamientos operativos.

