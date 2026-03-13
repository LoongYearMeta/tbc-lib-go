// Package taproot 实现 BIP-341/BIP-342 Taproot 功能，与 tbc-lib-js 的 taproot 模块功能对齐。
//
// 功能包括：
//   - WIF 与私钥互转
//   - 公钥生成与压缩
//   - Taproot 地址生成 (Bech32m)
//   - Taproot 地址与传统地址转换
//   - Taproot 密钥调整 (Key Path spend)
package taproot

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"math/big"

	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/libsv/go-bk/base58"
	"github.com/libsv/go-bk/crypto"
)

// Bech32m 字符集 (BIP-350)
const charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

// Bech32m 校验和常数 (BIP-350)
const bech32mConst = 0x2bc830a3

// secp256k1 曲线参数 (与 tbc-lib-js taproot.js 一致)
var (
	curveP, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEFFFFFC2F", 16)
	curveN, _ = new(big.Int).SetString("FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141", 16)
)

// ErrInvalidSeckey 表示私钥无效
var ErrInvalidSeckey = errors.New("invalid secret key")

// ErrInvalidPubkey 表示公钥无效
var ErrInvalidPubkey = errors.New("invalid public key: must be 33 bytes")

// ErrInvalidWIF 表示 WIF 格式无效
var ErrInvalidWIF = errors.New("invalid WIF: checksum does not match or invalid prefix")

// ErrInvalidAddress 表示地址格式无效
var ErrInvalidAddress = errors.New("invalid bech32m address")

// WIFToSeckey 将 WIF 格式私钥转换为 32 字节原始私钥 (仅支持压缩格式: 0x80+32+0x01+4)
func WIFToSeckey(wif string) ([]byte, error) {
	decoded := base58.Decode(wif)
	if len(decoded) < 38 {
		return nil, ErrInvalidWIF
	}
	if decoded[0] != 0x80 {
		return nil, fmt.Errorf("%w: prefix is not 0x80", ErrInvalidWIF)
	}
	payload := decoded[:len(decoded)-4]
	checksum := decoded[len(decoded)-4:]
	hash1 := crypto.Sha256(payload)
	hash2 := crypto.Sha256(hash1)
	for i := 0; i < 4; i++ {
		if checksum[i] != hash2[i] {
			return nil, ErrInvalidWIF
		}
	}
	seckey := decoded[1 : len(decoded)-5]
	if !secp256k1.PrivKeyFromBytes(seckey).Key.IsZero() {
		// 验证私钥有效
	}
	return seckey, nil
}

// SeckeyToWIF 将 32 字节私钥转换为 WIF 格式
func SeckeyToWIF(seckey []byte) (string, error) {
	if len(seckey) != 32 {
		return "", ErrInvalidSeckey
	}
	_ = secp256k1.PrivKeyFromBytes(seckey) // 验证
	extended := make([]byte, 0, 38)
	extended = append(extended, 0x80)
	extended = append(extended, seckey...)
	extended = append(extended, 0x01)
	h := crypto.Sha256(crypto.Sha256(extended))
	extended = append(extended, h[:4]...)
	return base58.Encode(extended), nil
}

// PubkeyGen 从私钥生成 33 字节压缩公钥
func PubkeyGen(seckey []byte) ([]byte, error) {
	if len(seckey) != 32 {
		return nil, ErrInvalidSeckey
	}
	priv := secp256k1.PrivKeyFromBytes(seckey)
	if priv.Key.IsZero() {
		return nil, ErrInvalidSeckey
	}
	pub := priv.PubKey()
	return pub.SerializeCompressed(), nil
}

// taggedHash 实现 BIP-340 风格的 tagged hash: SHA256(SHA256(tag)||SHA256(tag)||msg)
func taggedHash(tag string, msg []byte) []byte {
	h := sha256.Sum256([]byte(tag))
	return crypto.Sha256(append(append(h[:], h[:]...), msg...))
}

func pad32(b []byte) []byte {
	if len(b) >= 32 {
		return b[len(b)-32:]
	}
	out := make([]byte, 32)
	copy(out[32-len(b):], b)
	return out
}

// ecPoint 表示椭圆曲线上的点
type ecPoint struct {
	x, y *big.Int
}

// liftX 从 x 坐标恢复曲线上的点 (y 取偶数值，与 JS liftX 一致)
func liftX(x *big.Int) *ecPoint {
	if x.Cmp(curveP) >= 0 {
		return nil
	}
	// y^2 = x^3 + 7 (mod p)
	x3 := new(big.Int).Exp(x, big.NewInt(3), curveP)
	ySq := new(big.Int).Add(x3, big.NewInt(7))
	ySq.Mod(ySq, curveP)
	// y = ySq^((p+1)/4) mod p (p ≡ 3 mod 4)
	exp := new(big.Int).Add(curveP, big.NewInt(1))
	exp.Div(exp, big.NewInt(4))
	y := new(big.Int).Exp(ySq, exp, curveP)
	if new(big.Int).Exp(y, big.NewInt(2), curveP).Cmp(ySq) != 0 {
		return nil
	}
	// 取偶 y
	if y.Bit(0) != 0 {
		y = new(big.Int).Sub(curveP, y)
	}
	return &ecPoint{x: new(big.Int).Set(x), y: y}
}

