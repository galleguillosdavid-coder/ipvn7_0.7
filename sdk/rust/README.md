# IPVN7 Client SDK for Rust

Sovereign client bindings in Rust for the **IPVN7 Network OS Smart Component Gateway** (`/api/v1/*`).

## Installation

Add to your `Cargo.toml`:
```toml
[dependencies]
ipvn7-sdk = { path = "../sdk/rust" }
```

## Features
* Pure Safe Rust, zero unneeded dependencies.
* Built-in validation for the canonical 1280-byte MTU.
* Serde-compatible data structures for all Gateway endpoints.
