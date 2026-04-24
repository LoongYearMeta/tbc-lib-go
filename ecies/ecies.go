package ecies

// ECIES 实现（参考 tbc-lib-js 的 electrum-ecies.js）
//
// 说明：
// - 采用 secp256k1 椭圆曲线做 ECDH，派生共享密钥；
// - 使用 AES-128-CBC + PKCS7 填充进行对称加密；
// - 使用 HMAC-SHA256 对完整密文进行完整性校验；
// - 消息格式（BIE1 变体）：
//     BIE1 | [Rbuf?] | ciphertext | HMAC
//   其中：
//     - BIE1: 4 字节 ASCII 常量；
//     - Rbuf: 发送方临时公钥（可选，压缩格式）；
//     - ciphertext: AES-128-CBC 密文；
//     - HMAC: 32 字节（或 4 字节 shortTag）。

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// ECIESOptions 与 JS 侧的 opts 对应
type ECIESOptions struct {
	// EphemeralKey 为 true 时，如果未显式设置私钥，将随机生成临时私钥
	EphemeralKey bool
	// NoKey 为 true 时，密文中不包含发送方公钥（Rbuf）
	NoKey bool
	// ShortTag 为 true 时，HMAC 截断为前 4 字节
	ShortTag bool
	// Algorithm 当前仅支持 "BIE1"
	Algorithm string
}

// 默认选项：BIE1 + 临时私钥 + 包含公钥 + 完整 HMAC
var defaultECIESOptions = ECIESOptions{
	EphemeralKey: true,
	NoKey:        false,
	ShortTag:     false,
	Algorithm:    "BIE1",
}

// ECIES 实例
type ECIES struct {
	priv *secp.PrivateKey
	pub  *secp.PublicKey
	opts ECIESOptions
}

// NewECIES 创建一个 ECIES 实例
func NewECIES(opts *ECIESOptions) *ECIES {
	o := defaultECIESOptions
	if opts != nil {
		if opts.Algorithm != "" {
			o.Algorithm = opts.Algorithm
		}
		if opts.EphemeralKey {
			o.EphemeralKey = true
		}
		if opts.NoKey {
			o.NoKey = true
		}
		if opts.ShortTag {
			o.ShortTag = true
		}
	}
	return &ECIES{opts: o}
}

// PrivateKey 设置发送方私钥（如果设置，则 EphemeralKey 自动视为 false）
func (e *ECIES) PrivateKey(priv *secp.PrivateKey) *ECIES {
	if priv == nil {
		return e
	}
	e.priv = priv
	e.opts.EphemeralKey = false
	return e
}

// PublicKey 设置接收方公钥
func (e *ECIES) PublicKey(pub *secp.PublicKey) *ECIES {
	if pub == nil {
		return e
	}
	e.pub = pub
	return e
}

// Rbuf 返回发送方公钥（压缩格式）
func (e *ECIES) Rbuf() ([]byte, error) {
	if e.priv == nil {
		return nil, errors.New("ecies: private key not set")
	}
	return e.priv.PubKey().SerializeCompressed(), nil
}

// ivkEkM: SHA512(Sbuf)
// 这里为了与 JS BIE1 接近，采用压缩公钥作为共享“秘密”的输入。
func (e *ECIES) ivkEkM() ([]byte, error) {
	if e.priv == nil {
		return nil, errors.New("ecies: private key not set")
	}
	if e.pub == nil {
		return nil, errors.New("ecies: public key not set")
	}

	// ECDH：直接使用库提供的共享密钥接口，避免依赖内部点类型。
	sbuf := secp.GenerateSharedSecret(e.priv, e.pub)

	sum := sha512.Sum512(sbuf)
	return sum[:], nil
}

func (e *ECIES) iv() ([]byte, error) {
	ivk, err := e.ivkEkM()
	if err != nil {
		return nil, err
	}
	return ivk[0:16], nil
}

func (e *ECIES) kE() ([]byte, error) {
	ivk, err := e.ivkEkM()
	if err != nil {
		return nil, err
	}
	return ivk[16:32], nil
}

func (e *ECIES) kM() ([]byte, error) {
	ivk, err := e.ivkEkM()
	if err != nil {
		return nil, err
	}
	return ivk[32:64], nil
}

// EncryptBIE1 按 BIE1 规格加密
func (e *ECIES) EncryptBIE1(message []byte) ([]byte, error) {
	if len(message) == 0 {
		return nil, errors.New("ecies: empty message")
	}
	if e.pub == nil {
		return nil, errors.New("ecies: public key required")
	}

	// 如未设置私钥且允许 EphemeralKey，则生成随机私钥
	if e.priv == nil {
		if !e.opts.EphemeralKey {
			return nil, errors.New("ecies: private key not set and EphemeralKey is false")
		}
		priv, err := secp.GeneratePrivateKey()
		if err != nil {
			return nil, fmt.Errorf("ecies: generate ephemeral key failed: %w", err)
		}
		e.priv = priv
	}

	iv, err := e.iv()
	if err != nil {
		return nil, err
	}
	kE, err := e.kE()
	if err != nil {
		return nil, err
	}
	kM, err := e.kM()
	if err != nil {
		return nil, err
	}

	ciphertext, err := aesCBCEncrypt(message, kE, iv)
	if err != nil {
		return nil, err
	}

	const magic = "BIE1"
	var encbuf []byte

	if e.opts.NoKey && !e.opts.EphemeralKey {
		encbuf = append([]byte(magic), ciphertext...)
	} else {
		rbuf, err := e.Rbuf()
		if err != nil {
			return nil, err
		}
		encbuf = append(append([]byte(magic), rbuf...), ciphertext...)
	}

	hmacTag := hmacSHA256(encbuf, kM)
	if e.opts.ShortTag {
		hmacTag = hmacTag[:4]
	}

	return append(encbuf, hmacTag...), nil
}

