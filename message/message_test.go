package message

import (
	"encoding/base64"
	"testing"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	bkcrypto "github.com/LoongYearMeta/tbc-lib-go/crypto"
	"github.com/LoongYearMeta/tbc-lib-go/script"
)

func TestMessage_SignAndVerify(t *testing.T) {
	// 生成测试私钥
	priv, err := secp.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}

	message := "Hello, TBC!"
	msg := NewMessageFromString(message)

	// 签名
	sigBase64, err := msg.Sign(priv)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	if sigBase64 == "" {
		t.Fatal("signature is empty")
	}

	// 使用公钥验证
	pub := priv.PubKey()
	ok, err := msg.Verify(pub, sigBase64)
	if err != nil {
		t.Fatalf("Verify failed: %v", err)
	}
	if !ok {
		t.Fatal("signature verification failed")
	}
}

func TestSignMessage(t *testing.T) {
	// 生成测试私钥
	priv, err := secp.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}

	message := "Test message"
	sigBase64, err := SignMessage(message, priv)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}
	if sigBase64 == "" {
		t.Fatal("signature is empty")
	}

	// 验证
	pub := priv.PubKey()
	ok, err := VerifyMessageWithPubKey(message, pub, sigBase64)
	if err != nil {
		t.Fatalf("VerifyMessageWithPubKey failed: %v", err)
	}
	if !ok {
		t.Fatal("signature verification failed")
	}

	// 兼容 64 字节签名输入（不带头字节）
	raw, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		t.Fatalf("decode base64 signature failed: %v", err)
	}
	if len(raw) != 65 {
		t.Fatalf("expected compact signature 65 bytes, got %d", len(raw))
	}
	sig64 := base64.StdEncoding.EncodeToString(raw[1:])
	ok, err = VerifyMessageWithPubKey(message, pub, sig64)
	if err != nil {
		t.Fatalf("VerifyMessageWithPubKey(64) failed: %v", err)
	}
	if !ok {
		t.Fatal("64-byte signature verification failed")
	}
}

func TestVerifyMessageWithAddress(t *testing.T) {
	priv, err := secp.GeneratePrivateKey()
	if err != nil {
		t.Fatalf("GeneratePrivateKey failed: %v", err)
	}
	message := "verify with address"

	sigBase64, err := SignMessage(message, priv)
	if err != nil {
		t.Fatalf("SignMessage failed: %v", err)
	}

	pub := priv.PubKey()
	addr, err := script.NewAddressFromPublicKeyHash(bkcrypto.Hash160(pub.SerializeCompressed()), true)
	if err != nil {
		t.Fatalf("NewAddressFromPublicKeyHash failed: %v", err)
	}

	ok, err := VerifyMessageWithAddress(message, addr.AddressString, sigBase64, "mainnet")
	if err != nil {
		t.Fatalf("VerifyMessageWithAddress failed: %v", err)
	}
	if !ok {
		t.Fatal("VerifyMessageWithAddress should be true")
	}

	ok, err = VerifyMessageWithAddress(message, "1BoatSLRHtKNngkdXEeobR76b53LETtpyT", sigBase64, "mainnet")
	if err != nil {
		t.Fatalf("VerifyMessageWithAddress(wrong address) failed: %v", err)
	}
	if ok {
		t.Fatal("VerifyMessageWithAddress should be false for mismatched address")
	}
}

func TestMessage_ToJSON(t *testing.T) {
	message := "Test JSON"
	msg := NewMessageFromString(message)

	jsonStr, err := msg.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}
	if jsonStr == "" {
		t.Fatal("JSON string is empty")
	}

	// 反序列化
	msg2, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if msg2.String() != message {
		t.Errorf("message mismatch: expected %q, got %q", message, msg2.String())
	}
}
