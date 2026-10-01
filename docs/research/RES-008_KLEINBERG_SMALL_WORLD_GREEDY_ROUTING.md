# RES-008: ENRUTAMIENTO GREEDY EN REDES SMALL-WORLD DE KLEINBERG CON MÉTIRCAS RTT

## 1. Contexto y Pregunta de Investigación
¿Cómo puede una red soberana de millones de nodos dispersos en el mundo enrutar paquetes sin servidores centrales de enrutamiento (como DERP de Tailscale) ni tablas globales de estado (BGP/OSPF que colapsan por memoria y churn)?

## 2. Fundamentos Teóricos (Kleinberg Small-World Navigation)
* **Teorema de Kleinberg (Nature 2000):** En una malla augmentada con enlaces de largo alcance cuya probabilidad de conexión decae como $P(u, v) \propto d(u, v)^{-r}$ (donde $r = d$, dimensión de la variedad), el algoritmo de enrutamiento greedy descentralizado alcanza el destino en un tiempo esperado de $O(\log^2 N)$ saltos usando únicamente información local.
* **Optimización RTT Práctica:** Cada salto en la red overlay acumula un RTT físico. Si el enrutador greedy clásico solo minimiza la distancia lógica, puede elegir un nodo lógicamente más cercano pero geográficamente transoceánico, disparando la latencia acumulada.

## 3. Implementación y Síntesis en IPVN7
* `ipvn7` implementa en `pkg/core/routing.go` una estructura de 12 anillos concéntricos con particionamiento exponencial de distancia.
* **Métrica de Costo Compuesta:**
  $$\text{Costo}(vecino) = \log_2(\text{DistanciaID}(vecino, \text{Destino})) \times W_{dist} + \text{RTT}_{ms}(vecino) \times W_{lat}$$
* **Inmunidad a Churn y Caídas Repentinas:**
  - Si un nodo intermedio se desconecta, el prober activo O(1) lo desaloja de la tabla en $<100$ ms.
  - El datagrama salta inmediatamente al siguiente candidato óptimo en el mismo anillo sin recalcular grafos globales.

## 4. Decisión Técnica Adoptada
* Formalizar la métrica compuesta distancia-RTT en el enrutador Kleinberg de `ipvn7`.
* Registrado como DEC-100 en el ADR.

## 5. Fuentes Primarias
* Kleinberg, J. (2000): "Navigation in a small world". Nature, 406(6798), 845-845.
* RFC 7011 / RFC 5444: Generalized Mobile Ad Hoc Network (MANET) Packet/Message Format.
