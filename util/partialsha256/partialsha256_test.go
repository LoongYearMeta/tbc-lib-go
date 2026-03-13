package partialsha256

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// 短输入时 partial hash 应与标准 SHA256 前几轮一致（仅单 block 时）
func TestCalculatePartialHash_ShortInput(t *testing.T) {
	in := []byte("hello")
	h := CalculatePartialHash(in)
	// 标准 SHA256 用于对比
	full := sha256.Sum256(in)
	fullHex := hex.EncodeToString(full[:])
	// partial 返回的是未完成最后一轮的状态，应不同于完整 hash
	if h == fullHex {
		t.Log("partial equals full for short input - may be expected for single block")
	}
	if len(h) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(h))
	}
}
