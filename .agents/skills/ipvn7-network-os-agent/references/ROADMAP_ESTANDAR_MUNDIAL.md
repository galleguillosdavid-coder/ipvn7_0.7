# ROADMAP DE ESTANDARIZACIÓN MUNDIAL: IPVN7 NETWORK OS

Este documento traza las 7 fases fundamentales para consolidar a **ipvn7** como el estándar de facto en sistemas operativos de red soberanos, descentralizados y post-cuánticos.

---

## Fase 1: Consolidación del Núcleo Zero-Copy (Nivel Wire-Speed) [COMPLETADA]
- [x] Arquitectura de Pipeline en 5 capas desacopladas (L0 Transporte, L1 Cripto, L2 Enrutamiento, L3 Semántica, L4 Aplicación).
- [x] Implementación de un `sync.Pool` de buffers y reciclaje de `PacketContext` con **0 allocs/op** en la ruta crítica de ingesta y despacho.
- [x] Optimización SIMD / Vectorial de operaciones de checksum (BLAKE2s RFC 7693) y verificación en tiempo constante con 0 allocs/op (> 493 MB/s).



---

## Fase 2: Protocolo Criptográfico Post-Cuántica Nivel 3 (NIST PQC) [COMPLETADA]
- [x] Soporte base ML-KEM-768 (Kyber) y ML-DSA-65 (Dilithium).
- [x] Mecanismo de handshake post-cuántico híbrido en 1 RTT (X25519 + ML-KEM-768 a 0.25 ms).
- [x] Ventana anti-replay determinista de 1024 bits multi-palabra (RFC 4303) con tolerancia a jitter extremo y 0 allocs/op.



---

## Fase 3: Anonimato Fisiológico y Rutas Cebolla (Sphinx 3-hop) [COMPLETADA]
- [x] Encapsulación de tramas fijas de 1280B para derrotar análisis de flujo por tamaño.
- [x] Cifrado multicapa Sphinx real entre nodos de la malla P2P para tráfico sensible (Guard -> Middle -> Exit).
- [x] Intervalos estocásticos de transmisión de tramas dummy/padding para neutralizar ataques de correlación temporal y DPI (Deep Packet Inspection).


---

## Fase 4: Desacople UIN y Resolución Distribuida [COMPLETADA]
- [x] Identificador soberano `did:ipvn7:<pubkey>` desacoplado de IP física.
- [x] Gossip protocol de propagación de `BindingRecords` con expiración criptográfica y firmas delegadas (Fanout $k=3$ y deduplicación $O(1)$).
- [x] Mecanismo de agujereado NAT (NAT Hole Punching) bidireccional UDP (STUN/ICE descentralizado) sin servidores de señalización centralizados.


---

## Fase 5: VPN Corporativa Fricción Cero en Espacio de Usuario [COMPLETADA]
- [x] Servidores integrados SOCKS5 (`:10807`) y HTTP CONNECT (`:10808`).
- [x] Soporte de conmutación sin permisos de root en Windows, Linux y macOS.
- [x] Enmascaramiento de tráfico saliente simulando perfiles TLS 1.3 auténticos (RFC 8446).

---

## Fase 6: Gobernanza del Árbitro de Memoria contra DoS [COMPLETADA]
- [x] Cuotas estrictas de RAM por subsistema (Replay 20%, QoS 30%, Trust 20%, Bindings 20%).
- [x] Pruebas de estrés y saturación de enlaces con 100,000 datagramas/seg sin fugas de memoria ni desbordamiento de colas (2.84M pps logrados).

---

## Fase 7: Especificación Formal RFC e Interoperabilidad Universal [COMPLETADA]
- [x] Redacción de la especificación técnica en formato RFC de Internet (`RFC_IPVN7_CORE.md` y `RFC_IPVN7_SPHINX.md`).
- [x] SDK multiplataforma universal (Go `sdk/go/`, Python `sdk/python/`, Rust `sdk/rust/`).
- [x] Conectores nativos para agentes de Inteligencia Artificial (MCP Server en `pkg/l3` y REST API en `pkg/core`).

---

## Fase 8: Resiliencia Planetaria y Verificación Continua en Dos Nodos [EN EJECUCIÓN]
- [x] Protocolo de verificación física bidireccional certificado entre Nodo A (192.168.1.198) y Nodo B (192.168.1.106).
- [x] Demonio autodisparador continuo cada 10 minutos con cerrojo de exclusión mutua (`scripts/run_autonomous_daemon.ps1`).
- [ ] Monitoreo automatizado de degradación de enlaces WAN con auto-reparación vía rutas alternativas Kleinberg.
- [ ] Empaquetado binario reproducible y contenedor sin privilegios (`scratch`/musl) para despliegues perimetrales.

