# MEMORIA HISTÓRICA Y LECCIONES APRENDIDAS — IPVN7 (v0.1 a v0.6)

> **Estado:** ARCHIVO HISTÓRICO CONSOLIDADO  
> **Fecha de Consolidación:** Octubre 2026  
> **Alcance:** Génesis, evolución y lecciones de versiones preliminares  

---

## 1. GÉNESIS DEL PROYECTO
El proyecto IPVN7 nació con el objetivo de diseñar un **Sistema Operativo de Red Autónomo, Descentralizado y Criptográficamente Soberano**.
En sus versiones preliminares (v0.1 a v0.6), el desarrollo exploró múltiples paradigmas:
* Integración monolítica de servicios de chat, archivos y VPN.
* Intentos de enrutamiento basados en inundación (flooding) y DHTs no acotadas.
* Criptografía heterogénea y mutación dinámica de estados de pares sin autenticación mutua rigurosa.

---

## 2. LECCIONES FUNDAMENTALES DE INGENIERÍA

1. **La falacia del monolito de red:** Incorporar aplicaciones dentro del núcleo de transporte degrada el rendimiento, genera condiciones de carrera y rompe el invariante de memoria zero-copy.
2. **El principio de velocidad constante:** El Core debe operar como un reloj suizo: datagramas deterministas de <= 1280 bytes, preasignación estricta de memoria (0 alocaciones por paquete) y $O(1)$ en descarte de paquetes malformados.
3. **Identidad no es Autorización:** La verificación de firma digital Ed25519 no puede autorizar automáticamente el tráfico de red; el modelo ZTNA Default-Deny es indispensable para prevenir suplantaciones y movimiento lateral.
4. **La regla de las 400 líneas (Axioma III):** Archivos extensos ocultan bugs de concurrencia y dependencias circulares. La modularidad radical es la mejor defensa arquitectónica.
5. **No inventar ni proclamar victorias prematuras:** La honestidad taxonómica (distinguir HECHO de HIPÓTESIS) es el único camino hacia una ingeniería de grado de producción.

---

## 3. TRANSICIÓN A LA ARQUITECTURA v0.7
Todo el aprendizaje previo decantó en la versión **0.7.0**, donde el Núcleo I7 queda formalmente congelado en sus **10 primitivas inmutables**, el control plane se rige por [`sistema/CONSTITUCION.md`](../../sistema/CONSTITUCION.md) y la orquestación recae en el Agente Rector [FrondaBrick_01](../../agentes/Frondabrick01/AGENTE.md).