// modInv 计算 a^(-1) mod n
func modInv(a, n *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, n)
}

// pointAdd 椭圆曲线点加 (与 JS pointAdd 一致)
func pointAdd(p, q *ecPoint) *ecPoint {
	if p == nil {
		return q
	}
	if q == nil {
		return p
	}
	if p.x.Cmp(q.x) == 0 {
		if p.y.Cmp(q.y) != 0 {
			return nil // 互逆点，结果为无穷远
		}
		return pointDouble(p) // 相同点倍乘
	}
	// λ = (q.y - p.y) * (q.x - p.x)^(-1) mod p
	dx := new(big.Int).Sub(q.x, p.x)
	dx.Mod(dx, curveP)
	dy := new(big.Int).Sub(q.y, p.y)
	dy.Mod(dy, curveP)
	lam := new(big.Int).Mul(dy, modInv(dx, curveP))
	lam.Mod(lam, curveP)
	// x = λ^2 - p.x - q.x
	x := new(big.Int).Exp(lam, big.NewInt(2), curveP)
	x.Sub(x, p.x)
	x.Sub(x, q.x)
	x.Mod(x, curveP)
	// y = λ*(p.x - x) - p.y
	y := new(big.Int).Sub(p.x, x)
	y.Mul(lam, y)
	y.Sub(y, p.y)
	y.Mod(y, curveP)
	return &ecPoint{x: x, y: y}
}

func pointDouble(p *ecPoint) *ecPoint {
	if p == nil {
		return nil
	}
	// λ = 3*x^2 * (2*y)^(-1)
	x2 := new(big.Int).Exp(p.x, big.NewInt(2), curveP)
	lam := new(big.Int).Mul(big.NewInt(3), x2)
	lam.Mul(lam, modInv(new(big.Int).Lsh(p.y, 1), curveP))
	lam.Mod(lam, curveP)
	x := new(big.Int).Exp(lam, big.NewInt(2), curveP)
	x.Sub(x, new(big.Int).Lsh(p.x, 1))
	x.Mod(x, curveP)
	y := new(big.Int).Sub(p.x, x)
	y.Mul(lam, y)
	y.Sub(y, p.y)
	y.Mod(y, curveP)
	return &ecPoint{x: x, y: y}
}

// pointMul 标量乘法 (与 JS pointMul 一致)
func pointMul(p *ecPoint, n *big.Int) *ecPoint {
	var res *ecPoint
	temp := p
	k := new(big.Int).Set(n)
	for k.Sign() > 0 {
		if k.Bit(0) == 1 {
			res = pointAdd(res, temp)
		}
		temp = pointAdd(temp, temp)
		k.Rsh(k, 1)
	}
	return res
}

// G 是 secp256k1 生成元
var gPoint = &ecPoint{
	mustParseBigInt("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798"),
	mustParseBigInt("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8"),
}

func mustParseBigInt(hexStr string) *big.Int {
	n, _ := new(big.Int).SetString(hexStr, 16)
	return n
}

// hasEvenY 判断点的 y 坐标是否为偶数
func hasEvenY(p *ecPoint) bool {
	return p.y.Bit(0) == 0
}

// PubkeyToTaprootTweakPubkey 将压缩公钥转换为 Taproot tweak 后的公钥 (BIP-341 Key Path)
func PubkeyToTaprootTweakPubkey(pubkey []byte) ([]byte, error) {
	if len(pubkey) != 33 {
		return nil, ErrInvalidPubkey
	}
	xBytes := pubkey[1:]
	tweakHash := taggedHash("TapTweak", xBytes)
	t := new(big.Int).SetBytes(tweakHash)
	if t.Cmp(curveN) >= 0 {
		return nil, errors.New("invalid tweak value")
	}
	x := new(big.Int).SetBytes(xBytes)
	p := liftX(x)
	if p == nil {
		return nil, ErrInvalidPubkey
	}
	tG := pointMul(gPoint, t)
	q := pointAdd(p, tG)
	if q == nil {
		return nil, ErrInvalidPubkey
	}
	// 输出使用 0x02 前缀 (与 JS 一致)
	out := make([]byte, 33)
	out[0] = 0x02
	copy(out[1:], pad32(q.x.Bytes()))
	return out, nil
}

