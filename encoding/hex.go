package encoding

import "encoding/hex"

// EncodeHex returns b as lowercase hex.
func EncodeHex(b []byte) string { return hex.EncodeToString(b) }

// DecodeHex parses s as hex. Returns an error if s is malformed.
func DecodeHex(s string) ([]byte, error) { return hex.DecodeString(s) }

// IsHex reports whether s is a valid, even-length hex string.
func IsHex(s string) bool {
	if len(s)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