// DecryptBIE1 解密 BIE1 消息
func (e *ECIES) DecryptBIE1(enc []byte) ([]byte, error) {
	if len(enc) < 4+16+4 { // magic + 至少 16 字节密文 + HMAC(截断)
		return nil, errors.New("ecies: ciphertext too short")
	}
	if e.priv == nil {
		return nil, errors.New("ecies: private key required")
	}

	const magic = "BIE1"
	if !bytes.Equal(enc[:4], []byte(magic)) {
		return nil, errors.New("ecies: invalid magic header")
	}

	offset := 4
	tagLen := 32
	if e.opts.ShortTag {
		tagLen = 4
	}
	if len(enc) < offset+tagLen+16 {
		return nil, errors.New("ecies: invalid ciphertext length")
	}

	// 如果密文里包含 Rbuf，则从中恢复对方公钥
	if !e.opts.NoKey {
		// BIE1 使用压缩公钥，固定 33 字节
		if len(enc) < offset+33+tagLen {
			return nil, errors.New("ecies: invalid ciphertext (missing Rbuf)")
		}
		rbuf := enc[offset : offset+33]
		pub, err := secp.ParsePubKey(rbuf)
		if err != nil {
			return nil, fmt.Errorf("ecies: parse Rbuf failed: %w", err)
		}
		e.pub = pub
		offset += 33
	}

	if len(enc) < offset+tagLen+16 {
		return nil, errors.New("ecies: invalid ciphertext (too short after header)")
	}

	ciphertext := enc[offset : len(enc)-tagLen]
	tag := enc[len(enc)-tagLen:]

	iv, err := e.iv()
	if err != nil {
		return nil, err
	}
	kE, err := e.kE()
	if err != nil {
		return nil, err
	}
	kM, err := e.kM()
	if err != nil {
		return nil, err
	}

	expectedTag := hmacSHA256(enc[:len(enc)-tagLen], kM)
	if e.opts.ShortTag {
		expectedTag = expectedTag[:4]
	}
	if !hmac.Equal(tag, expectedTag) {
		return nil, errors.New("ecies: hmac mismatch")
	}

	plain, err := aesCBCDecrypt(ciphertext, kE, iv)
	if err != nil {
		return nil, err
	}

	return plain, nil
}

// EncryptFor 是一个简单封装：使用临时私钥给指定公钥加密
func EncryptFor(pub *secp.PublicKey, msg []byte, opts *ECIESOptions) ([]byte, error) {
	ec := NewECIES(opts).PublicKey(pub)
	return ec.EncryptBIE1(msg)
}

// DecryptWith 使用给定私钥解密 BIE1 消息
func DecryptWith(priv *secp.PrivateKey, enc []byte, opts *ECIESOptions) ([]byte, error) {
	ec := NewECIES(opts).PrivateKey(priv)
	return ec.DecryptBIE1(enc)
}

// --- 辅助函数：AES-CBC + PKCS7 填充 ---

func aesCBCEncrypt(plain, key, iv []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, fmt.Errorf("ecies: AES-128 key length must be 16 bytes, got %d", len(key))
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("ecies: IV length must be %d bytes, got %d", aes.BlockSize, len(iv))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	padded := pkcs7Pad(plain, aes.BlockSize)
	out := make([]byte, len(padded))

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(out, padded)

	return out, nil
}

func aesCBCDecrypt(ciphertext, key, iv []byte) ([]byte, error) {
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ecies: ciphertext is not a multiple of block size")
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("ecies: AES-128 key length must be 16 bytes, got %d", len(key))
	}
	if len(iv) != aes.BlockSize {
		return nil, fmt.Errorf("ecies: IV length must be %d bytes, got %d", aes.BlockSize, len(iv))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	out := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(out, ciphertext)

	plain, err := pkcs7Unpad(out, aes.BlockSize)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

func pkcs7Pad(data []byte, blockSize int) []byte {
	if blockSize <= 0 {
		panic("ecies: invalid block size")
	}
	padLen := blockSize - (len(data) % blockSize)
	if padLen == 0 {
		padLen = blockSize
	}
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}

func pkcs7Unpad(data []byte, blockSize int) ([]byte, error) {
	if len(data) == 0 || len(data)%blockSize != 0 {
		return nil, errors.New("ecies: invalid padded data")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > blockSize || padLen > len(data) {
		return nil, errors.New("ecies: invalid pkcs7 padding")
	}
	for i := 0; i < padLen; i++ {
		if data[len(data)-1-i] != byte(padLen) {
			return nil, errors.New("ecies: invalid pkcs7 padding content")
		}
	}
	return data[:len(data)-padLen], nil
}

func hmacSHA256(msg, key []byte) []byte {
	h := hmac.New(sha256.New, key)
	// sha256.Hash.Write 实现中不会返回错误，仅为实现 io.Writer 接口
	if _, err := h.Write(msg); err != nil {
		panic("hmac: sha256 write failed: " + err.Error())
	}
	return h.Sum(nil)
}

// RandomBytes 返回指定长度的安全随机字节
func RandomBytes(n int) ([]byte, error) {
	if n <= 0 {
		return nil, errors.New("ecies: invalid random length")
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, fmt.Errorf("ecies: random read failed: %w", err)
	}
	return b, nil
}
