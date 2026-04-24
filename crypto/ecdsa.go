package crypto

import "github.com/libsv/go-bk/bec"

// PrivateKey is a secp256k1 private key.
type PrivateKey = bec.PrivateKey

// PublicKey is a secp256k1 public key.
type PublicKey = bec.PublicKey

// ParsePubKey parses a compressed or uncompressed public key.
var ParsePubKey = bec.ParsePubKey

// ParseDERSignature parses a DER-encoded ECDSA signature.
var ParseDERSignature = bec.ParseDERSignature

// NewPrivateKey generates a new random private key on secp256k1.
var NewPrivateKey = bec.NewPrivateKey

// S256 returns the secp256k1 curve.
var S256 = bec.S256
