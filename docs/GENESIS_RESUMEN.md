# Genesis IPVN7 - Arquitectura Conceptual (RESUMEN)

## ¿Qué es IPVN7?
IPVN7 es un **Sistema Operativo de Red Autónomo** (Network OS) que opera como malla overlay de escala planetaria. No es una VPN tradicional ni librería de cifrado: es una red soberana programable.

## Ruptura Radical con IPv4/IPv6
**Problema tradicional:** IP amalgama identidad + localización. Cambios de red (Wi-Fi→4G) destruyen sesiones y exponen ubicación.

**Solución IPVN7:** Separación total. Un dispositivo es exclusivamente su **DID (Identificador Descentralizado Soberano)**. La Internet física es sustrato de transporte transitorio.

## Mecánica Operativa

### 1. Nacimiento del Nodo (Autonomía Criptográfica)
- Genera par Ed25519 sin servidor central
- DID = clave pública (32 bytes, 256 bits)
- Keystore local protegido
- Independencia total de IP física

### 2. Descubrimiento de Pares
- **LAN Discovery:** Balizas UDP <2ms en red local
- **NAT Traversal:** STUN + UPnP para cruce de firewalls
- **DHT Kademlia:** Descubrimiento global 256-bit XOR
- **DERP Relays:** Fallback para CGNAT simétrico

### 3. Canal Seguro (Noise XX + PQC)
- **3 etapas:** Clave efímera → Autenticación → Secreto compartido
- **PFS (Perfect Forward Secrecy):** Claves efímeras destruidas en RAM
- **PQC Inyectado:** ML-KEM/Kyber post-cuántico paralelo

### 4. Adaptador TUN/TAP
- **IPv4 Virtual:** 10.7.0.0/16 con ARP sintético <1µs
- **IPv6 Soberano:** fd07::/64 derivado de SHA-256(DID)
- **MTU Fijo:** 1280 bytes (anti-fragmentación)

### 5. Movilidad Extrema (Roaming sin Caídas)
- Cambio Wi-Fi→4G: actualización firmada a pares
- Sesiones activas continúan sin desconexión
- **Certificado:** 0% pérdida, 1.10ms latencia conmutación

## Paradigma Fundamental
**Identidad ≠ Localización**

| Paradigma | IP Tradicional | IPVN7 |
|-----------|---------------|--------|
| Identidad | IP física | DID Ed25519 |
| Localización | Cambia con red | Sustrato transitorio |
| Sesiones | Roto por cambio IP | Persistente |
| Privacidad | Exposición geográfica | Soberana |

---

**Conclusión:** IPVN7 es una red donde tú eres tu clave criptográfica, no tu dirección IP. El sustrato físico es irrelevante; la identidad es permanente.
