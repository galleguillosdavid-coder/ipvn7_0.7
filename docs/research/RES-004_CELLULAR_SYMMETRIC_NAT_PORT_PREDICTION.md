# RES-004: Perforación de NAT Simétrico Celular 4G/5G: Algoritmo Birthday Paradox vs Predicción Delta

* **Fecha de Emisión:** 2026-09-25
* **Estado:** ACTIVO / BASELINE PARA PERFORACIÓN WAN RESISTENTE
* **Fuentes Primarias:** Tailscale NAT Traversal Paper, RFC 5389 (STUN), RFC 8445 (ICE), APNIC Research.
* **Área de Impacto:** Capa L1/L2 (NAT Traversal, Resiliencia Celular sin Servidores Centrales).

---

## 1. El Desafío Físico del CGNAT Celular (Dual Symmetric NAT)

En conexiones celulares móviles (4G/5G) y Wi-Fi corporativo, los operadores asignan **NAT Simétrico**:
- Cada nueva conexión a una IP o puerto de destino diferente recibe un puerto de origen público diferente en el CGNAT del operador.
- **Limitación de Soluciones Existentes (Tailscale):**
  - Ante un único NAT simétrico, Tailscale utiliza el *Birthday Paradox* disparando ~2,048 sondas para colisionar puertos con >99% de éxito.
  - **Fallo Crítico:** Ante **Dual Symmetric NAT** (dos teléfonos móviles intentando conectarse directamente), el espacio de búsqueda es bidimensional ($65,535 \times 65,535 \approx 4.3 \times 10^9$ pares), lo que hace inviable el sondeo. Tailscale se rinde y canaliza todo por sus servidores centrales DERP.

---

## 2. La Innovación Soberana de IPVN7

Para mantener la comunicación P2P directa sin recurrir jamás a servidores centrales:

1. **Detección de Delta Secuencial en CGNAT ($\Delta_{seq}$):**
   - La mayoría de CGNATs de operadores (Ericsson, Huawei, Nokia, Linux iptables) asignan puertos de forma secuencial ($\Delta = +1, +2, +16$) o en bloques de subred contiguos.
   - ipvn7 realiza 3 consultas STUN a diferentes reflectores de la malla y calcula la tasa de incremento $\Delta$.
   - Conociendo $\Delta$, el espacio de búsqueda colapsa de $O(N^2)$ a $O(N)$, permitiendo perforar Dual Symmetric NAT en menos de 2 segundos disparando únicamente 32 a 64 paquetes UDP coordinados.

2. **Relay Efímero Descentralizado en Kleinberg (Cero Servidores Centrales):**
   - Si la asignación de puertos es puramente aleatoria (Full Cone Random Allocation), ipvn7 no usa servidores de una empresa; utiliza cualquier nodo soberano de la malla en los 12 anillos de Kleinberg como relay efímero cifrado con MASQUE RFC 9298 en puerto 443.

---

## 3. Conclusión y Plan de Acción (DEC-096)

* Integrar la detección de delta en `pkg/l1/nat_traversal.go`.
* Preservar el presupuesto Zero-Copy (0 B/op) y mantener el tiempo de convergencia por debajo de 2.5 segundos.
