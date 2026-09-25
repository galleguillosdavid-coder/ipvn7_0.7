# IPVN7 Python SDK

SDK ligero y sin dependencias pesadas para integrar aplicaciones Python, modelos LLM y agentes de IA al Sistema Operativo de Redes Soberanas **IPVN7**.

## Instalación
```bash
pip install -e .
```

## Uso Rápido
```python
from ipvn7 import IPVN7GatewayClient, DatagramEnvelope

client = IPVN7GatewayClient("http://127.0.0.1:7070")

# Verificar estado
if client.ping():
    print("Conectado físicamente al daemon IPVN7")

# Enviar datagrama con MTU canónico
env = DatagramEnvelope(
    source_did="did:ipvn7:local",
    target_did="did:ipvn7:notebook-dvd",
    protocol="agent:query",
    payload=b"Hola desde Agente Python!"
)
client.send_datagram(env)
```
