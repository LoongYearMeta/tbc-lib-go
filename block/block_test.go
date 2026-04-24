package block

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/LoongYearMeta/tbc-lib-go/transaction"
)

func TestBlockHeader_RoundTrip(t *testing.T) {
	prevHash, _ := hex.DecodeString("000000000000000000021adf2d73f6fddcf8f0ad880b42f5f7a3136ce4b0f24c")
	merkleRoot, _ := hex.DecodeString("3a3bd3f5f0f8b5a39ec6f133def96f2992d9dd9dc828f6dd6f095f72f9e5f4e1")

	header := &BlockHeader{
		Version:    1,
		PrevHash:   prevHash,
		MerkleRoot: merkleRoot,
		Time:       1231469665,
		Bits:       GenesisBits,
		Nonce:      2083236893,
	}

	encoded := header.Bytes()
	decoded, err := NewBlockHeaderFromBytes(encoded)
	if err != nil {
		t.Fatalf("decode header failed: %v", err)
	}

	if header.Version != decoded.Version {
		t.Fatalf("version mismatch: got %d want %d", decoded.Version, header.Version)
	}
	if hex.EncodeToString(decoded.PrevHash) != hex.EncodeToString(prevHash) {
		t.Fatalf("prevhash mismatch")
	}
	if hex.EncodeToString(decoded.MerkleRoot) != hex.EncodeToString(merkleRoot) {
		t.Fatalf("merkleroot mismatch")
	}
	if decoded.ID() == "" || len(decoded.ID()) != 64 {
		t.Fatalf("invalid header id: %q", decoded.ID())
	}
	if !decoded.ValidTimestamp(time.Unix(int64(decoded.Time), 0)) {
		t.Fatal("timestamp should be valid when checked against itself")
	}
}

func TestBlock_MerkleAndRoundTrip(t *testing.T) {
	tx1, err := transaction.NewTxFromString("02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c0000000049483045022100c4df63202a9aa2bea5c24ebf4418d145e81712072ef744a4b108174f1ef59218022006eb54cf904707b51625f521f8ed2226f7d34b62492ebe4ddcb1c639caf16c3c41ffffffff0140420f00000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000")
	if err != nil {
		t.Fatalf("decode tx1 failed: %v", err)
	}
	tx2, err := transaction.NewTxFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
	if err != nil {
		t.Fatalf("decode tx2 failed: %v", err)
	}

	block := &Block{
		Header: &BlockHeader{
			Version: 1,
			Time:    1231469665,
			Bits:    GenesisBits,
			Nonce:   1,
		},
		Transactions: []*transaction.Tx{tx1, tx2},
	}
	block.Header.MerkleRoot = block.GetMerkleRoot()

	if !block.ValidMerkleRoot() {
		t.Fatal("expected valid merkle root")
	}

	encoded := block.Bytes()
	decoded, err := NewBlockFromBytes(encoded)
	if err != nil {
		t.Fatalf("decode block failed: %v", err)
	}
	if len(decoded.Transactions) != 2 {
		t.Fatalf("tx count mismatch: got %d", len(decoded.Transactions))
	}
	if !decoded.ValidMerkleRoot() {
		t.Fatal("decoded block should keep valid merkle root")
	}
	if decoded.ID() == "" {
		t.Fatal("decoded block id should not be empty")
	}
}
