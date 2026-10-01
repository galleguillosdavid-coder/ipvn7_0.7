/* ==========================================================================
   GUÍA INTERACTIVA SOBERANA IPVN7 - DICCIONARIO COGNITIVO TRI-NIVEL
   Nivel 1: Simple (0 jerga) | Nivel 2: Analogía & Ejemplo | Nivel 3: Técnico / IA
   Axioma III: <250 líneas | Vanilla JS | Realismo Físico Exacto
   ========================================================================== */

const EXPLANATIONS = {
  p2p: {
    title: "De Tú a Tú (P2P Real): Conexión Directa",
    subtitle: "Cómo viaja la información entre tus dispositivos sin intermediarios",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>Es una conexión directa entre tus dos aparatos. Tu foto, mensaje o archivo va directamente desde tu casa hasta la de tu amigo, <b>sin pasar por la computadora de ninguna empresa ni nadie en medio</b>. Como hablar cara a cara sin que nadie escuche.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>Imagina que tienes una ventana frente a la de tu vecino y tienden un hilo con dos latas para hablar directamente. En la Internet tradicional de hoy, para decirle "Hola" a tu vecino, tienes que subirte a un avión, viajar 8.000 km hasta una oficina central en EE.UU., hacer una fila, abrir tu carta para que la revisen y luego esperar a que un cartero se la lleve a tu vecino. <b>Con IPvN7 hablas por tu propio hilo directo al instante.</b></p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Implementa perforación NAT (UDP Hole Punching según RFC 5389 / STUN serverless). Los nodos descubren recíprocamente sus sockets externos <code>IP:Puerto</code> mediante señalización efímera autenticada. Una vez establecida la paridad, los datagramas L1 fluyen por la <b>geodésica de fibra óptica más corta posible (~50 µs de propagación)</b> con $t_{RTT} = 2 \cdot d / c_{glass}$, eliminando relays de terceros y garantizando latencia mínima física sin colas de procesamiento intermedio.</p>
    `
  },

  routers: {
    title: "Lo que Ve y lo que NO Puede Leer el Enrutador del ISP",
    subtitle: "La barrera matemática ante los intermediarios de telecomunicaciones",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>Tu empresa de internet solo ve que envías cajitas idénticas y cerradas con candado. <b>No puede saber qué haces adentro</b>: no sabe si estás jugando, viendo fotos de tu familia, trabajando o haciendo una llamada.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>Es como enviar correspondencia dentro de sobres de titanio sellados al vacío, todos con el mismo peso, color y tamaño exacto. El cartero sabe tu dirección de casa y a qué casa la llevas, pero no puede abrir el sobre, no sabe si lleva un contrato o una receta, y si intenta romperlo, la carta se autodestruye en el acto.</p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Carga útil blindada con <b>ChaCha20-Poly1305 (RFC 8439)</b> con claves efímeras derivadas de <b>NIST ML-KEM-768</b>. <b>Padding rígido a 1280 bytes</b> que neutraliza ataques de inferencia por análisis de longitud de tráfico (*traffic analysis padding*). Autenticación de mensaje Poly1305 por paquete: 1 bit alterado produce descarte inmediato en 0 nanosegundos (Zero-Alloc). Camuflaje TLS 1.3 (RFC 8446) en puerto 443 ante inspección profunda de paquetes (DPI).</p>
    `
  },

  quantum: {
    title: "Blindaje Post-Cuántico (ML-KEM-768 / Kyber)",
    subtitle: "Inmunidad criptográfica ante las supercomputadoras del futuro",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>Es un candado matemático del futuro. Ni siquiera las supercomputadoras cuánticas más potentes que inventen en los próximos 50 años podrán abrir ni descifrar lo que envíes hoy por esta red.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>La seguridad de los bancos y páginas web de hoy es como una cerradura tradicional de llave: una computadora cuántica es como una ganzúa mágica que la abre en 2 segundos. IPvN7 en cambio es como un laberinto en 768 dimensiones geométricas: aunque la supercomputadora intente millones de caminos simultáneos, es matemáticamente imposible resolverlo.</p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Implementación del estándar oficial <b>NIST FIPS 203 (ML-KEM-768)</b> basado en la dureza matemática de retículos modulares con aprendizaje de errores (Module-LWE). Complejidad de ataque cuántico $\\ge 2^{192}$ operaciones de compuerta. Neutraliza por diseño el vector de amenaza <i>"Harvest Now, Decrypt Later"</i> (cosechar hoy tráfico cifrado clásico RSA/ECC para descifrarlo en la era cuántica con el algoritmo de Shor).</p>
    `
  },

  kleinberg: {
    title: "Malla Indestructible Kleinberg O(log² N)",
    subtitle: "Reconexión autónoma en 1 milisegundo sin servidores de rutas",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>Si se corta un cable de internet o se apaga una antena en la ciudad, tu conexión no se cae: busca automáticamente otro camino en un abrir y cerrar de ojos sin que te des cuenta.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>Piensa en el tráfico de una ciudad. Si una avenida principal se inunda repentinamente, tu sistema de navegación te desvía al instante por una calle lateral despejada antes de que tengas que frenar, sin necesidad de esperar a que un policía central te autorice a doblar.</p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Enrutamiento descentralizado basado en el modelo de <b>Pequeños Mundos de Jon Kleinberg</b> con enlaces de largo alcance distribuidos según probabilidad $P(u,v) \\propto d(u,v)^{-r}$ ($r=2$). La convergencia de entrega entre trillones de nodos está acotada en saltos $O(\\log^2 N)$, sin requerir tablas de rutas globales BGP monolíticas ni puntos únicos de falla. Failover dinámico en $<1$ ms al detectar pérdida de pulso UDP.</p>
    `
  },

  zen: {
    title: "Control Zen y Bio-Ritmo Cardíaco a 60 FPS",
    subtitle: "Telemetría vital humana en tiempo real sin engaños de interfaz",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>En lugar de pantallas llenas de números raros que nadie entiende, tienes <b>un solo botón y un latido como el de tu corazón</b>. Si está verde y late, todo funciona; si el cable se desconecta, se vuelve rojo y la línea queda plana. Cero confusiones.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>Como el monitor de signos vitales de un hospital. El doctor no se pone a leer 500 páginas de registros técnicos: mira la pantalla y ve una onda limpia que late. Si hay pulso, el paciente está vivo y saludable. Si se detiene, suena la alarma de inmediato.</p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Pipeline de telemetría L2 con cálculo determinista de pulso: $\\text{BPM} = 68 + \\min(\\text{KB/s} \\times 1.5, 52)$. Renderizado en Canvas 2D acelerado por hardware a 60 FPS. <b>Axioma de Falsabilidad HIL (Zero Fake Toasts)</b>: el estado visual está estrictamente acoplado a la respuesta física del socket. Ante timeout o pérdida de datagrama, conmuta en 0 ms a asistolia (línea plana roja a 0 BPM). Prohibido emitir feedback optimista simulado.</p>
    `
  },

  egress: {
    title: "Acelerador de Salida Soberana (Egress Gateway)",
    subtitle: "Navegación comunitaria de alta velocidad y evasión de congestión local",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>Si el internet de tu casa o trabajo anda lento o bloquea una página que necesitas, puedes <b>salir a internet usando la conexión despejada de un amigo o vecino</b> de confianza con solo presionar un botón.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>Si la salida del estacionamiento de tu edificio está bloqueada por obras o basura, tu vecino de al lado te abre el portón de su casa para que salgas directamente a la autopista rápida sin perder tiempo.</p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Pasarela de egreso L4 con tunelización transparente y control de congestión <b>BBR (Bottleneck Bandwidth and RTT)</b>. Evalúa continuamente jitter, pérdida de tramas y RTT frente al enlace local. Si el gateway remoto ofrece mayor throughput por un margen superior a la histéresis del 25%, el flujo se conmuta por el túnel cifrado PQC. Si el nodo remoto se desactiva, un centinela ejecuta fallback local en $<500$ ms sin desconectar sockets TCP establecidos.</p>
    `
  },

  shadow: {
    title: "Dispositivos Sombra para el Hogar y la Oficina (Shadow DIDs)",
    subtitle: "Protección perimetral de Smart TVs, impresoras y cámaras sin tocar su software",
    simple: `
      <h4>🌱 1. En palabras muy simples:</h4>
      <p>Protege la tele, la impresora y las cámaras de seguridad de tu casa para que <b>ningún hacker en el mundo pueda verlas ni atacarlas</b>, pero tú y tu familia puedan usarlas normalmente desde cualquier lugar.</p>
    `,
    analogy: `
      <h4>💡 2. Ejemplo y Analogía Cotidiana:</h4>
      <p>Es como ponerle una capa de invisibilidad mágica y un guardia de seguridad privado a la impresora de tu casa. Para cualquier extraño que pase por la calle o por internet, la impresora simplemente no existe en el universo. Solo tú y tus equipos de confianza tienen los lentes especiales para verla y usarla.</p>
    `,
    technical: `
      <h4>🔬 3. Especificación Técnica (Ingeniero / Agente IA):</h4>
      <p>Aislamiento ZTNA (Zero Trust Network Access) mediante identificadores virtuales <code>did:ipvn7:shadow:&lt;sha256&gt;</code> mapeados en la subred virtual <code>10.7.100.0/24</code>. El nodo anfitrión actúa como Embajador Guardián (Rol L), reenviando flujos L4 zero-copy e inyectando datagramas Wake-on-LAN (WoL) para encendido remoto. Los dispositivos IoT permanecen con puertos completamente cerrados hacia Internet pública (Default-Deny estricto).</p>
    `
  }
};
