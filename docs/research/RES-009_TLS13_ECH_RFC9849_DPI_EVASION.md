# RES-009: CAMUFLAJE TLS 1.3 Y EVASIÓN DPI AVANZADA MEDIANTE ENCRYPTED CLIENT HELLO (RFC 9849 ECH)

## 1. Contexto y Pregunta de Investigación
¿Cómo debe camuflarse el tráfico de `ipvn7` cuando atraviesa firewalls corporativos hostiles o censura estatal (Great Firewall) que bloquean puertos UDP y emplean inspección profunda de paquetes (DPI L7) basada en Server Name Indication (SNI)?

## 2. Hallazgos en Estándares Mundiales de Frontera
* **RFC 9849 (IETF Standard):** *"Encrypted Client Hello (ECH) for TLS 1.3"* divide el handshake TLS en:
  - **ClientHelloOuter:** Visible, inocuo y dirigido a dominios/CDNs benignos (ej. cloudflare.com, google.com).
  - **ClientHelloInner:** Cifrado con HPKE (RFC 9180), conteniendo la identidad de red soberana o datagrama inicial de `ipvn7`.
* **RFC 9848:** Uso del registro HTTPS de DNS para distribuir de forma autenticada las llaves de configuración ECH.
* **Técnica GREASE:** Generación de extensiones ECH sintéticas para evitar que el DPI clasifique el tráfico por presencia/ausencia de campos opcionales.

## 3. Integración en la Arquitectura IPVN7
* En `pkg/l1/tls_options.go` y `pkg/l1/tls_masquerade.go`, `ipvn7` implementa camuflaje TLS 1.3 puro sobre puerto 443 TCP/UDP (HTTP/3).
* **Mejora ECH RFC 9849:**
  - El datagrama inicial de handshake X-Wing KEM se encapsula en la extensión ECH (`0xfe0d` / GREASE) dentro de una trama TLS 1.3 canónica.
  - El firewall observa un flujo HTTPS 100% válido y sintácticamente idéntico al de un navegador moderno (Chrome/Firefox).
  - Zero-Copy Invariante: El envoltorio utiliza buffers fijos prealocados de 1280B sin copias intermedias.

## 4. Decisión Técnica Adoptada
* Adoptar la especificación RFC 9849 ECH para la evasión de inspección SNI en el motor de camuflaje de `ipvn7`.
* Registrado como DEC-101 en el ADR.

## 5. Fuentes Primarias
* IETF RFC 9849: Encrypted Client Hello (ECH) for TLS 1.3
* IETF RFC 9848: Indicating HTTPS Support in DNS (SVCB and HTTPS RR Types)
* IETF RFC 9180: Hybrid Public Key Encryption (HPKE)
