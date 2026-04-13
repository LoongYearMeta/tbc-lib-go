package partialsha256

import (
	"encoding/binary"
)

// SuffixBlockSHA256 completes SHA256(prefix || suffix) when prefix was absorbed with
// the same state as calculate_partial_hash(prefix) / getPartialSha256(prefix).
// Matches tbc-lib-js lib/util/partial-sha256.js binb_partial_sha256 (OP_PARTIAL_HASH).
//
// partialState: 32 bytes big-endian SHA256 midstate, or empty for full IV.
// suffix: remaining bytes after the split (e.g. lockScript[1856:] for FT v2).
// completeMessageByteLen: total message length in bytes (prefix+suffix), e.g. 1884 for FT v2 code.
func SuffixBlockSHA256(partialState []byte, suffix []byte, completeMessageByteLen uint64) []byte {
	completeSizeBits := completeMessageByteLen * 8
	remainLengthBits := uint64(len(suffix)) * 8

	var hashState [8]uint32
	if len(partialState) == 32 {
		for i := 0; i < 8; i++ {
			hashState[i] = binary.BigEndian.Uint32(partialState[i*4 : (i+1)*4])
		}
	} else {
		hashState = [8]uint32{
			0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
			0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
		}
	}

	padBlocks := int((remainLengthBits + 65 + 511) / 512)
	requiredWords := padBlocks * 16
	msgWords := make([]uint32, requiredWords)

	for i := 0; i < len(suffix); i++ {
		wordIdx := i / 4
		byteIdx := i % 4
		msgWords[wordIdx] |= uint32(suffix[i]) << (24 - byteIdx*8)
	}

	paddingWordIdx := int(remainLengthBits / 32)
	paddingBitPos := int(remainLengthBits % 32)
	msgWords[paddingWordIdx] |= 0x80 << (24 - paddingBitPos)

	lengthWordIdx := int(((remainLengthBits+64)>>9)<<4) + 15
	if lengthWordIdx >= len(msgWords) {
		nw := ((lengthWordIdx + 1 + 15) / 16) * 16
		extended := make([]uint32, nw)
		copy(extended, msgWords)
		msgWords = extended
	}
	msgWords[lengthWordIdx] = uint32(completeSizeBits)

	for i := 0; i < len(msgWords); i += 16 {
		var w [64]uint32
		for j := 0; j < 16 && i+j < len(msgWords); j++ {
			w[j] = msgWords[i+j]
		}
		for j := 16; j < 64; j++ {
			s0 := rotr32(w[j-15], 7) ^ rotr32(w[j-15], 18) ^ (w[j-15] >> 3)
			s1 := rotr32(w[j-2], 17) ^ rotr32(w[j-2], 19) ^ (w[j-2] >> 10)
			w[j] = w[j-16] + s0 + w[j-7] + s1
		}

		a, b, c, d, e, f, g, hh := hashState[0], hashState[1], hashState[2], hashState[3],
			hashState[4], hashState[5], hashState[6], hashState[7]

		for j := 0; j < 64; j++ {
			S1 := rotr32(e, 6) ^ rotr32(e, 11) ^ rotr32(e, 25)
			ch := (e & f) ^ (^e & g)
			t1 := hh + S1 + ch + k[j] + w[j]
			S0 := rotr32(a, 2) ^ rotr32(a, 13) ^ rotr32(a, 22)
			maj := (a & b) ^ (a & c) ^ (b & c)
			t2 := S0 + maj

			hh, g, f, e = g, f, e, d+t1
			d, c, b, a = c, b, a, t1+t2
		}

		hashState[0] += a
		hashState[1] += b
		hashState[2] += c
		hashState[3] += d
		hashState[4] += e
		hashState[5] += f
		hashState[6] += g
		hashState[7] += hh
	}

	out := make([]byte, 32)
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(out[i*4:(i+1)*4], hashState[i])
	}
	return out
}

func rotr32(value uint32, bits uint) uint32 {
	return (value >> bits) | (value << (32 - bits))
}
