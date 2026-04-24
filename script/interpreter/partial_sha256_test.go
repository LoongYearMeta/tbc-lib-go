package interpreter

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/LoongYearMeta/tbc-lib-go/util/partialsha256"
)

func TestPartialSHA256_MatchesStdlibSHA256_NoPrefix(t *testing.T) {
	t.Parallel()
	cases := [][]byte{
		[]byte("a"),
		bytes.Repeat([]byte{0xab}, 55),
		bytes.Repeat([]byte{0xcd}, 56),
		bytes.Repeat([]byte{0xef}, 57),
		bytes.Repeat([]byte{0x01}, 64),
		bytes.Repeat([]byte{0x02}, 119),
		bytes.Repeat([]byte{0x03}, 120),
	}
	for _, m := range cases {
		m := m
		t.Run(fmt.Sprintf("len_%d", len(m)), func(t *testing.T) {
			t.Parallel()
			l := uint64(len(m))
			got := partialsha256.SuffixBlockSHA256(nil, m, l)
			want := sha256.Sum256(m)
			if !bytes.Equal(got, want[:]) {
				t.Fatalf("len(m)=%d: got %x want %x", len(m), got, want)
			}
		})
	}
}

func TestPartialSHA256_FTV2ResumeMatchesFullSHA256(t *testing.T) {
	t.Parallel()
	prefix := bytes.Repeat([]byte{0x11}, 1856)
	suffix := bytes.Repeat([]byte{0x22}, 28)
	full := append(append([]byte{}, prefix...), suffix...)
	phHex := partialsha256.CalculatePartialHash(prefix)
	ph, err := hex.DecodeString(phHex)
	if err != nil || len(ph) != 32 {
		t.Fatalf("CalculatePartialHash: %v len=%d", err, len(ph))
	}
	got := partialsha256.SuffixBlockSHA256(ph, suffix, uint64(len(full)))
	want := sha256.Sum256(full)
	if !bytes.Equal(got, want[:]) {
		t.Fatalf("partialSHA256(suffix, partial(prefix), %d) != SHA256(prefix||suffix)\ngot  %x\nwant %x", len(full), got, want)
	}
}
