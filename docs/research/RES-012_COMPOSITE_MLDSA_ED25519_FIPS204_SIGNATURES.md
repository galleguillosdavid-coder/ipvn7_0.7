# RES-012: FIRMAS HÍBRIDAS COMPUESTAS POST-CUÁNTICAS (ED25519 + ML-DSA FIPS 204 Y RFC 9955)

## 1. Contexto y Pregunta de Investigación
¿Cómo proteger las identidades soberanas de los nodos de `ipvn7` y los manifiestos de auto-actualización contra falsificación por ordenadores cuánticos (amenaza Q-Day), sin degradar el throughput de paquetes ni fragmentar las tramas fijas de 1280B?

## 2. Hallazgos en Estándares Mundiales de Frontera
* **NIST FIPS 204 (ML-DSA):** Estándar formalizado de firma digital basado en retículos (Module-Lattice-Based Digital Signature Algorithm - Dilithium).
* **IETF RFC 9955 & draft-ietf-lamps-pq-composite-sigs:** Define la construcción compuesta dual:
  $$\text{FirmaCompuesta} = \text{Ed25519\_Sign}(SK_{ed}, M) \parallel \text{ML-DSA\_Sign}(SK_{ml}, M)$$
  Garantiza seguridad inquebrantable: un atacante cuántico no puede romper la parte ML-DSA, y cualquier fallo criptoanalítico teórico en retículos sigue protegido por Ed25519.
* **Separación de Planos (Control vs Datos):**
  - **Plano de Control / Identidad:** Firmas compuestas para certificados, vinculación (SAS RFC 6189) y binarios de actualización.
  - **Plano de Datos en Vuelo:** Prohibición de firmas PQC por datagrama (para no inflar tramas más allá de 1280B). El tráfico se autentica simétricamente con Poly1305 (16B) usando claves post-cuánticas derivadas de X-Wing KEM (FIPS 203).

## 3. Implementación y Síntesis en IPVN7
* `ipvn7` adopta este modelo en la validación de identidades y en el subsistema de actualización atómica con reversión (`pkg/core/identity.go` y `auto_update`).
* Mantiene el presupuesto zero-copy: los datos en vuelo procesados en L0-L2 conservan 0 alocaciones y sub-40ns por paquete.

## 4. Decisión Técnica Adoptada
* Adoptar la especificación de firmas compuestas Ed25519 + ML-DSA (FIPS 204 / RFC 9955) para identidades soberanas y manifiestos de actualización.
* Registrado como DEC-104 en el ADR.

## 5. Fuentes Primarias
* NIST FIPS 204: Module-Lattice-Based Digital Signature Standard (ML-DSA).
* IETF RFC 9955: Hybrid Signature Spectrums.
* IETF draft-ietf-lamps-pq-composite-sigs: Composite Signatures For Use In Internet PKI.
