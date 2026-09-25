//! # IPVN7 Client SDK for Rust
//!
//! Provides a sovereign, zero-overhead client for interfacing with the
//! IPVN7 Network OS Smart Component Gateway (`/api/v1/*`).

use serde::{Deserialize, Serialize};

/// Canonical MTU enforced by IPVN7 Network OS.
pub const CANONICAL_MTU: usize = 1280;

/// Standard Gateway API endpoints.
pub const EP_STATUS: &str = "/api/v1/status";
pub const EP_PEERS: &str = "/api/v1/peers";
pub const EP_SEND: &str = "/api/v1/send";
pub const EP_COMPONENTS: &str = "/api/v1/components";

/// Node identity and telemetry snapshot.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct NodeStatus {
    pub did: String,
    pub sovereign_ipv6: String,
    pub virtual_ipv4: String,
    pub peers_count: usize,
    pub uptime_seconds: f64,
}

/// Peer summary in Kleinberg 12-ring topology.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct PeerSummary {
    pub did: String,
    pub ring: u8,
    pub latency_ms: f64,
    pub physical_addr: Option<String>,
}

/// Outbound datagram injection request.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct SendRequest {
    pub target_did: String,
    pub protocol: String,
    pub priority: u8,
    pub payload: Vec<u8>,
}

/// Response returned by the gateway after datagram dispatch.
#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct SendResponse {
    pub status: String,
    pub sequence: u64,
    pub bytes_written: usize,
}

/// Client handle for IPVN7 Smart Component Gateway.
#[derive(Debug, Clone)]
pub struct Ipvn7Client {
    gateway_url: String,
}

impl Ipvn7Client {
    /// Instantiates a new client pointing to the specified Gateway base URL.
    pub fn new(gateway_url: impl Into<String>) -> Self {
        let mut url = gateway_url.into();
        if url.ends_with('/') {
            url.pop();
        }
        Self { gateway_url: url }
    }

    /// Returns the full endpoint URL for node status.
    pub fn status_url(&self) -> String {
        format!("{}{}", self.gateway_url, EP_STATUS)
    }

    /// Returns the full endpoint URL for peer discovery.
    pub fn peers_url(&self) -> String {
        format!("{}{}", self.gateway_url, EP_PEERS)
    }

    /// Returns the full endpoint URL for datagram injection.
    pub fn send_url(&self) -> String {
        format!("{}{}", self.gateway_url, EP_SEND)
    }

    /// Returns the full endpoint URL for external component registry.
    pub fn components_url(&self) -> String {
        format!("{}{}", self.gateway_url, EP_COMPONENTS)
    }

    /// Validates datagram constraints before transmission.
    pub fn validate_payload(payload: &[u8]) -> Result<(), &'static str> {
        if payload.is_empty() {
            return Err("Payload cannot be empty");
        }
        if payload.len() > CANONICAL_MTU {
            return Err("Payload exceeds IPVN7 canonical MTU (1280 octets)");
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_client_endpoints() {
        let client = Ipvn7Client::new("http://127.0.0.1:7070/");
        assert_eq!(client.status_url(), "http://127.0.0.1:7070/api/v1/status");
        assert_eq!(client.peers_url(), "http://127.0.0.1:7070/api/v1/peers");
        assert_eq!(client.send_url(), "http://127.0.0.1:7070/api/v1/send");
    }

    #[test]
    fn test_payload_validation() {
        assert!(Ipvn7Client::validate_payload(&[]).is_err());
        assert!(Ipvn7Client::validate_payload(&[1, 2, 3]).is_ok());

        let oversized = vec![0u8; CANONICAL_MTU + 1];
        assert!(Ipvn7Client::validate_payload(&oversized).is_err());
    }

    #[test]
    fn test_serde_serialization() {
        let req = SendRequest {
            target_did: "did:ipvn7:0123456789abcdef".into(),
            protocol: "mesh:data".into(),
            priority: 1,
            payload: vec![10, 20, 30],
        };
        let json_str = serde_json::to_string(&req).expect("Failed to serialize");
        let decoded: SendRequest = serde_json::from_str(&json_str).expect("Failed to deserialize");
        assert_eq!(req, decoded);
    }
}
