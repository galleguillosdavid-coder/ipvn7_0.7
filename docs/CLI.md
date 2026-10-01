# ipvn7 — Manual Operativo CLI (`ipvn7-cli`)

**Versión:** 0.7.0  
**Binario:** `bin/windows_amd64/ipvn7-cli.exe` / `bin/linux_amd64/ipvn7-cli`

---

## 0. Banderas del Demonio Principal (`ipvn7.exe`)

```bash
# Ejecución estándar (intenta auto-elevación a Administrador; si se cancela, pasa a Modo Usuario)
ipvn7.exe

# Forzar Modo Usuario sin solicitar elevación de Administrador
ipvn7.exe -no-elevate

# Forzar adaptador de red kernel Wintun L3 (TCP + UDP)
ipvn7.exe -tun

# Ejecución con depuración detallada y volcado de eventos
ipvn7.exe -debug

# Especificar ruta personalizada de archivo de registro persistente
ipvn7.exe -logfile ./data/ipvn7.log

# Seguir el registro de ejecución en vivo en tiempo real (PowerShell)
powershell -File scripts/view_logs.ps1 -Follow
```

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

---

## 4. Compilación, Distribución y Blindaje de Ejecutables (Roles K y N)

```bash
# Compilación universal multiplataforma estándar (Windows, Linux, macOS)
powershell -File scripts/build_all_platforms.ps1

# Compilación blindada y ofuscada local (Rol N - Garble AST + cifrado de strings + stripping DWARF)
powershell -File scripts/build_hardened.ps1

# Compilación blindada para todas las plataformas (Windows, Linux, macOS)
powershell -File scripts/build_hardened.ps1 -AllTargets

# Compilación blindada en entornos Unix / Bash
bash scripts/build_hardened.sh linux amd64
```
