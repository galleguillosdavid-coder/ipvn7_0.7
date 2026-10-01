# CONTRATO DEL CORE: CONTAINER

## 1. ¿Qué representa?
La envoltura compacta y estandarizada que agrupa uno o más objetos lógicos para su transmisión segura sobre el transporte físico.

## 2. ¿Qué datos contiene?
* `ContainerID`: Identificador de contenedor (uint64 o UUID hash).
* `EncryptedFlags`: Máscara de bits de compresión, cifrado y perfil.
* `ObjectCount`: Cantidad de objetos contenidos.
* `Objects`: Lista de objetos lógicos tipados serializados.
* `AuthTag`: Tag Poly1305 / MAC de integridad del contenedor completo.

## 3. ¿Quién lo crea?
La capa de empaquetado L0/L1 antes de entregar la trama al adaptador de transporte (UDP/TUN).

## 4. ¿Quién lo modifica?
Inmutable. El contenedor no puede ser fraccionado ni modificado por nodos intermedios.

## 5. ¿Quién lo destruye?
El pipeline de recepción L0 tras desempaquetar y validar la integridad de los objetos contenidos.

## 6. ¿Qué invariantes tiene?
* Tamaño del contenedor con cabeceras `<= 1280` bytes.
* Verificación de autenticidad en tiempo constante antes de procesar objetos individuales.

## 7. ¿Qué errores puede producir?
* `ErrContainerOverflow`: Tamaño acumulado de objetos supera el MTU.
* `ErrAuthenticationFailed`: Tag de autenticación inválido.
* `ErrMalformedContainer`: Desalineación de offsets de objetos.

## 8. ¿Cómo se prueba?
* `src/pkg/l0/l0_test.go`: Empaquetado y desempaquetado de contenedores.
* `src/pkg/core/pipeline_bench_test.go`: Desempaquetado sin copias en caliente.
