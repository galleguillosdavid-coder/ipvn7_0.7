# ipvn7 — Domótica Soberana y Dispositivos IoT (`ipvn7-bridge`)

**Versión:** 0.7.0  
**Binario:** `bin/windows_amd64/ipvn7-bridge.exe`  
**Módulos:** [`pkg/components/device_bridge/`](../pkg/components/device_bridge/), [`pkg/core/server_home.go`](../pkg/core/server_home.go)

---

## 1. El Puente del "Último Metro"
El **Sovereign Device Bridge** integra periféricos locales (Smart TVs, impresoras, cámaras) que carecen de criptografía Ed25519 nativa, asignándoles identidades virtuales seguras (*Shadow DIDs*):
```text
[LAN Doméstica: Smart TV / Impresora] ──► [ipvn7-bridge] ──► [Smart Gateway :7070] ──► [Malla ipvn7]
```

---

## 2. Identidades Sombra (*Shadow DIDs*) y ZTNA
* **Cálculo Determinista:**
  $$\text{ShadowDID} = \text{did:ipvn7:shadow:} \parallel \text{SHA-256}(\text{MAC} \parallel \text{IP})[0:32]$$
* **Aislamiento ZTNA:** Son catalogados como terminales pasivos. No pueden originar tráfico hacia la malla ni pivotar entre nodos; solo reciben flujos autorizados por el usuario.

---

## 3. Capacidades y Endpoints Integrados

### 3.1 Emisión a Smart TV (UPnP/DLNA)
```bash
curl -X POST http://localhost:7070/api/home/cast \
  -H "Content-Type: application/json" \
  -d '{
    "target_tv": "Smart TV Samsung",
    "media_url": "http://10.7.0.42:7070/media/video.mp4",
    "title": "Conferencia Soberana"
  }'
```

### 3.2 Servidor de Impresión IPP (RFC 8011) con Wake-on-LAN
Activa automáticamente la impresora con un *Magic Packet* UDP antes de enviar el documento:
```bash
curl -X POST http://localhost:7070/api/home/print \
  -H "Content-Type: application/json" \
  -d '{
    "target_printer": "HP LaserJet",
    "doc_name": "reporte.pdf",
    "copies": 1
  }'
```

---

## 4. Compilación y Ejecución

```bash
# Compilar binario satelital
go build -o bin/windows_amd64/ipvn7-bridge.exe ./cmd/ipvn7-bridge

# Ejecutar escáner con cadencia de 30 segundos
./bin/windows_amd64/ipvn7-bridge.exe -gateway http://127.0.0.1:7070 -interval 30s
```
