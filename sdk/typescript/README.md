# @ipvn7/sdk

TypeScript / JavaScript client library for the **IPVN7 Sovereign Network OS**. Compatible with Node.js, Deno, Bun, and browser environments.

## Installation
```bash
npm install @ipvn7/sdk
```

## Quick Example
```typescript
import { IPVN7Client, DatagramEnvelope } from "@ipvn7/sdk";

const client = new IPVN7Client("http://127.0.0.1:7070");

const envelope: DatagramEnvelope = {
  sourceDID: "did:ipvn7:local-agent",
  targetDID: "did:ipvn7:target-node",
  protocol: "agent:event",
  payload: new TextEncoder().encode("Hello from TypeScript!"),
};

const result = await client.sendDatagram(envelope);
console.log(`Delivered via ${result.viaPeer} in ${result.latencyMs}ms`);
```
