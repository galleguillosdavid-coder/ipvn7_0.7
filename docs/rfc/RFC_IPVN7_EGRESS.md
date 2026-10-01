```text
Network Working Group                                         D. Galleguillos
Request for Comments: 9709                                    IPVN7 Standards
Category: Standards Track                                      September 2026
ISSN: 2070-1721
```

# IPVN7 Sovereign Egress Protocol Specification (IPVN7-EGRESS)

## Abstract
This document specifies the sovereign egress architecture for IP Version 7 Network Operating System (IPVN7). IPVN7-EGRESS enables arbitrary mesh nodes to advertise Internet transit capabilities, provides client-side dynamic gateway scoring with anti-flapping hysteresis, mandates TCP Maximum Segment Size (MSS) clamping to 1220 octets over 1280-octet canonical frames, and ensures zero-leak DNS resolution and zero-friction fallback to local physical links.

---

## 1. Introduction and Scope

In decentralized overlay networks, participants frequently suffer from asymmetric egress conditions: hostile local ISP throttling, deep packet inspection (DPI), censorship, or suboptimal international routing. IPVN7-EGRESS formalizes a sovereign transit mechanism allowing authenticated endpoints identified by Edwards-curve Digital Signature Algorithm (Ed25519) Decentralized Identifiers (`did:ipvn7:`) to route public IPv4 and IPv6 traffic across optimal mesh gateways without reliance on centralized proxy servers or cloud aggregators.

---

## 2. Terminology and Invariants

* **Sovereign Egress Gateway:** An IPVN7 node that declares `CanExit = true` and relays decapsulated traffic to the public Internet via dual-stack dialers or OS kernel masquerading.
* **Canonical MTU Constraint:** The strict universal datagram size of 1280 octets defined in RFC 9707.
* **MSS Clamping:** Forced reduction of the TCP Maximum Segment Size option to 1220 octets (1280 octets minus IPv6/TCP/PQC wire encapsulation overhead).
* **Anti-Flapping Hysteresis:** A minimum 25% sustained quality improvement requirement prior to switching active egress gateways.

---

## 3. Protocol Wire Format and Capability Announcement

Endpoints offering transit MUST broadcast their capacity within standard periodic keepalive frames serialized as deterministic Concise Binary Object Representation (CBOR) [RFC8949]:

```text
 0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
| CanExit (bool)|              Reserved                         |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                    Bandwidth Capacity (Mbps)                  |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|       Active Sessions         |  Country Code (ISO 3166-1)    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      Timestamp (64 bits)                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

---

## 4. Dynamic Gateway Scoring Algorithm

Clients MUST compute a weighted composite score $S$ for every active peer advertising transit:

$$S = (RTT_{ms} \times 0.40) + (LossRate_{\%} \times 0.40) + (LoadRatio \times 0.20)$$

Where:
* $RTT_{ms}$ is the round-trip latency measured via high-frequency probers.
* $LossRate_{\%}$ is the moving packet loss ratio scaled to $[0, 100]$.
* $LoadRatio$ is proportional to active outbound sessions ($\min(100, \text{sessions} \times 2)$).
* **Lower scores denote superior candidate gateways.**

### 4.1. Hysteresis and Stability Rule
To prevent route oscillation (*flapping*), a client MUST NOT switch from current gateway $G_{active}$ to candidate $G_{cand}$ unless:

$$S(G_{cand}) \le S(G_{active}) \times (1.0 - 0.25)$$

And the condition has persisted for at least 3.0 seconds, unless $G_{active}$ has entered `HealthStateUnreachable` or `HealthStateQuarantined`, in which case failover MUST occur in $<500$ milliseconds.

---

## 5. Gateway Forwarding and Zero-Trust Security (ZTNA)

1. **Default-Deny Access Control:** Gateways MUST verify the sender's Ed25519 DID. Unauthenticated requests MUST be dropped immediately with zero heap allocation.
2. **MSS Clamping:** All TCP SYN packets relayed through the gateway MUST have their MSS option clamped to $\le 1220$ octets to eliminate path MTU fragmentation.
3. **Zero-Leak DNS:** Client DNS queries MUST be encapsulated inside the post-quantum encrypted tunnel toward the gateway resolver, preventing DNS inspection by the client's local ISP.

---

## 6. References
* [RFC2119] Bradner, S., "Key words for use in RFCs to Indicate Requirement Levels", March 1997.
* [RFC8949] Bormann, C., "Concise Binary Object Representation (CBOR)", December 2020.
* [RFC9707] Galleguillos, D., "IPVN7 Core Protocol Specification", September 2026.
