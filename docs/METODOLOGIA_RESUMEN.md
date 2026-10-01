# Metodología de Investigación y Evolución IPVN7 (RESUMEN)

> "Más inteligente que documentar tareas es documentar cómo descubrirlas y de qué fuentes basarse."

## Objetivo
IPVN7 lidera mediante vigilancia tecnológica exógena, no invención en aislamiento. Sistema determinista de descubrimiento e investigación proactiva.

## 6 Ejes Temáticos Fundacionales

### 1. Criptografía PQC & Identidad Soberana (L0/L1)
- **Fuentes:** NIST CSRC (FIPS 203 ML-KEM, FIPS 204 ML-DSA), IETF CFRG (X-Wing KEM), Ed25519/X25519
- **Pregunta:** ¿Garantizar confidencialidad post-cuántica sin superar MTU 1280B?

### 2. Rendimiento Zero-Copy (L0/L1)
- **Fuentes:** Windows IOCP/RIO, Linux io_uring/eBPF/XDP, sync.Pool, BBRv3 pacing
- **Pregunta:** ¿Conmutar paquetes wire-speed con 0 B/op y <40ns/op en userspace?

### 3. Resiliencia WAN Hostil (L1/L2)
- **Fuentes:** IETF RFCs (STUN 5389, ICE 8445, MASQUE 9298, TLS 1.3 8446), anti-DPI
- **Pregunta:** ¿Atravesar CGNAT móvil/firewall restrictivo sin servidores centrales?

### 4. Enrutamiento Descentralizado (L1/L2)
- **Fuentes:** Kleinberg small-world (12 anillos), S/Kademlia Sybil-resistant, CRDTs, gossip
- **Pregunta:** ¿Ruta óptima O(log²N) sin conocimiento topológico completo?

### 5. UX Radical Zero-Friction
- **Fuentes:** Agentes Persona, diseño humano, lenguaje simple, onboarding guiado
- **Pregunta:** ¿Instalar y usar en 30 segundos sin jerga técnica?

### 6. Hardware & Vehículos
- **Fuentes:** MAVLink v2, CAN Bus ISO 11898, OBD-II, J1939, WoL autenticado
- **Pregunta:** ¿Control robótica con telemetría tiempo real RFC 9221?

## Filtro Antihumo (4 Pasos)
1. **Realismo Físico:** Sin nube privativa, ejecutable único
2. **Algoritmo 5 Pasos:** Poda previa 30-50% antes de codificar
3. **Invariante Zero-Copy:** ≤400 líneas, 0 B/op hot-path
4. **Falsabilidad HIL:** Camino de error probado físicamente

## Ciclo Operativo Trazable
```
Búsqueda Web (RFCs/NIST) → Ficha RES-XXX → ADR DEC-XXX → Tarea TASK-XXX
```

---

**Última actualización:** 2026-09-25  
**Vigilancia:** Radar IETF/NIST + repositorios líderes (WireGuard, Tailscale, Linux eBPF)
