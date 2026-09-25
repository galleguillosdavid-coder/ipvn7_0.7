# ipvn7 — Manual Operativo CLI (`ipvn7-cli`)

**Versión:** 0.7.0  
**Binario:** `bin/windows_amd64/ipvn7-cli.exe` / `bin/linux_amd64/ipvn7-cli`

---

## 1. Comandos del Núcleo Universal

```bash
# Estado de identidad DID, IPs virtuales y telemetría
ipvn7-cli status

# Lista de pares en los 12 anillos de Kleinberg
ipvn7-cli peers

# Componentes satelitales acoplados al Smart Gateway
ipvn7-cli components

# Máquina de Estados Finitos (FSM) de salud de enlaces (Healthy, Degraded...)
ipvn7-cli fsm

# Visualización ASCII de anillos concéntricos
ipvn7-cli radar

# Telemetría del marcapasos Packet Pacing (MTU 1280B)
ipvn7-cli pace

# Iniciar servidor Model Context Protocol sobre stdio
ipvn7-cli mcp
```

---

## 2. Comandos de Innovación y Malla

```bash
# Cortafuegos ZTNA Default-Deny
ipvn7-cli firewall list
ipvn7-cli firewall allow <did>
ipvn7-cli firewall test <did>

# Almacén inmutable DAG y DTN Store-Forward
ipvn7-cli dag put "Datos inmutables de prueba"
ipvn7-cli dag status

# Contabilidad de reciprocidad Tit-for-Tat
ipvn7-cli accounting

# Red de confianza Web-of-Trust (WoT)
ipvn7-cli wot vouch <did> 85 "Nodo estable de enrutamiento"
ipvn7-cli wot score <did>

# Resolución y nombres contextuales dDNS Petnames
ipvn7-cli alias asignar "notebook" did:ipvn7:<pubkey>
ipvn7-cli ddns resolve "notebook"
ipvn7-cli ddns export-hosts

# Planificador Multi-Camino
ipvn7-cli multipath list
ipvn7-cli multipath strategy <aggregate|priority|failover>

# Copiloto de IA y auto-curación
ipvn7-cli copilot diagnose
ipvn7-cli copilot stats
```

---

## 3. Comandos de Gobernanza e Inmunología

```bash
# Constitución Digital Soberana
ipvn7-cli constitution view
ipvn7-cli constitution verify <archivo.go>

# Senado de Agentes y Democracia Líquida
ipvn7-cli senate list
ipvn7-cli senate propose --title="Optimización WDRR" --cat=routing
ipvn7-cli senate vote <prop_id> --approve
ipvn7-cli senate veto <prop_id>
ipvn7-cli senate report

# Centinelas e Inmunología Celular
ipvn7-cli sentinel status
ipvn7-cli sentinel audit <did>
ipvn7-cli sentinel alerts
```
