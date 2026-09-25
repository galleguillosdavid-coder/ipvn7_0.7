"""
IPVN7 Network OS - Python SDK
Minimalist, high-performance gateway client for sovereign P2P networking.
"""

import json
from dataclasses import dataclass
from typing import Dict, Any, Optional

MAX_PACKET_SIZE = 1280

class IPVN7Error(Exception):
    """Excepción base para errores de comunicación en la malla IPVN7."""
    pass

class MTUExceededError(IPVN7Error):
    """Se lanza cuando el paquete excede el límite de 1280 bytes."""
    pass

@dataclass
class DatagramEnvelope:
    source_did: str
    target_did: str
    protocol: str
    payload: bytes

    def validate(self) -> None:
        total_len = len(self.payload) + len(self.source_did) + len(self.target_did)
        if total_len > MAX_PACKET_SIZE:
            raise MTUExceededError(
                f"El datagrama ({total_len}B) excede el MTU canónico de {MAX_PACKET_SIZE}B"
            )

class IPVN7GatewayClient:
    """Cliente para acoplar agentes y microservicios Python al nodo IPVN7 local."""

    def __init__(self, endpoint: str = "http://127.0.0.1:7070"):
        self.endpoint = endpoint.rstrip("/")

    def ping(self) -> bool:
        """Verifica conectividad física con el nodo local."""
        try:
            import urllib.request
            req = urllib.request.Request(f"{self.endpoint}/api/v1/system/status")
            with urllib.request.urlopen(req, timeout=2.0) as resp:
                return resp.status == 200
        except Exception:
            return False

    def send_datagram(self, envelope: DatagramEnvelope) -> Dict[str, Any]:
        """Envía un datagrama hacia cualquier nodo DID de la malla."""
        envelope.validate()
        import urllib.request

        payload_data = {
            "source_did": envelope.source_did,
            "target_did": envelope.target_did,
            "protocol": envelope.protocol,
            "payload": envelope.payload.hex(),
        }

        data = json.dumps(payload_data).encode("utf-8")
        req = urllib.request.Request(
            f"{self.endpoint}/api/v1/send",
            data=data,
            headers={"Content-Type": "application/json"},
            method="POST",
        )

        try:
            with urllib.request.urlopen(req, timeout=5.0) as resp:
                res_body = resp.read().decode("utf-8")
                return json.loads(res_body)
        except Exception as e:
            raise IPVN7Error(f"Fallo al transmitir datagrama en malla: {e}") from e
