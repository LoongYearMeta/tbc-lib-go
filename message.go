package tbc

// Message 实现（参考 tbc-lib-js 的 lib/message/message.js）
//
// 特点：
// - 使用 "Bitcoin Signed Message:\n" 作为前缀；
// - varint 编码前缀长度与消息长度；
// - 对拼接后的数据做双 SHA256，得到 32 字节 magic hash；
// - 使用 secp256k1 做 ECDSA 签名（紧凑签名 + 恢复公钥）；
// - 签名以 base64 字符串形式返回 / 校验。

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/libsv/go-bk/base58"
	bkcrypto "github.com/libsv/go-bk/crypto"

	"github.com/LoongYearMeta/tbc-lib-go/script"
	"github.com/LoongYearMeta/tbc-lib-go/encoding"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpECDSA "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// messageMagicBytes 与 JS 端 Message.MAGIC_BYTES 对齐
var messageMagicBytes = []byte("Bitcoin Signed Message:\n")

// Message 封装待签名/验证的消息
type Message struct {
	Data []byte
	// Error 对齐 JS 侧 this.error 的语义，用于保存最近一次校验失败原因。
	Error string
}

// NewMessageFromString 从字符串构造 Message
func NewMessageFromString(s string) *Message {
	return &Message{Data: []byte(s)}
}

// NewMessageFromBytes 从字节数组构造 Message
func NewMessageFromBytes(b []byte) *Message {
	cp := make([]byte, len(b))
	copy(cp, b)
	return &Message{Data: cp}
}

// FromString 等价于 NewMessageFromString，与 JS 的 Message.fromString 对齐
func FromString(s string) *Message {
	return NewMessageFromString(s)
}

// ToJSON 将消息序列化为 JSON（与 JS 的 toJSON 对齐）
func (m *Message) ToJSON() (string, error) {
	if m == nil {
		return "", fmt.Errorf("message: nil")
	}
	obj := map[string]string{
		"messageHex": hex.EncodeToString(m.Data),
	}
	data, err := json.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("message: json marshal failed: %w", err)
	}
	return string(data), nil
}

