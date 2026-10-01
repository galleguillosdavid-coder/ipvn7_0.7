// Package interfaces define los contratos raíz universales e inmutables del protocolo IPVN7 (v0.7).
package interfaces

import "net"

// IdentityAuthority representa la entidad criptográfica soberana del nodo local o remoto.
// En IPVN7 la identidad es inmutable, criptográficamente demostrable y no depende de autoridades
// centrales ni certificados de terceros (Sovereign Decentralized Identity).
type IdentityAuthority interface {
	// DID retorna el Identificador Descentralizado soberano (ej: 'did:ipvn7:pubkey_hex').
	// Constituye la dirección global inmutable de capa L0.
	DID() string

	// UIN retorna el Universal Identity Number canónico derivado de la clave pública maestra.
	UIN() string

	// PublicKeyBytes retorna la representación binaria cruda de la clave pública Ed25519 (32 bytes).
	PublicKeyBytes() []byte

	// IPv6 retorna la dirección IPv6 estática generada criptográficamente (prefijo fc00::/7 o fd00::/8).
	IPv6() net.IP

	// IPv4 retorna la dirección IPv4 mapeada en el rango privado virtual (10.127.x.y).
	IPv4() net.IP

	// Sign firma criptográficamente un mensaje o digest utilizando la clave privada soberana.
	Sign(message []byte) []byte

	// Verify valida la autenticidad de una firma contra la clave pública del nodo.
	Verify(message, signature []byte) bool
}

// SovereignCryptographer expone las operaciones criptográficas híbridas clásicas y post-cuánticas.
// Combina curvas elípticas de alta velocidad (Ed25519, X25519) con encapsulación post-cuántica (Kyber/ML-KEM).
type SovereignCryptographer interface {
	// DeriveSharedSecret calcula un secreto compartido Diffie-Hellman efímero o Noise XX.
	DeriveSharedSecret(peerPublicKey []byte) ([]byte, error)

	// EncapsulatePQC genera un texto cifrado y un secreto compartido cuántico-resistente para el par.
	EncapsulatePQC(peerPQCPublicKey []byte) (ciphertext, sharedSecret []byte, err error)

	// DecapsulatePQC recupera el secreto compartido cuántico-resistente a partir del texto cifrado.
	DecapsulatePQC(ciphertext []byte) (sharedSecret []byte, err error)
}

// CanonicalPacket representa la estructura unificada de un datagrama IPVN7 serializado en el cable.
type CanonicalPacket interface {
	// HeaderMagic retorna el número mágico de protocolo IPVN7 (4 bytes).
	HeaderMagic() uint32

	// ProtocolVersion retorna la versión de especificación del protocolo (ej: 0x07).
	ProtocolVersion() uint8

	// PacketType retorna la categoría de mensaje (ej: Data, Ping, RouteDiscover, RoamingUpdate, Handshake).
	PacketType() uint8

	// GetSourceDID retorna el DID del nodo emisor soberano verificado.
	GetSourceDID() string

	// GetDestinationDID retorna el DID del nodo destinatario final o broadcast de anillo.
	GetDestinationDID() string

	// GetSequence retorna el número secuencial monotónico para control de flujo y anti-replay.
	GetSequence() uint64

	// GetPayload retorna los datos transportados en claro o cifrados mediante Noise AEAD.
	GetPayload() []byte
}

// WireCodec abstrae la serialización y deserialización binaria canónica (ej: CBOR v2).
type WireCodec interface {
	// EncodePacket serializa la estructura del paquete a su representación binaria canónica.
	EncodePacket(pkt CanonicalPacket) ([]byte, error)

	// DecodePacket reconstruye un CanonicalPacket desde un slice de bytes binarios.
	DecodePacket(raw []byte) (CanonicalPacket, error)
}
