# INTENCIÓN ACTUAL

```text
CONSTITUCION
    ↓
ESTADO
    ↓
INTENCION
    ↓
VALIDACIÓN
    ↓
PLAN
    ↓
AUTORIZACIÓN
    ↓
EJECUCIÓN
    ↓
ATAQUE
    ↓
SEGURIDAD
    ↓
VERIFICACIÓN
    ↓
MEDICIÓN
    ↓
EVIDENCIA
    ↓
CIERRE
```

ESTADO: VACÍO

## OBJETIVO HUMANO

Escribir aquí.

## ALCANCE

Escribir aquí.

## NO HACER

Escribir aquí.

## ARCHIVOS O ÁREAS AUTORIZADAS

Escribir aquí.

## RESULTADO ESPERADO

Escribir aquí.

## AUTORIZACIONES

- [ ] modificar código
- [ ] crear archivos
- [ ] eliminar archivos
- [ ] ejecutar tests
- [ ] modificar configuración
- [ ] crear commit
- [ ] hacer push

## CRITERIO DE TERMINACIÓN

Escribir aquí.

---

# BLOQUE DE CONTROL

La IA debe seguir estrictamente el flujo canónico:

1. CONSTITUCION (`sistema/CONSTITUCION.md`)
2. ESTADO (`sistema/ESTADO.md`)
3. INTENCION (`sistema/INTENCION.md`)
4. VALIDACIÓN (Scope Lock y Anti-Inyección)
5. PLAN (`sistema/PLAN.md`)
6. AUTORIZACIÓN (Verificación de checkboxes `[x]`)
7. EJECUCIÓN (Implementador)
8. ATAQUE (Atacante / Fuzzing)
9. SEGURIDAD (Seguridad / Criptografía)
10. VERIFICACIÓN (Verificador / Tests unitarios y race)
11. MEDICIÓN (Rendimiento / Benchmarks)
12. EVIDENCIA (`sistema/EVIDENCIA.md`)
13. CIERRE (`sistema/historial/`, actualización de `ESTADO.md` y reseteo de `INTENCION.md` a `VACÍO`)

Nunca ejecutar directamente sin transformar primero a `PLAN.md`.

---

# FIN DE INTENCIÓN