// FromJSON 从 JSON 反序列化消息（与 JS 的 fromJSON 对齐）
func FromJSON(jsonStr string) (*Message, error) {
	var obj struct {
		MessageHex string `json:"messageHex"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &obj); err != nil {
		return nil, fmt.Errorf("message: json unmarshal failed: %w", err)
	}
	data, err := hex.DecodeString(obj.MessageHex)
	if err != nil {
		return nil, fmt.Errorf("message: invalid hex: %w", err)
	}
	return NewMessageFromBytes(data), nil
}

// ToObject 返回消息的普通对象表示（与 JS 的 toObject 对齐）
func (m *Message) ToObject() map[string]string {
	if m == nil {
		return nil
	}
	return map[string]string{
		"messageHex": hex.EncodeToString(m.Data),
	}
}

// FromObject 从对象构造消息（与 JS 的 fromObject 对齐）
func FromObject(obj map[string]string) (*Message, error) {
	hexStr, ok := obj["messageHex"]
	if !ok {
		return nil, fmt.Errorf("message: missing messageHex field")
	}
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("message: invalid hex: %w", err)
	}
	return NewMessageFromBytes(data), nil
}

// String 返回消息的字符串表示
func (m *Message) String() string {
	if m == nil {
		return ""
	}
	return string(m.Data)
}

// MagicHash 计算 "Bitcoin Signed Message" 风格的双 SHA256 哈希
func (m *Message) MagicHash() []byte {
	if m == nil {
		return nil
	}
	prefix1 := encoding.VarInt(uint64(len(messageMagicBytes))).Bytes()
	prefix2 := encoding.VarInt(uint64(len(m.Data))).Bytes()

	buf := make([]byte, 0, len(prefix1)+len(messageMagicBytes)+len(prefix2)+len(m.Data))
	buf = append(buf, prefix1...)
	buf = append(buf, messageMagicBytes...)
	buf = append(buf, prefix2...)
	buf = append(buf, m.Data...)

	h1 := sha256.Sum256(buf)
	h2 := sha256.Sum256(h1[:])
	return h2[:]
}

// Sign 使用 secp256k1 私钥对消息进行签名，返回 base64 编码的紧凑签名
// 注意：当前实现返回标准签名（64字节），如需完整 compact 签名（65字节含 recovery_id），
// 请使用 SignMessage 函数，它会自动计算 recovery_id
func (m *Message) Sign(priv *secp.PrivateKey) (string, error) {
	return SignMessage(m.String(), priv)
}

// Verify 使用给定公钥验证签名（紧凑签名 + base64）
//
// 支持两种签名格式：
// - 64 字节：标准紧凑签名（r + s）
// - 65 字节：带 recovery_id 的紧凑签名（recovery_id + r + s）
func (m *Message) Verify(pub *secp.PublicKey, sigBase64 string) (bool, error) {
	m.Error = ""
	if pub == nil {
		m.Error = "First argument should be an instance of PublicKey"
	}
	return VerifyMessageWithPubKey(m.String(), pub, sigBase64)
}

// VerifyWithAddress 使用地址验证签名（与 JS 的 Message.verify 对齐）
// 支持从 compact 签名恢复公钥并验证地址
func (m *Message) VerifyWithAddress(address, sigBase64, network string) (bool, error) {
	m.Error = ""
	ok, err := VerifyMessageWithAddress(m.String(), address, sigBase64, network)
	if err != nil {
		m.Error = err.Error()
		return false, err
	}
	if !ok {
		m.Error = "The signature did not match the message digest"
	}
	return ok, nil
}

// ========== 静态函数：高层 API（与 JS 对齐） ==========

// SignMessage 签名消息（等价 JS 的 Message.sign(message, privateKey)）
func SignMessage(message string, priv *secp.PrivateKey) (string, error) {
	if priv == nil {
		return "", fmt.Errorf("message: private key is nil")
	}
	msg := NewMessageFromString(message)
	hash := msg.MagicHash()
	if len(hash) != 32 {
		return "", fmt.Errorf("message: invalid hash length %d", len(hash))
	}

	// 与 JS ECDSA.signWithCalcI 对齐：直接输出 compact 签名（含 recovery id）
	compact := secpECDSA.SignCompact(priv, hash, true)
	if len(compact) != 65 {
		return "", fmt.Errorf("message: invalid compact signature length %d", len(compact))
	}
	sigB64 := base64.StdEncoding.EncodeToString(compact)
	return sigB64, nil
}

// VerifyMessage 等价 JS 的 Message.verify(message, address, signature)。
func VerifyMessage(message, address, sigBase64 string) (bool, error) {
	return VerifyMessageWithAddress(message, address, sigBase64, "")
}

// VerifyMessageWithPubKey 使用公钥验证消息签名
func VerifyMessageWithPubKey(message string, pub *secp.PublicKey, sigBase64 string) (bool, error) {
	if pub == nil {
		return false, fmt.Errorf("message: public key is nil")
	}
	if sigBase64 == "" {
		return false, fmt.Errorf("message: empty signature")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return false, fmt.Errorf("message: invalid base64 signature: %w", err)
	}

	msg := NewMessageFromString(message)
	hash := msg.MagicHash()
	if len(hash) != 32 {
		return false, fmt.Errorf("message: invalid hash length %d", len(hash))
	}

	// 处理两种签名格式：65 字节 compact 或 64 字节 r||s
	if len(sigBytes) == 65 {
		recovered, _, err := secpECDSA.RecoverCompact(sigBytes, hash)
		if err != nil {
			return false, nil
		}
		return publicKeysEqual(recovered, pub), nil
	}
	if len(sigBytes) == 64 {
		for header := byte(27); header <= byte(34); header++ {
			compact := make([]byte, 65)
			compact[0] = header
			copy(compact[1:], sigBytes)
			recovered, _, err := secpECDSA.RecoverCompact(compact, hash)
			if err != nil {
				continue
			}
			if publicKeysEqual(recovered, pub) {
				return true, nil
			}
		}
		return false, nil
	}
	return false, fmt.Errorf("message: invalid signature length %d (expected 64 or 65)", len(sigBytes))
}

// VerifyMessageWithAddress 使用地址验证消息签名（等价 JS 的 Message.verify）
// 从 compact 签名恢复公钥，然后验证地址是否匹配
func VerifyMessageWithAddress(message, address, sigBase64, network string) (bool, error) {
	if address == "" {
		return false, fmt.Errorf("message: empty address")
	}
	if sigBase64 == "" {
		return false, fmt.Errorf("message: empty signature")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sigBase64)
	if err != nil {
		return false, fmt.Errorf("message: invalid base64 signature: %w", err)
	}

	if len(sigBytes) != 65 {
		return false, fmt.Errorf("message: compact signature must be 65 bytes, got %d", len(sigBytes))
	}

	msg := NewMessageFromString(message)
	hash := msg.MagicHash()
	if len(hash) != 32 {
		return false, fmt.Errorf("message: invalid hash length %d", len(hash))
	}

	// recover 公钥，行为与 JS 里的 ecdsa.toPublicKey 对齐
	recoveredPub, wasCompressed, err := secpECDSA.RecoverCompact(sigBytes, hash)
	if err != nil {
		return false, fmt.Errorf("message: recover public key failed: %w", err)
	}

	// 以传入地址所属网络为准（与 JS Address.fromPublicKey(publicKey, bitcoinAddress.network) 一致）
	addrNetMainnet, err := inferMainnetFromAddress(address, network)
	if err != nil {
		return false, err
	}
	recoveredAddr, err := addressFromRecoveredKey(recoveredPub, wasCompressed, addrNetMainnet)
	if err != nil {
		return false, fmt.Errorf("message: create address from public key failed: %w", err)
	}

	// 验证地址是否匹配
	if recoveredAddr.AddressString != address {
		return false, nil
	}
	return true, nil
}

func addressFromRecoveredKey(pub *secp.PublicKey, compressed bool, mainnet bool) (*script.Address, error) {
	var serialized []byte
	if compressed {
		serialized = pub.SerializeCompressed()
	} else {
		serialized = pub.SerializeUncompressed()
	}
	return script.NewAddressFromPublicKeyHash(bkcrypto.Hash160(serialized), mainnet)
}

func publicKeysEqual(a, b *secp.PublicKey) bool {
	if a == nil || b == nil {
		return false
	}
	return bytes.Equal(a.SerializeCompressed(), b.SerializeCompressed())
}

func inferMainnetFromAddress(address string, networkHint string) (bool, error) {
	decoded := base58.Decode(address)
	if len(decoded) == 25 {
		switch decoded[0] {
		case 0x00, 0x05:
			return true, nil
		case 0x6f, 0xc4:
			return false, nil
		default:
			return false, fmt.Errorf("message: unsupported address version 0x%x", decoded[0])
		}
	}

	// 兜底：若地址格式无法识别，退回 network hint 判断
	switch networkHint {
	case "mainnet", "livenet", "":
		return true, nil
	case "testnet", "regtest", "stn":
		return false, nil
	default:
		return false, fmt.Errorf("message: invalid address and network hint")
	}
}

// Inspect 返回便于调试的字符串（对齐 JS 的 inspect）。
func (m *Message) Inspect() string {
	if m == nil {
		return "<Message: >"
	}
	return "<Message: " + m.String() + ">"
}
