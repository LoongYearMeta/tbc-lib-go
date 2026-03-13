package taproot

import (
	"encoding/hex"
	"fmt"
	"testing"
)

// TestAlignmentWithJS 输出与 tbc-lib-js 对比的参考值
// 运行: go test -v -run TestAlignmentWithJS ./taproot/
// 与 taproot-alignment-check.js 输出逐项对比
func TestAlignmentWithJS(t *testing.T) {
	seckey, _ := hex.DecodeString("0000000000000000000000000000000000000000000000000000000000000001")

	fmt.Println("=== tbc-lib-go Taproot 输出 (用于与 JS 对比) ===")

	// 1. SeckeyToWIF
	wif, _ := SeckeyToWIF(seckey)
	fmt.Println("1. SeckeyToWIF(0x00...01):")
	fmt.Println("   WIF:", wif)
	seckeyBack, _ := WIFToSeckey(wif)
	fmt.Println("   WIFToSeckey roundtrip:", hex.EncodeToString(seckeyBack))

	// 2. PubkeyGen
	pub, _ := PubkeyGen(seckey)
	fmt.Println("\n2. PubkeyGen(seckey=1):")
	fmt.Println("   Pubkey:", hex.EncodeToString(pub))
	fmt.Println("   Expected: 0279be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798")

	// 3. PubkeyToTaprootTweakPubkey
	tapPub, _ := PubkeyToTaprootTweakPubkey(pub)
	fmt.Println("\n3. PubkeyToTaprootTweakPubkey:")
	fmt.Println("   TapTweakPubkey:", hex.EncodeToString(tapPub))

	// 4. SeckeyToTaprootTweakSeckey
	tapSeckey, _ := SeckeyToTaprootTweakSeckey(seckey)
	fmt.Println("\n4. SeckeyToTaprootTweakSeckey:")
	fmt.Println("   TapTweakSeckey:", hex.EncodeToString(tapSeckey))

	// 5. TaprootTweakPubkeyToTaprootAddress
	addr, _ := TaprootTweakPubkeyToTaprootAddress(tapPub, "bc")
	fmt.Println("\n5. TaprootTweakPubkeyToTaprootAddress:")
	fmt.Println("   Address:", addr)

	// 6. WIFToTaprootAddress
	addr2, _ := WIFToTaprootAddress(wif, "bc")
	fmt.Println("\n6. WIFToTaprootAddress(wif):")
	fmt.Println("   Address:", addr2)
	fmt.Println("   Same as #5:", addr == addr2)

	// 7. TaprootAddressToTaprootTweakPubkey
	decodedPub, _ := TaprootAddressToTaprootTweakPubkey(addr, "bc")
	fmt.Println("\n7. TaprootAddressToTaprootTweakPubkey:")
	fmt.Println("   Decoded:", hex.EncodeToString(decodedPub))
	fmt.Println("   Same as tapPub:", hex.EncodeToString(decodedPub) == hex.EncodeToString(tapPub))

	// 8. TaprootAddressToTaprootTweakLegacyAddress
	legacy, _ := TaprootAddressToTaprootTweakLegacyAddress(addr, "bc")
	fmt.Println("\n8. TaprootAddressToTaprootTweakLegacyAddress:")
	fmt.Println("   Legacy:", legacy)

	fmt.Println("\n=== 与 taproot-alignment-check.js 输出对比 ===")
}
