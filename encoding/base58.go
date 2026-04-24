package encoding

import "github.com/libsv/go-bk/base58"

// Base58Encode returns the modified-base58 (Bitcoin alphabet) encoding of b.
func Base58Encode(b []byte) string { return base58.Encode(b) }

// Base58Decode decodes a modified-base58 string. Returns an empty slice when
// s contains characters outside the Bitcoin alphabet (matching the underlying
// libsv/go-bk/base58.Decode behaviour, which does not surface an error).
func Base58Decode(s string) []byte { return base58.Decode(s) }