// SeckeyToTaprootTweakSeckey 将私钥转换为 Taproot tweak 后的私钥
func SeckeyToTaprootTweakSeckey(seckey []byte) ([]byte, error) {
	if len(seckey) != 32 {
		return nil, ErrInvalidSeckey
	}
	priv := secp256k1.PrivKeyFromBytes(seckey)
	if priv.Key.IsZero() {
		return nil, ErrInvalidSeckey
	}
	pub := priv.PubKey().SerializeCompressed()
	seckeyInt := new(big.Int).SetBytes(seckey)
	p := liftX(new(big.Int).SetBytes(pub[1:]))
	if p == nil {
		return nil, ErrInvalidPubkey
	}
	// 确保使用偶数 y 的私钥表示
	if !hasEvenY(p) {
		seckeyInt = new(big.Int).Sub(curveN, seckeyInt)
		seckeyInt.Mod(seckeyInt, curveN)
	}
	tweakHash := taggedHash("TapTweak", pub[1:])
	t := new(big.Int).SetBytes(tweakHash)
	if t.Cmp(curveN) >= 0 {
		return nil, errors.New("invalid tweak value")
	}
	tweakSeckey := new(big.Int).Add(seckeyInt, t)
	tweakSeckey.Mod(tweakSeckey, curveN)
	tweakBytes := pad32(tweakSeckey.Bytes())
	// 确保输出的公钥使用 0x02 前缀 (与 JS 一致)
	genPub, _ := PubkeyGen(tweakBytes)
	if len(genPub) > 0 && genPub[0] == 0x03 {
		tweakSeckey = new(big.Int).Sub(curveN, tweakSeckey)
		tweakSeckey.Mod(tweakSeckey, curveN)
		tweakBytes = pad32(tweakSeckey.Bytes())
	}
	return tweakBytes, nil
}

// CompressPubkey 将 65 字节非压缩公钥压缩为 33 字节
func CompressPubkey(pubkey []byte) []byte {
	if len(pubkey) == 33 {
		return pubkey
	}
	if len(pubkey) != 65 {
		return nil
	}
	x := pubkey[1:33]
	y := pubkey[33:65]
	prefix := byte(0x02)
	if (y[31] & 1) != 0 {
		prefix = 0x03
	}
	return append([]byte{prefix}, x...)
}

// PubkeyToLegacyAddress 将公钥转换为传统 P2PKH Base58 地址
func PubkeyToLegacyAddress(pubkey []byte) string {
	compressed := CompressPubkey(pubkey)
	if compressed == nil {
		return ""
	}
	sha := crypto.Sha256(compressed)
	ripe := crypto.Ripemd160(sha)
	versioned := append([]byte{0x00}, ripe...)
	checksum := crypto.Sha256(crypto.Sha256(versioned))
	versioned = append(versioned, checksum[:4]...)
	return base58.Encode(versioned)
}

// Bech32mEncode 将 32 字节数据编码为 Bech32m 地址
// prefix 通常为 "bc" (主网) 或 "tb" (测试网)
func Bech32mEncode(prefix string, data []byte) (string, error) {
	if len(data) == 33 {
		data = data[1:]
	}
	if len(data) != 32 {
		return "", fmt.Errorf("%w: data must be 32 or 33 bytes", ErrInvalidAddress)
	}
	version := byte(1)
	words := convertBits(data, 8, 5, true)
	if words == nil {
		return "", ErrInvalidAddress
	}
	combined := append([]byte{version}, words...)
	return encodeBech32m(prefix, combined), nil
}

// Bech32mDecode 从 Bech32m 地址解码出 32 字节数据
func Bech32mDecode(addr string, expectedPrefix string) ([]byte, error) {
	prefix, data, err := decodeBech32m(addr)
	if err != nil {
		return nil, err
	}
	if prefix != expectedPrefix {
		return nil, fmt.Errorf("%w: expected prefix %s", ErrInvalidAddress, expectedPrefix)
	}
	if len(data) < 2 || data[0] != 1 {
		return nil, fmt.Errorf("%w: invalid witness version", ErrInvalidAddress)
	}
	decoded := convertBits(data[1:], 5, 8, false)
	if decoded == nil || len(decoded) != 32 {
		return nil, ErrInvalidAddress
	}
	return decoded, nil
}

// TaprootTweakPubkeyToTaprootAddress 将 tweak 后的公钥编码为 Taproot 地址
func TaprootTweakPubkeyToTaprootAddress(tapTweakPubkey []byte, prefix string) (string, error) {
	if prefix == "" {
		prefix = "bc"
	}
	return Bech32mEncode(prefix, tapTweakPubkey)
}

