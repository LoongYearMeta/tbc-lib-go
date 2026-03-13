# ECIES 加密

`bt.ECIES` 提供与 Electrum 兼容的 ECIES 消息加密，可与 tbc-lib-js 的 `ecies` 模块互操作。

## 实现说明

- 使用 **secp256k1** 椭圆曲线进行 ECDH 密钥交换
- 使用 **AES-128-CBC** + PKCS7 填充进行对称加密
- 使用 **HMAC-SHA256** 对完整密文进行完整性校验
- 消息格式（BIE1 变体）：`BIE1 | [Rbuf?] | ciphertext | HMAC`

其中：

- **BIE1**：4 字节 ASCII 常量
- **Rbuf**：发送方临时公钥（可选，压缩格式）
- **ciphertext**：AES-128-CBC 密文
- **HMAC**：32 字节（或 4 字节 shortTag）

## 选项 ECIESOptions

```go
type ECIESOptions struct {
    EphemeralKey bool   // 为 true 时使用临时密钥加密（默认 true）
    NoKey        bool   // 为 true 时密文不包含公钥 Rbuf（默认 false）
    ShortTag     bool   // 为 true 时 HMAC 截断为 4 字节（默认 false）
    Algorithm    string // 当前仅支持 "BIE1"
}
```

## 示例

### 发送消息给 Bob

```go
import (
    "github.com/sCrypt-Inc/go-bt/v2"
    "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

bobPriv, _ := secp256k1.GeneratePrivateKey()
bobPub := bobPriv.PubKey()

// 加密
ecies := bt.NewECIES(nil)
ecies.PublicKey(bobPub)
ciphertext, err := ecies.EncryptBIE1([]byte("a message"))

// Bob 解密
eciesDec := bt.NewECIES(nil)
eciesDec.PrivateKey(bobPriv)
plaintext, err := eciesDec.DecryptBIE1(ciphertext)
```

### Alice 与 Bob 双向通信（NoKey 模式）

```go
alicePriv, _ := secp256k1.GeneratePrivateKey()
bobPriv, _ := secp256k1.GeneratePrivateKey()

opts := &bt.ECIESOptions{NoKey: true}

iesAlice := bt.NewECIES(opts).PrivateKey(alicePriv).PublicKey(bobPriv.PubKey())
iesBob := bt.NewECIES(opts).PrivateKey(bobPriv).PublicKey(alicePriv.PubKey())

msgAlice, _ := iesAlice.EncryptBIE1([]byte("Hello Bob"))
plainBob, _ := iesBob.DecryptBIE1(msgAlice)

msgBob, _ := iesBob.EncryptBIE1([]byte("Hi Alice"))
plainAlice, _ := iesAlice.DecryptBIE1(msgBob)
```

### 发送方恢复消息

当 `EphemeralKey` 为 `false` 时，发送方可以解密自己发送的消息：

```go
opts := &bt.ECIESOptions{NoKey: true}
iesAlice := bt.NewECIES(opts).PrivateKey(alicePriv).PublicKey(bobPub)

msg, _ := iesAlice.EncryptBIE1([]byte("Hello Bob"))
recovered, _ := iesAlice.DecryptBIE1(msg)
```

### 便捷函数

```go
// 快速加密：使用临时密钥给指定公钥加密
ciphertext, err := bt.EncryptFor(bobPub, []byte("hello"), nil)

// 快速解密：使用私钥解密 BIE1 消息
plaintext, err := bt.DecryptWith(bobPriv, ciphertext, nil)
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js                    | tbc-lib-go                          |
|-------------------------------|-------------------------------------|
| new IES(opts)                 | bt.NewECIES(opts)                   |
| .publicKey(pub)               | .PublicKey(pub)                     |
| .privateKey(priv)             | .PrivateKey(priv)                   |
| .encrypt(msg)                 | .EncryptBIE1(msg)                   |
| .decrypt(enc)                 | .DecryptBIE1(enc)                   |
| opts.ephemeralKey             | opts.EphemeralKey                   |
| opts.noKey                    | opts.NoKey                          |
| opts.shortTag                 | opts.ShortTag                       |
