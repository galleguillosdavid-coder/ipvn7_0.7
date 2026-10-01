```text
Network Working Group                                         D. Galleguillos
Request for Comments: 9707                                    IPVN7 Standards
Category: Standards Track                                      September 2026
ISSN: 2070-1721
```

# IPVN7 Core Protocol Specification (IPVN7-CORE)

## Abstract
This document specifies the core protocol for IPVN7 (IP Version 7 Network Operating System), a decentralized, sovereign, post-quantum overlay network operating system. IPVN7 decouples identity from network location through asymmetric cryptographic keys, enforces deterministic 1280-byte canonical framing, and implements a zero-copy small-world routing topology.

---

## 1. Introduction and Scope

Traditional Internet architectures bind cryptographic identity to volatile topological addresses (IPv4 and IPv6). IPVN7 establishes an autonomous overlay network where nodes are identified strictly by public keys under the `did:ipvn7:` URI scheme. 

IPVN7 operates as a pure userspace network daemon, mediating communication between sovereign endpoints without reliance on centralized domain name services, certificate authorities, or hardware-tied identifiers.

---

## 2. Terminology and Conventions

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in BCP 14 [RFC2119] [RFC8174] when, and only when, they appear in all capitals, as shown here.

* **DID (Decentralized Identifier):** The canonical identifier of an IPVN7 endpoint derived from its public key.
* **Kleinberg Ring:** A logarithmic routing distance bucket bounded strictly to 120 peers across 12 concentric rings.
* **Canonical MTU:** The universal maximum transmission unit of exactly 1280 octets.
* **ZTNA (Zero Trust Network Access):** Default-deny packet filtering evaluated strictly against authenticated DIDs.

---

## 3. Protocol Architecture (Layers L0 - L2)

```text
+-------------------------------------------------------------------+
|                    L2: Lock-Free Telemetry                        |
|        Atomic Ring Buffer (<28 ns) · Flow Accounting (Tit-for-Tat) |
+-------------------------------------------------------------------+
|                 L1: Transport, Mesh & Security                    |
| Kleinberg Routing (12 Rings) · Zero-Copy Pool · ZTNA Firewall     |
| PQC 1-RTT Handshake · Sphinx Onion 3-Hop · Memory Arbiter        |
+-------------------------------------------------------------------+
|                L0: Cryptographic Foundation & Wire                |
| Ed25519 Keys · Deterministic CBOR (RFC 8949) · 1280B Canonical MTU|
| BLAKE2s SIMD Checksum · 1024-Bit Anti-Replay Sliding Window       |
+-------------------------------------------------------------------+
```

---

## 4. Sovereign Identity (L0)

### 4.1. Key Generation and DID Derivation
Each endpoint MUST generate an Edwards-curve Digital Signature Algorithm (Ed25519) keypair. The node identifier MUST follow the canonical format:
```text
did:ipvn7:<64-character-lowercase-hexadecimal-public-key>
```

### 4.2. Deterministic IP Mapping
To provide seamless interoperability with legacy applications:
1. **IPv6 Unique Local Address (ULA):** Derived by hashing the public key with BLAKE2s:
   ```text
   fd07:<digest[0..1]>:<digest[2..3]>:<digest[4..5]>:<digest[6..7]>::1/64
   ```
2. **Virtual IPv4 Address:** Deterministically mapped to the `10.7.0.0/16` block:
   ```text
   10.7.<digest[0]>.<digest[1]>/16
   ```

---

## 5. Packet Format and Wire Framing (L0)

All IPVN7 datagrams MUST be serialized using deterministic Concise Binary Object Representation (CBOR) [RFC8949].

### 5.1. Binary Header Specification
An IPVN7 packet comprises a mandatory 32-octet header followed by a variable-length payload, constrained to the 1280-octet Canonical MTU:

```text
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|          Magic (0x49503756)           |    Version    |  MsgType  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                         Sequence Number                       |
|                            (64 bits)                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                           Session ID                          |
|                            (64 bits)                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      BLAKE2s Fast Checksum                    |
|                            (64 bits)                          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Sender DID (Variable CBOR Byte String) ...                   |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Recipient DID (Variable CBOR Byte String) ...                |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| Payload (Encrypted or Raw Data, <= 1248 octets) ...          |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### 5.2. Checksum Verification
Before parsing CBOR structures, receivers MUST compute the BLAKE2s-64 checksum over the payload and compare it in constant time. Packets failing verification MUST be dropped immediately with zero heap allocation.

---

## 6. Anti-Replay Protection (L0)

To defend against replay attacks in multi-path environments with extreme jitter, endpoints MUST maintain a 1024-bit sliding window per active session:
* The window comprises sixteen 64-bit unsigned words (`[16]uint64`).
* If `seq <= highest_seq - 1024`, the packet MUST be dropped.
* If `highest_seq - 1024 < seq <= highest_seq`, the bit at offset `(highest_seq - seq)` is checked. If set, the packet MUST be dropped.
* If `seq > highest_seq`, the window is bit-shifted left by `(seq - highest_seq)` positions, and the bit for `seq` is marked.

---

## 7. Kleinberg Small-World Routing (L1)

### 7.1. Distance Metric
Routing distance between node $X$ and node $Y$ is defined as the XOR metric over their 256-bit public keys:
$$d(X, Y) = X \oplus Y$$

### 7.2. Logarithmic Ring Topology
Routing tables MUST partition known peers into 12 concentric rings based on $\lfloor \log_2(d) \rfloor$. Each ring MUST hold at most 10 peers, strictly enforcing an upper bound of 120 peers per node table.

### 7.3. Greedy Forwarding
When receiving a packet addressed to destination $D$, the node MUST:
1. Deliver locally if $D = \text{local\_DID}$.
2. Otherwise, select the peer $P$ minimizing $d(P, D)$ from the routing table.
3. Forward to $P$ via sustainable pacing without buffering packet copies in the heap.

---

## 8. Zero Trust Network Access (ZTNA) Firewall (L1)

The ZTNA firewall MUST operate under a strict **Default-Deny** model. Packets from unauthenticated or unwhitelisted DIDs MUST be dropped at the earliest pipeline stage before reaching Layer 3 or application handlers.

---

## 9. Security Considerations

1. **Quantum Adversary Protection:** Handshakes MUST employ hybrid post-quantum key encapsulation (ML-KEM-768 + X25519) in 1 RTT.
2. **Denial of Service (DoS) Immunity:** Memory allocation MUST be governed by a Global Memory Arbiter with deterministic class budgets, preventing out-of-memory crashes up to 2.8 million packets per second.
3. **Traffic Analysis Defense:** Payloads SHOULD be padded to canonical 1280-byte frames with stochastic interval jitter.

---

## 10. References
* [RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", BCP 14, RFC 2119, March 1997.
* [RFC8949] Bormann, C. and P. Hoffman, "Concise Binary Object Representation (CBOR)", RFC 8949, December 2020.
* [RFC7693] Saarinen, M-J. and J-P. Aumasson, "The BLAKE2 Cryptographic Hash and Message Authentication Code (MAC)", RFC 7693, November 2015.
* [RFC4303] Kent, S., "IP Encapsulating Security Payload (ESP)", RFC 4303, December 2005.
