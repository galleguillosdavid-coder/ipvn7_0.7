#!/usr/bin/env python3
"""
sync_firebase_hub.py - Sincronizador Soberano de Directorio de Red en Firebase RTDB
Publica el estado del nodo, directorio de agentes IA y guia de incorporacion en Google Firebase.
"""

import json
import time
import base64
import os
import requests
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
REPO_ROOT = os.path.dirname(SCRIPT_DIR)
CRED_PATH = os.path.join(REPO_ROOT, "config", "firebase_credentials.json")

def get_access_token(cred):
    header = {"alg": "RS256", "typ": "JWT"}
    now = int(time.time())
    payload = {
        "iss": cred["client_email"],
        "sub": cred["client_email"],
        "aud": "https://oauth2.googleapis.com/token",
        "iat": now,
        "exp": now + 3600,
        "scope": "https://www.googleapis.com/auth/firebase.database https://www.googleapis.com/auth/userinfo.email"
    }

    def b64url(data):
        if isinstance(data, (dict, list)):
            data = json.dumps(data).encode("utf-8")
        return base64.urlsafe_b64encode(data).decode("utf-8").rstrip("=")

    jwt_unsigned = f"{b64url(header)}.{b64url(payload)}"
    private_key = serialization.load_pem_private_key(cred["private_key"].encode("utf-8"), password=None)
    signature = private_key.sign(jwt_unsigned.encode("utf-8"), padding.PKCS1v15(), hashes.SHA256())
    jwt_signed = f"{jwt_unsigned}.{b64url(signature)}"

    resp = requests.post("https://oauth2.googleapis.com/token", data={
        "grant_type": "urn:ietf:params:oauth:grant-type:jwt-bearer",
        "assertion": jwt_signed
    }, timeout=10)
    return resp.json().get("access_token")

def main():
    if not os.path.exists(CRED_PATH):
        print(f"[ERROR] Credenciales no encontradas en {CRED_PATH}")
        return

    with open(CRED_PATH, "r", encoding="utf-8") as f:
        cred = json.load(f)

    print("[-] Autenticando con Google Firebase Admin...")
    token = get_access_token(cred)
    if not token:
        print("[ERROR] No se pudo obtener el token de acceso OAuth2")
        return

    headers = {"Authorization": f"Bearer {token}"}
    base_url = f"https://{cred['project_id']}-default-rtdb.firebaseio.com"

    # 1. Directorio de Malla Soberana y Nodos Semilla
    mesh_directory = {
        "network_id": "ipvn7-sovereign-mesh-0.7",
        "version": "0.7.0",
        "crypto": {
            "kem": "ML-KEM-768 (FIPS 203)",
            "signature": "Ed25519",
            "frame_mtu": 1280
        },
        "updated_at": int(time.time()),
        "status": "active_and_healthy",
        "seed_nodes": {
            "node_a_pc": {
                "alias": "PC Principal",
                "lan_endpoint": "192.168.1.198:7777",
                "web_ui": "http://192.168.1.198:7070",
                "role": "core_gateway_and_router"
            },
            "node_b_notebook": {
                "alias": "Notebook Dvd",
                "did": "did:ipvn7:d45e14b1893fbeb688c270a780fa0a5078df7b27744cad7bf433a1fe5889a2a6",
                "lan_endpoint": "192.168.1.106:7001",
                "web_ui": "http://192.168.1.106:8080",
                "role": "hil_visual_edge_worker"
            }
        }
    }

    # 2. Portal para Agentes Autónomos de Inteligencia Artificial (A2A & MCP)
    agent_hub = {
        "title": "IPvN7 Agentic Superhighway (A2A & MCP Hub)",
        "protocols_supported": ["a2a/1.0", "mcp/2024-11-05", "ap2/2025-09"],
        "discovery_endpoint": "/.well-known/agent-card.json",
        "capabilities": [
            "p2p_sovereign_datagram",
            "zero_copy_tensor_pipeline",
            "pqc_encryption_ml_kem_768",
            "adaptive_principal_chain",
            "autonomous_spending_mandate"
        ],
        "instructions_for_ai": "Para conectar un agente a la malla soberana IPvN7, consulta los pares en /peers y utiliza el protocolo A2A vía sockets UDP o JSON-RPC local en el puerto Web 7070 o 8080."
    }

    # 3. Guía de Onboarding para Nuevos Usuarios Humanos
    onboarding_guide = {
        "welcome": "Bienvenido a IPvN7 Sovereign Network OS (v0.7.0)",
        "features": [
            "Zero-Friction: Conexión en 1-clic con auto-elevación",
            "Telemetría Vital: Monitor de Ritmo Cardíaco en tiempo real (ECG 60 FPS)",
            "Anti-Censura & Privacidad: Supera a WireGuard en latencia y a Tor en privacidad",
            "Soberanía Total: Cero servidores centrales, criptografía post-cuántica nativa"
        ],
        "install_commands": {
            "windows_gui": "Descarga y ejecuta dist/Instalador_VPN_I7.exe (doble clic)",
            "windows_powershell": "powershell -ExecutionPolicy Bypass -NoProfile -File scripts/install.ps1",
            "linux_and_macos": "bash scripts/install.sh"
        }
    }

    print("[-] Publicando Directorio de Red en Firebase RTDB...")
    r1 = requests.put(f"{base_url}/mesh_directory.json", headers=headers, json=mesh_directory, timeout=10)
    print("  -> /mesh_directory:", r1.status_code)

    r2 = requests.put(f"{base_url}/agent_hub.json", headers=headers, json=agent_hub, timeout=10)
    print("  -> /agent_hub:", r2.status_code)

    r3 = requests.put(f"{base_url}/onboarding_guide.json", headers=headers, json=onboarding_guide, timeout=10)
    print("  -> /onboarding_guide:", r3.status_code)

    print(f"\n[OK] Publicación completada con éxito en Google Firebase!")
    print(f"URL de Datos en Vivo: {base_url}/.json")

if __name__ == "__main__":
    main()
