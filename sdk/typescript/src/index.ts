/**
 * IPVN7 Network OS - TypeScript SDK
 * High-performance client library for Node.js and Browser/WASM environments.
 */

export const MAX_PACKET_SIZE = 1280;

export interface DatagramEnvelope {
  sourceDID: string;
  targetDID: string;
  protocol: string;
  payload: Uint8Array;
}

export interface SendResult {
  delivered: boolean;
  latencyMs: number;
  viaPeer?: string;
  error?: string;
}

export class IPVN7Client {
  private endpoint: string;

  constructor(endpoint: string = "http://127.0.0.1:7070") {
    this.endpoint = endpoint.replace(/\/+$/, "");
  }

  /**
   * Valida la longitud total del datagrama conforme al MTU canónico de 1280 bytes
   */
  public validateEnvelope(env: DatagramEnvelope): void {
    const totalBytes = env.payload.byteLength + env.sourceDID.length + env.targetDID.length;
    if (totalBytes > MAX_PACKET_SIZE) {
      throw new Error(`Datagram size (${totalBytes}B) exceeds canonical MTU of ${MAX_PACKET_SIZE}B`);
    }
  }

  /**
   * Envía un datagrama hacia la malla a través de la API Gateway local
   */
  public async sendDatagram(env: DatagramEnvelope): Promise<SendResult> {
    this.validateEnvelope(env);

    // Convertir Uint8Array a cadena hexadecimal
    const hexPayload = Array.from(env.payload)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");

    const response = await fetch(`${this.endpoint}/api/v1/send`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        source_did: env.sourceDID,
        target_did: env.targetDID,
        protocol: env.protocol,
        payload: hexPayload,
      }),
    });

    if (!response.ok) {
      throw new Error(`IPVN7 Gateway HTTP error: ${response.statusText}`);
    }

    return (await response.json()) as SendResult;
  }
}
