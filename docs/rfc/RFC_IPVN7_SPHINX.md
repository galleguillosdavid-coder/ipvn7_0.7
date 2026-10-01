```text
Network Working Group                                         D. Galleguillos
Request for Comments: 9708                                    IPVN7 Standards
Category: Standards Track                                      September 2026
ISSN: 2070-1721
```

# IPVN7 Onion Routing Protocol Specification (IPVN7-SPHINX)

## Abstract
This document specifies the IPVN7-SPHINX protocol, a cryptographic multi-hop onion routing mechanism for the IPVN7 Network Operating System. IPVN7-SPHINX enforces constant 1280-byte canonical frames, post-quantum hybrid key derivation, multi-layer peeling, and stochastic dummy padding to resist deep packet inspection (DPI), flow correlation, and quantum timing analysis.

---

## 1. Introduction and Threat Model

Traditional overlay networks leak metadata (packet length distribution, inter-packet arrival times, and topological intermediate hops) to autonomous systems and passive network adversaries. 

IPVN7-SPHINX addresses these vulnerabilities through three architectural guarantees:
1. **Length Invariance:** All sphinx packets are strictly fixed at 1280 octets, eliminating packet size signatures.
2. **Hop Unlinkability:** Forwarding nodes learn exclusively the immediate predecessor and successor; intermediate hops cannot discern path length or terminal endpoints.
3. **Traffic Normalization:** Stochastic dummy frames mimic active payload transmission, defeating flow-correlation timing attacks.

---

## 2. Terminology and Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in BCP 14 [RFC2119] [RFC8174].

* **Circuit:** A unidirectional, 3-hop cryptographic path through the Kleinberg mesh (Guard -> Middle -> Exit).
* **Peeling:** The process where an intermediate node strips exactly one encryption layer using its derived shared secret.
* **Dummy Packet:** An indistinguishable synthetic 1280-byte frame injected at pseudo-random intervals to normalize traffic entropy.

---

## 3. Cryptographic Key Agreement

### 3.1. Hybrid Ephemeral Exchange
For each hop $i \in \{1, 2, 3\}$, the circuit initiator MUST compute a hybrid post-quantum shared secret:
1. An ephemeral Diffie-Hellman exchange using Curve25519 ($ss_{\text{dh}, i}$).
2. An ephemeral post-quantum key encapsulation using ML-KEM-768 ($ss_{\text{pqc}, i}$).
3. A key derivation function combining both secrets via HKDF-SHA256:
   $$K_i = \text{HKDF-Expand}(\text{HKDF-Extract}(\text{salt}, ss_{\text{dh}, i} \parallel ss_{\text{pqc}, i}), \text{"ipvn7-sphinx-v1"}, 32)$$

---

## 4. Onion Packet Structure

All IPVN7-SPHINX frames MUST adhere to the canonical 1280-octet MTU:

```text
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Magic (0x49503753)           |    Version    | Flags (Dummy) |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       Circuit ID (64 bits)                    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                    Ephemeral Public Key Hop 1                 |
|                            (256 bits)                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                Hop Routing Info Header (128 octets)           |
|                (Layer-encrypted next-hop instructions)        |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      Encrypted Payload Body                   |
|           (Layered ChaCha20-Poly1305, 1104 octets)            |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

---

## 5. Packet Processing and Peeling Algorithm

When intermediate node $N$ receives an IPVN7-SPHINX frame:
1. **Checksum & Framing Validation:** $N$ validates that the frame is exactly 1280 octets. Packets with invalid length MUST be discarded with zero allocations.
2. **Replay Check:** $N$ queries its 1024-bit anti-replay window for the `Circuit ID` and packet sequence. Replays MUST be silently dropped.
3. **Layer Decryption:** Using the negotiated hop secret $K_N$, $N$ decrypts the Routing Info Header.
4. **Next-Hop Determination:**
   - If the decrypted instruction indicates an intermediate hop, $N$ pads the tail of the header with pseudo-random bytes, rewrites the next-hop destination, and forwards via the sustainable flow pacer.
   - If the instruction indicates terminal exit, $N$ delivers the unwrapped payload to the local component gateway or destination endpoint.

---

## 6. Stochastic Dummy Traffic (Anti-DPI Engine)

To neutralize passive flow correlation attacks:
1. Endpoints MUST implement a Poisson or uniform pseudo-random interval generator with boundaries $T_{\min} = 50\text{ ms}$ and $T_{\max} = 250\text{ ms}$.
2. If no legitimate payload is queued within interval $\Delta t \in [T_{\min}, T_{\max}]$, the engine MUST transmit a synthetic dummy frame.
3. Dummy frames MUST use authentic post-quantum ephemeral keys and random ciphertexts indistinguishable from real payloads under chosen-ciphertext attack (IND-CCA2).
4. The terminal exit node MUST discard dummy frames without generating transport acknowledgments or side-channel emissions.

---

## 7. Security Considerations

1. **Compromised Relays:** A single compromised hop cannot determine both the origin and terminal destination of the circuit.
2. **Zero-Copy Hot-Path:** Processing and peeling of frames MUST occur using preallocated `BufferPool` slices, guaranteeing zero heap allocations during transit.
3. **Quantum Forward Secrecy:** Interception and storage of frames by adversaries cannot be retroactively decrypted by quantum computers due to ML-KEM-768 ephemeral encapsulation.

---

## 8. References
* [RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997.
* [RFC8439] Nir, Y. and A. Langley, "ChaCha20 and Poly1305 for IETF Protocols", June 2018.
* [FIPS203] NIST, "Module-Lattice-Based Key-Encapsulation Mechanism Standard (ML-KEM)", August 2024.
