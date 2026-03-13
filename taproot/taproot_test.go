package taproot

import (
	"encoding/hex"
	"testing"
)

// 测试与 tbc-lib-js taproot 模块的对齐
func TestWIFToTaprootAddress(t *testing.T) {
	// 从私钥生成 WIF 再测试完整链路 (避免依赖外部 WIF 有效性)
	seckey, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	wif, err := SeckeyToWIF(seckey)
	if err != nil {
		t.Fatalf("SeckeyToWIF: %v", err)
	}
	addr, err := WIFToTaprootAddress(wif, "bc")
	if err != nil {
		t.Fatalf("WIFToTaprootAddress failed: %v", err)
	}
	if addr == "" {
		t.Fatal("expected non-empty address")
	}
	if len(addr) < 62 || addr[:4] != "bc1p" {
		t.Errorf("unexpected address format: %s", addr)
	}
}

func TestWIFRoundtrip(t *testing.T) {
	seckey, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	wif, err := SeckeyToWIF(seckey)
	if err != nil {
		t.Fatalf("SeckeyToWIF: %v", err)
	}
	seckey2, err := WIFToSeckey(wif)
	if err != nil {
		t.Fatalf("WIFToSeckey: %v", err)
	}
	if hex.EncodeToString(seckey2) != hex.EncodeToString(seckey) {
		t.Error("WIF roundtrip seckey mismatch")
	}
}

func TestPubkeyGen(t *testing.T) {
	seckeyHex := "0000000000000000000000000000000000000000000000000000000000000001"
	seckey, _ := hex.DecodeString(seckeyHex)
	pub, err := PubkeyGen(seckey)
	if err != nil {
		t.Fatalf("PubkeyGen: %v", err)
	}
	// 众所周知：私钥 1 的压缩公钥
	expected := "0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	if hex.EncodeToString(pub) != expected {
		t.Errorf("PubkeyGen: got %s, want %s", hex.EncodeToString(pub), expected)
	}
}

func TestBech32mRoundtrip(t *testing.T) {
	data, _ := hex.DecodeString("79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")
	addr, err := Bech32mEncode("bc", data)
	if err != nil {
		t.Fatalf("Bech32mEncode: %v", err)
	}
	decoded, err := Bech32mDecode(addr, "bc")
	if err != nil {
		t.Fatalf("Bech32mDecode: %v", err)
	}
	if hex.EncodeToString(decoded) != hex.EncodeToString(data) {
		t.Errorf("Bech32m roundtrip mismatch")
	}
}

func TestTaprootAddressRoundtrip(t *testing.T) {
	seckey, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	wif, _ := SeckeyToWIF(seckey)
	addr, err := WIFToTaprootAddress(wif, "bc")
	if err != nil {
		t.Fatalf("WIFToTaprootAddress: %v", err)
	}
	pub, err := TaprootAddressToTaprootTweakPubkey(addr, "bc")
	if err != nil {
		t.Fatalf("TaprootAddressToTaprootTweakPubkey: %v", err)
	}
	addr2, err := TaprootTweakPubkeyToTaprootAddress(pub, "bc")
	if err != nil {
		t.Fatalf("TaprootTweakPubkeyToTaprootAddress: %v", err)
	}
	if addr != addr2 {
		t.Errorf("Taproot address roundtrip: got %s, want %s", addr2, addr)
	}
}

func TestTaprootToLegacyAddress(t *testing.T) {
	seckey, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")
	wif, _ := SeckeyToWIF(seckey)
	addr, err := WIFToTaprootAddress(wif, "bc")
	if err != nil {
		t.Fatalf("WIFToTaprootAddress: %v", err)
	}
	legacy, err := TaprootAddressToTaprootTweakLegacyAddress(addr, "bc")
	if err != nil {
		t.Fatalf("TaprootAddressToTaprootTweakLegacyAddress: %v", err)
	}
	if legacy == "" || legacy[0] != '1' {
		t.Errorf("expected mainnet legacy address starting with 1, got %s", legacy)
	}
}
