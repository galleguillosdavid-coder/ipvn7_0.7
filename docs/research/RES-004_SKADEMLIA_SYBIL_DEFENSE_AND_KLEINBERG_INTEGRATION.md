# RES-004: Blindaje Anti-Sybil en DHTs Soberanas: Integración de S/Kademlia Micro-PoW y Anillos Kleinberg

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE DE IDENTIDAD Y ENRUTAMIENTO L0-L1
* **Área de Impacto:** Capa L0 (Generación de DID / Micro-PoW), Capa L1 (Kleinberg Router & Métrica XOR).

---

## 1. El Problema de los Ataques Sybil y Eclipse en Mallas Soberanas

En redes sin servidores centrales ni autoridades de certificación (CAs), un atacante con capacidad de cómputo puede generar millones de identidades virtuales ("Sybils") y posicionarse estratégicamente en el espacio de claves de la DHT para secuestrar rutas o eclipsar nodos legítimos.

---

## 2. Comparativa Técnica de Mecanismos de Admisión

| Mecanismo | Costo Económico / Operativo | Centralización | Viabilidad en ipvn7 |
|---|---|---|---|
| **Autoridad Central (CAs / Tailscale Coordinators)** | Requiere servidor central y cuentas de usuario. | **TOTAL:** El servidor central puede censurar o revocar nodos. | ❌ **INADMISIBLE** (Viola el Mandato Soberano). |
| **Proof-of-Stake / Blockchain** | Requiere tokens, gas fees y consenso pesado de bloques. | **Distribuido pero pesado:** Latencia de segundos/minutos. | ❌ **DESCARTADO** (Sobre-ingeniería anti-pragmática). |
| **S/Kademlia Crypto Puzzle (Static + Dynamic PoW)** | Costo computacional local (ej. 16-20 bits de dificultad en SHA-256 al generar identidad). | **CERO:** Verificable en $O(1)$ por cualquier par sin servidores. | ✅ **ADOPTADO COMO ESTÁNDAR CANÓNICO**. |

---

## 3. Integración en ipvn7

1. **DID Vinculado a Micro-PoW:** Toda clave pública Ed25519 de ipvn7 (`did:ipvn7:<pubkey>`) debe resolver un puzzle estático ligero ($c_1 = 16$ bits de ceros) en menos de 200 ms en CPUs estándar, pero imponiendo un costo prohibitivo de miles de horas de CPU para quien intente generar un ataque Sybil masivo.
2. **Métrica Combinada Kleinberg + XOR:** El enrutamiento navegable en 12 anillos de Kleinberg utiliza la métrica XOR de S/Kademlia como distancia base, garantizando saltos $O(\log^2 N)$ con inmunidad matemática a ataques de eclipse.
