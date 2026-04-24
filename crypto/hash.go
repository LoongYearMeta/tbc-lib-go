// Package crypto mirrors tbc-lib-js lib/crypto/ and re-exports the
// underlying libsv/go-bk primitives under a TBC-flavored API.
package crypto

import "github.com/libsv/go-bk/crypto"

// Sha256 returns the single SHA-256 of b.
func Sha256(b []byte) []byte { return crypto.Sha256(b) }

// Sha256d returns the double SHA-256 of b (Bitcoin's "hash256").
func Sha256d(b []byte) []byte { return crypto.Sha256d(b) }

// Ripemd160 returns the RIPEMD-160 of b.
func Ripemd160(b []byte) []byte { return crypto.Ripemd160(b) }

// Hash160 returns RIPEMD-160(SHA-256(b)) — Bitcoin's public-key-hash.
func Hash160(b []byte) []byte { return crypto.Hash160(b) }
