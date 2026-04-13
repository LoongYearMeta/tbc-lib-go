package partialsha256

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSuffixBlockSHA256_NoPrefixMatchesStdlib(t *testing.T) {
	t.Parallel()
	for _, n := range []int{0, 1, 55, 56, 57, 64, 119, 120} {
		n := n
		t.Run(fmt.Sprintf("n_%d", n), func(t *testing.T) {
			t.Parallel()
			msg := bytes.Repeat([]byte{byte(n + 1)}, n)
			want := sha256.Sum256(msg)
			got := SuffixBlockSHA256(nil, msg, uint64(len(msg)))
			if !bytes.Equal(got, want[:]) {
				t.Fatalf("len=%d got %x want %x", n, got, want)
			}
		})
	}
}

func TestSuffixBlockSHA256_FTV2ResumeMatchesStdlib(t *testing.T) {
	t.Parallel()
	prefix := bytes.Repeat([]byte{0x11}, 1856)
	suffix := bytes.Repeat([]byte{0x22}, 28)
	full := append(append([]byte{}, prefix...), suffix...)
	phHex := CalculatePartialHash(prefix)
	ph, err := hex.DecodeString(phHex)
	if err != nil || len(ph) != 32 {
		t.Fatalf("CalculatePartialHash: %v len=%d", err, len(ph))
	}
	got := SuffixBlockSHA256(ph, suffix, uint64(len(full)))
	want := sha256.Sum256(full)
	if !bytes.Equal(got, want[:]) {
		t.Fatalf("FT v2 chain mismatch\ngot  %x\nwant %x", got, want)
	}
}
