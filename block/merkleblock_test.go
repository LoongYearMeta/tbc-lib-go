package block

import (
	"encoding/hex"
	"testing"

	"github.com/LoongYearMeta/tbc-lib-go/crypto"

	"github.com/LoongYearMeta/tbc-lib-go/transaction"
)

func TestMerkleBlock_ValidateAndFilter(t *testing.T) {
	tx1, err := transaction.NewTxFromString("02000000011ccba787d421b98904da3329b2c7336f368b62e89bc896019b5eadaa28145b9c0000000049483045022100c4df63202a9aa2bea5c24ebf4418d145e81712072ef744a4b108174f1ef59218022006eb54cf904707b51625f521f8ed2226f7d34b62492ebe4ddcb1c639caf16c3c41ffffffff0140420f00000000001976a91418392a59fc1f76ad6a3c7ffcea20cfcb17bda9eb88ac00000000")
	if err != nil {
		t.Fatalf("decode tx1 failed: %v", err)
	}
	tx2, err := transaction.NewTxFromString("010000000193a35408b6068499e0d5abd799d3e827d9bfe70c9b75ebe209c91d2507232651000000006b483045022100c1d77036dc6cd1f3fa1214b0688391ab7f7a16cd31ea4e5a1f7a415ef167df820220751aced6d24649fa235132f1e6969e163b9400f80043a72879237dab4a1190ad412103b8b40a84123121d260f5c109bc5a46ec819c2e4002e5ba08638783bfb4e01435ffffffff02404b4c00000000001976a91404ff367be719efa79d76e4416ffb072cd53b208888acde94a905000000001976a91404d03f746652cfcb6cb55119ab473a045137d26588ac00000000")
	if err != nil {
		t.Fatalf("decode tx2 failed: %v", err)
	}

	h1 := crypto.Sha256d(tx1.Bytes())
	h2 := crypto.Sha256d(tx2.Bytes())
	root := crypto.Sha256d(append(append([]byte{}, h1...), h2...))

	mb := &MerkleBlock{
		Header: &BlockHeader{
			Version:    1,
			MerkleRoot: root,
			Time:       1231469665,
			Bits:       GenesisBits,
			Nonce:      1,
		},
		NumTransactions: 2,
		Hashes: []string{
			hex.EncodeToString(h1),
			hex.EncodeToString(h2),
		},
		Flags: []byte{0x03}, // root(1) + left leaf match(1) + right leaf no match(0)
	}

	if !mb.ValidMerkleTree() {
		t.Fatal("expected valid merkle tree")
	}

	txs, err := mb.FilteredTxsHash()
	if err != nil {
		t.Fatalf("filtered hashes failed: %v", err)
	}
	if len(txs) != 1 || txs[0] != hex.EncodeToString(h1) {
		t.Fatalf("unexpected filtered tx hashes: %v", txs)
	}

	if !mb.HasTransaction(hex.EncodeToString(h1)) {
		t.Fatal("expected has transaction by hash")
	}
	if !mb.HasTransaction(tx1) {
		t.Fatal("expected has transaction by tx object")
	}

	encoded := mb.Bytes()
	decoded, err := NewMerkleBlockFromBytes(encoded)
	if err != nil {
		t.Fatalf("decode merkle block failed: %v", err)
	}
	if !decoded.ValidMerkleTree() {
		t.Fatal("decoded merkle block should be valid")
	}
}
