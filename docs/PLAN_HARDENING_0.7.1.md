# Plan de Ejecución HARDENING-0.7.1 (Seguridad Criptográfica y Saneamiento)

> **"Ninguna función criptográfica puede degradar silenciosamente a otra función y seguir declarando el mismo algoritmo. Cero secretos en Git. Realismo factual absoluto."**
> Basado en la segunda auditoría externa independiente ([`docs/auditoria externa.md`](auditoria%20externa.md)).

---

## 🎯 Lista de Verificación HARDENING-0.7.1

- [x] **Check 1: 🔴 Eliminar clave privada de Git y configurar `.gitignore`**
  - Eliminado `keystore/node_identity.key` del árbol de Git (`git rm -f`).
  - Creado `.gitignore` protegiendo `keystore/*.key`, archivos `data/*.log`, credenciales y binarios.
  - Asegurada generación en caliente de identidad local sin distribuir claves privadas en el repositorio.

- [x] **Check 2: 🔴 Eliminar fallback PQC silencioso en `pqc_hybrid.go`**
  - Eliminada la función `deriveFallbackPQC`.
  - Si la clave ML-KEM-768 no tiene 1184 bytes o la encapsulación falla, retorna error explícito y **no establece sesión**.
  - Garantizado que `HybridKEMAlgorithm` contiene exclusivamente ML-KEM-768 real (FIPS 203 nativo) sin degradación silenciosa.

- [x] **Check 3: 🟠 Eliminar confusión de tamaños (128B vs 1088B) en `pqc_hybrid.go`**
  - Unificado `HybridKEMCiphertext` para exigir y validar los 1088 bytes canónicos FIPS 203 (`FullPQCCiphertext`) en `Decapsulate`.
  - Desencapsulado directo con `kp.MLKEMDecapsKey.Decapsulate` con rechazo implícito canónico.

- [x] **Check 4: 🔴 Clarificación de ML-DSA (Firma Post-Cuántica)**
  - Documentado con transparencia en `pqc_signatures.go` y `pqc_hybrid.go` que Ed25519 pura estándar (RFC 8032) es la firma canónica de autenticidad en producción, y el componente reticular es un vector experimental preliminar hasta la estabilización de FIPS 204.

- [x] **Check 5: 🟠 Depuración de `planetary_mesh.go` (Simulación Kleinberg)**
  - Aclarado que `planetary_mesh.go` es un modelo matemático de simulación teórica in-memory de la ley de potencias de Kleinberg ($r=2$) y no una malla planetaria física desplegada.
  - Renombrado el test a `TestPlanetaryMeshCluster_Scale50NodesAndFailureSimulation` para reflejar con precisión honesta lo probado.

- [x] **Check 6: 🟢 Verificación Universal y Cero Regresiones**
  - Ejecutada la compuerta universal `scripts/verify_ipvn7_standard.ps1` con 100% de tests unitarios PASS, Invariante Zero-Copy certificado (35.59 ns/op, 0 B/op, 0 allocs/op) y compilación limpia de `bin/ipvn7.exe`.
