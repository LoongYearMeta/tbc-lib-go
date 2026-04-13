package partialsha256

import (
	"encoding/binary"
	"encoding/hex"
)

// CalculatePartialHash 计算脚本的 partial SHA256
// 对应 JS partial_sha256.calculate_partial_hash(script)
// 在最后一轮压缩前返回 HASH 状态
func CalculatePartialHash(script []byte) string {
	h := getPartialSha256(script)
	return hashToHex(h)
}

func getPartialSha256(m []byte) [8]uint32 {
	length := uint64(len(m)) * 8
	words := bufferToWords(m)
	// Padding (匹配 JS)
	wordIdx := int(length >> 5)
	if wordIdx >= len(words) {
		words = append(words, make([]uint32, wordIdx+1-len(words))...)
	}
	words[wordIdx] |= 0x80 << (24 - length%32)
	padLen := ((int(length+64) >> 9) << 4) + 16
	for len(words) < padLen {
		words = append(words, 0)
	}
	words[padLen-1] = uint32(length)

	hash := [8]uint32{
		0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a,
		0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19,
	}
	var w [64]uint32
	lastBlockStart := len(words) - 16
	for i := 0; i < len(words); i += 16 {
		for j := 0; j < 16; j++ {
			w[j] = words[i+j]
		}
		for j := 16; j < 64; j++ {
			w[j] = gamma1(w[j-2]) + w[j-7] + gamma0(w[j-15]) + w[j-16]
		}
		a, b, c, d, e, f, g, hh := hash[0], hash[1], hash[2], hash[3], hash[4], hash[5], hash[6], hash[7]
		for j := 0; j < 64; j++ {
			t1 := hh + sigma1(e) + ch(e, f, g) + k[j] + w[j]
			t2 := sigma0(a) + maj(a, b, c)
			hh, g, f, e = g, f, e, d+t1
			d, c, b, a = c, b, a, t1+t2
		}
		if i == lastBlockStart {
			// JS partial_sha256 返回的是“最后一个 block 之前的 HASH 状态”，
			// 即此处 hash[] 还没做过 hash += (a..hh) 的那一轮累计结果。
			return hash
		}
		hash[0] += a
		hash[1] += b
		hash[2] += c
		hash[3] += d
		hash[4] += e
		hash[5] += f
		hash[6] += g
		hash[7] += hh
	}
	return hash
}

func bufferToWords(b []byte) []uint32 {
	n := (len(b) + 3) / 4
	words := make([]uint32, n)
	for i := 0; i < len(b)*8; i += 8 {
		words[i>>5] |= uint32(b[i/8]&0xff) << (24 - i%32)
	}
	return words
}

func ch(x, y, z uint32) uint32 { return (x & y) ^ (^x & z) }
func maj(x, y, z uint32) uint32 { return (x & y) ^ (x & z) ^ (y & z) }
func sigma0(x uint32) uint32 { return rotr(x, 2) ^ rotr(x, 13) ^ rotr(x, 22) }
func sigma1(x uint32) uint32 { return rotr(x, 6) ^ rotr(x, 11) ^ rotr(x, 25) }
func gamma0(x uint32) uint32 { return rotr(x, 7) ^ rotr(x, 18) ^ (x >> 3) }
func gamma1(x uint32) uint32 { return rotr(x, 17) ^ rotr(x, 19) ^ (x >> 10) }
func rotr(x uint32, n uint) uint32 { return (x >> n) | (x << (32 - n)) }

var k = [64]uint32{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
}

func hashToHex(h [8]uint32) string {
	buf := make([]byte, 32)
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(buf[i*4:], h[i])
	}
	return hex.EncodeToString(buf)
}
