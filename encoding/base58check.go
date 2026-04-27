package encoding

import "github.com/LoongYearMeta/tbc-lib-go/base58"

// Base58CheckEncode prepends the single-byte version prefix and appends a
// 4-byte double-SHA256 checksum, returning the modified-base58 encoding of
// (version || payload || checksum).
func Base58CheckEncode(version byte, payload []byte) string {
	return base58.CheckEncode(payload, version)
}

// Base58CheckDecode parses a Base58Check string. It returns the version byte
// and the payload (without version/checksum). An error is returned if the
// string is malformed or the checksum does not match.
func Base58CheckDecode(s string) (version byte, payload []byte, err error) {
	result, ver, err := base58.CheckDecode(s)
	if err != nil {
		return 0, nil, err
	}
	return ver, result, nil
}
