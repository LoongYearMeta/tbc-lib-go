# ECIES 加密

**参考：** [tbc-lib-js/docs/ecies.md](../../tbc-lib-js/docs/ecies.md)

`bt.ECIES` 提供与 Electrum 兼容的 BIE1 消息格式，可与官方 `tbc-lib-js` 的 `ecies` 子模块互操作（选项语义与 JS 文档中的 `ephemeralKey` / `noKey` / `shortTag` 一致）。

## 实现概要

- secp256k1 ECDH
- AES-128-CBC + PKCS7
- HMAC-SHA256（可选 `ShortTag`）
- 消息前缀 `BIE1` 及可选临时公钥 `Rbuf`

## 选项 `ECIESOptions`

| Go 字段 | JS 文档字段 | 含义 |
|---------|--------------|------|
| `EphemeralKey` | `ephemeralKey` | 默认 `true`；与私钥组合使用时的语义见源码 |
| `NoKey` | `noKey` | 密文是否省略对方用于解密的公钥材料 |
| `ShortTag` | `shortTag` | 是否使用截断 HMAC |
| `Algorithm` | — | 当前使用 `"BIE1"` |

## 发送给 Bob（与 JS 示例同构）

```go
import (
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
	bt "github.com/sCrypt-Inc/go-bt/v2"
)

bobPriv, _ := secp256k1.GeneratePrivateKey()
bobPub := bobPriv.PubKey()

cipher, err := bt.NewECIES(nil).PublicKey(bobPub).EncryptBIE1([]byte("a message"))
plain, err := bt.NewECIES(nil).PrivateKey(bobPriv).DecryptBIE1(cipher)
_ = plain
_ = err
```

## Alice / Bob 双向（NoKey）

```go
alicePriv, _ := secp256k1.GeneratePrivateKey()
bobPriv, _ := secp256k1.GeneratePrivateKey()
opts := &bt.ECIESOptions{NoKey: true}

iesAlice := bt.NewECIES(opts).PrivateKey(alicePriv).PublicKey(bobPriv.PubKey())
iesBob := bt.NewECIES(opts).PrivateKey(bobPriv).PublicKey(alicePriv.PubKey())

msgAlice, _ := iesAlice.EncryptBIE1([]byte("Hello Bob"))
plainBob, _ := iesBob.DecryptBIE1(msgAlice)
```

## 便捷函数

```go
cipher, err := bt.EncryptFor(bobPub, []byte("hello"), nil)
plain, err := bt.DecryptWith(bobPriv, cipher, nil)
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js | tbc-lib-go |
|------------|------------|
| `new IES(opts)` | `bt.NewECIES(opts)` |
| `.publicKey(pub)` | `.PublicKey(pub)` |
| `.privateKey(priv)` | `.PrivateKey(priv)` |
| `.encrypt(msg)` | `.EncryptBIE1([]byte(msg))` |
| `.decrypt(enc)` | `.DecryptBIE1(enc)` |
