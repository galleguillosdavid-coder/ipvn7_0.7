# RES-016: AUTORIZACIÓN ADAPTATIVA PARA IA AGÉNTICA Y CADENAS DE DELEGACIÓN CRIPTOGRÁFICAS EN IPVN7

## 1. Contexto y Pregunta de Investigación
Hacia finales de 2026, la estandarización de agentes de IA ha consolidado a **MCP** (Model Context Protocol, LF) para acceso a herramientas y a **A2A** (Agent2Agent, AAIF) para coordinación entre agentes. Sin embargo, la industria enfrenta lo que los investigadores del IETF y NIST denominan la **"Brecha de Gobernanza y Autorización" (Governance & Authorization Gap)**:
- Los protocolos actuales permiten que los agentes se descubran e invoquen herramientas sin fricción, pero carecen de mecanismos nativos en el formato de cable (*wire-format*) para **delegación multi-salto, atenuación de privilegios criptográfica y revocación verificable**.
- Si un agente orquestador delega una tarea a un sub-agente especializado, y este a su vez a un tercer agente, el receptor final (un ejecutor o herramienta crítica) no tiene forma criptográfica de verificar si los permisos fueron atenuados o si hubo escalamiento no autorizado de privilegios.

¿Cómo puede IPvN7 aprovechar su arquitectura soberana basada en **DIDs W3C, firmas Ed25519 y cifrado post-cuántico ML-KEM-768** para cerrar esta brecha y ofrecer el sustrato de autorización adaptativa y delegación más seguro y liviano del mundo?

## 2. Hallazgos en Estándares Mundiales de Frontera (Septiembre 2026)
1. **IETF Internet-Draft `draft-das-agentic-adaptive-authorization-00` (23 Septiembre 2026):**
   - *Título:* "Adaptive Authorization for Agentic AI: Graduated and Escalated Execution Control for Critical Infrastructure".
   - *Concepto Nuclear:* Rompe con el modelo binario de permitir/denegar (*allow/deny*). Clasifica las acciones en **Niveles Graduados de Ejecución**: *Ordinary*, *Reduced*, *Escalated*, *Canary*, *Sandbox*, *Review*, *Quarantine*, *Deny*.
   - *Handles de Ejecución Ligados a Destino:* Genera identificadores atados criptográficamente al sumidero de la acción (*sink-bound execution handles*) para garantizar que una llamada a una herramienta no pueda ser desviada o reutilizada fuera de contexto.
2. **IETF Internet-Draft `draft-liu-ai-agent-authorization-integration-00` & `draft-klrc-aiagent-auth-03`:**
   - Proponen la integración de identidades de carga de trabajo (WIMSE) con extensiones OAuth para rastrear la cadena de custodia de la delegación en entornos distribuidos.
3. **Investigación "Bounded Agents: Delegation Security for Multi-Agent AI Systems" (Agosto 2026):**
   - Introduce la **Agentic Principal Chain (APC)**: una estructura enlazada donde cada agente firma criptográficamente el alcance atenuado concedido al siguiente salto, previniendo que un agente subordinado ejecute acciones con privilegios mayores a los de su invocador original.

## 3. Oportunidades y Beneficios Estratégicos para IPvN7
1. **Cadena de Delegación Agéntica Soberana (PQC-APC Token):**
   - Cada nodo y cada agente en IPvN7 posee un DID soberano (`did:ipvn7:...`). En lugar de depender de servidores OAuth centralizados o nubes privativas, IPvN7 puede adjuntar un token compacto firmado con **Ed25519 y blindado con ML-KEM-768** directamente en los payloads A2A (`tasks/send`) y en las llamadas MCP (`tools/call`).
2. **Autorización Adaptativa ZTNA Integrada (Graduated Control):**
   - El motor ZTNA de IPvN7 aplica automáticamente control graduado:
     - **Acciones Ordinarias (Lectura de red, telemetría):** Aprobadas inmediatamente si el DID está autenticado.
     - **Acciones Escaladas (Conectar/Desconectar VPN, alterar rutas, puente LAN):** Requieren firma criptográfica del DID propietario de la máquina o token de delegación válido con alcance explícito.
     - **Acciones Cuarentena/Sandboxed (Agentes externos no verificados):** Confinadas a un proxy en userspace sin acceso a sockets locales.
3. **Cero-Fricción y Cero-Tokens de API:**
   - Toda la verificación de la cadena APC se ejecuta matemáticamente en microsegundos en el nodo receptor mediante criptografía de clave pública local, sin consultar APIs externas, garantizando latencia casi nula y soberanía absoluta.

## 4. Filtro Antihumo Estricto y Presupuesto de Código
- **Sin Dependencias Pesadas:** No se implementarán servidores de autorización monolíticos ni motores complejos de políticas OPA/Rego.
- **Implementación Compacta en Go:** Se define una estructura mínima de token de delegación (`DelegationToken`) de menos de 40 líneas en L0/L1, encapsulando `IssuerDID`, `SubjectDID`, `Scope`, `ExpiresAt` y `Signature`.
- **Invariante Zero-Copy:** El procesamiento y validación del token opera sobre buffers preasignados sin alocaciones de memoria en la ruta caliente.

## 5. Decisión Técnica Recomendada
- Registrar **DEC-125**: Adopción de la Cadena de Delegación Agéntica Criptográfica (PQC-APC) y Control Graduado ZTNA según especificaciones IETF `draft-das-agentic-adaptive-authorization-00` en IPvN7.

## 6. Fuentes Primarias
- IETF Internet-Draft: `draft-das-agentic-adaptive-authorization-00` (Septiembre 2026).
- IETF Internet-Draft: `draft-liu-ai-agent-authorization-integration-00` (2026).
- IETF Internet-Draft: `draft-klrc-aiagent-auth-03` (2026).
- ArXiv: "Bounded Agents: Delegation Security for Multi-Agent AI Systems" (Agosto 2026).
- Linux Foundation: Agentic AI Foundation (AAIF) & Model Context Protocol (MCP) Working Group.