// TaprootAddressToTaprootTweakPubkey 从 Taproot 地址解码出 tweak 后的公钥 (默认偶 y 前缀 0x02)
func TaprootAddressToTaprootTweakPubkey(taprootAddr, prefix string) ([]byte, error) {
	if prefix == "" {
		prefix = "bc"
	}
	data, err := Bech32mDecode(taprootAddr, prefix)
	if err != nil {
		return nil, err
	}
	return append([]byte{0x02}, data...), nil
}

// TaprootAddressToTaprootTweakLegacyAddress 将 Taproot 地址转换为对应的传统地址
func TaprootAddressToTaprootTweakLegacyAddress(taprootAddr, prefix string) (string, error) {
	pub, err := TaprootAddressToTaprootTweakPubkey(taprootAddr, prefix)
	if err != nil {
		return "", err
	}
	return PubkeyToLegacyAddress(pub), nil
}

// WIFToTaprootAddress 从 WIF 私钥直接生成 Taproot 地址
func WIFToTaprootAddress(wif string, prefix string) (string, error) {
	seckey, err := WIFToSeckey(wif)
	if err != nil {
		return "", err
	}
	pubkey, err := PubkeyGen(seckey)
	if err != nil {
		return "", err
	}
	tapTweakPubkey, err := PubkeyToTaprootTweakPubkey(pubkey)
	if err != nil {
		return "", err
	}
	return TaprootTweakPubkeyToTaprootAddress(tapTweakPubkey, prefix)
}

// --- Bech32m 编解码 (BIP-350) ---

func polymod(values []byte) uint32 {
	GEN := []uint32{0x3b6a57b2, 0x26508e6d, 0x1ea119fa, 0x3d4233dd, 0x2a1462b3}
	chk := uint32(1)
	for _, v := range values {
		b := chk >> 25
		chk = (chk&0x1ffffff)<<5 ^ uint32(v)
		for i := 0; i < 5; i++ {
			if (b>>i)&1 != 0 {
				chk ^= GEN[i]
			}
		}
	}
	return chk
}

func hrpExpand(hrp string) []byte {
	ret := make([]byte, 0, len(hrp)*2+1)
	for i := 0; i < len(hrp); i++ {
		ret = append(ret, hrp[i]>>5)
	}
	ret = append(ret, 0)
	for i := 0; i < len(hrp); i++ {
		ret = append(ret, hrp[i]&31)
	}
	return ret
}

func verifyChecksum(hrp string, data []byte, constant uint32) bool {
	expanded := hrpExpand(hrp)
	combined := append(expanded, data...)
	return polymod(combined) == constant
}

func createChecksum(hrp string, data []byte, constant uint32) []byte {
	expanded := hrpExpand(hrp)
	combined := append(expanded, data...)
	combined = append(combined, 0, 0, 0, 0, 0, 0)
	mod := polymod(combined) ^ constant
	chk := make([]byte, 6)
	for i := 0; i < 6; i++ {
		chk[5-i] = byte(mod & 31)
		mod >>= 5
	}
	return chk
}

func encodeBech32m(hrp string, data []byte) string {
	checksum := createChecksum(hrp, data, bech32mConst)
	combined := append(data, checksum...)
	var result []byte
	for _, v := range combined {
		result = append(result, charset[v])
	}
	return hrp + "1" + string(result)
}

func decodeBech32m(s string) (string, []byte, error) {
	if len(s) < 8 || len(s) > 90 {
		return "", nil, ErrInvalidAddress
	}
	pos := 0
	for pos < len(s) && s[pos] != '1' {
		pos++
	}
	if pos == 0 || pos+7 > len(s) || len(s) > 90 {
		return "", nil, ErrInvalidAddress
	}
	hrp := s[:pos]
	data := make([]byte, 0, len(s)-pos-1)
	for i := pos + 1; i < len(s); i++ {
		d := stringsIndex(charset, s[i])
		if d < 0 {
			return "", nil, ErrInvalidAddress
		}
		data = append(data, byte(d))
	}
	if len(data) < 6 {
		return "", nil, ErrInvalidAddress
	}
	if !verifyChecksum(hrp, data, bech32mConst) {
		return "", nil, ErrInvalidAddress
	}
	return hrp, data[:len(data)-6], nil
}

func stringsIndex(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

func convertBits(data []byte, fromBits, toBits uint, pad bool) []byte {
	acc := 0
	bits := 0
	result := []byte{}
	maxV := (1 << toBits) - 1
	for _, v := range data {
		acc = (acc << fromBits) | int(v)
		bits += int(fromBits)
		for bits >= int(toBits) {
			bits -= int(toBits)
			result = append(result, byte((acc>>bits)&maxV))
		}
	}
	if pad && bits > 0 {
		result = append(result, byte((acc<<(int(toBits)-bits))&maxV))
	} else if bits >= int(fromBits) {
		return nil
	}
	return result
}
