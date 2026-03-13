# tbc-lib-go 基础库文档

## 简介

tbc-lib-go 是 Turing BC (TBC) 区块链的 Go 语言基础库，提供创建和操作 TBC 交易所需的核心功能。与 tbc-lib-js  JavaScript 库在 API 设计上保持对应，便于跨语言开发。

## 快速开始

```bash
go get github.com/sCrypt-Inc/go-bt/v2
```

## 文档索引

### 地址与密钥管理

- [网络配置](networks.md) - 使用不同网络（livenet、testnet、regtest、stn）
- 私钥与公钥（参见 `bt` 包内文档）

### 支付处理

- [交易类](transaction.md) - Tx 交易构建与签名
- [未花费输出](unspentoutput.md) - UTXO 的表示与使用

### TBC 内部结构

- [脚本](script.md) - bscript 脚本的构建与解析
- [区块](block.md) - Block 与 BlockHeader

### 扩展功能

- [ECIES 加密](ecies.md) - 与 electrum 兼容的 ECIES 消息加密

### 合约与 API

- FT（同质化代币） - 参见 `api_ft.go`
- NFT（非同质化代币） - 参见 `api_nft.go`
- Pool NFT - 参见 `api_pool.go`

## 示例

### 创建交易

```go
package main

import (
    "github.com/sCrypt-Inc/go-bt/v2"
)

func main() {
    tx := bt.NewTx()
    tx.From(
        &bt.UTXO{
            TxID:          txIDBytes,
            Vout:          0,
            LockingScript: lockingScript,
            Satoshis:      100000,
        },
    )
    tx.To(recipientAddress, 50000)
    tx.Change(changeAddress)
    tx.Sign(privateKey)
}
```

### 解析区块

```go
block, err := bt.NewBlockFromString(hexEncodedBlock)
if err != nil {
    // 处理错误
}
for _, tx := range block.Transactions {
    // 遍历区块中的交易
}
```

### ECIES 加密消息

```go
ecies := bt.NewECIES(nil)
ecies.PublicKey(recipientPubKey)
ciphertext, _ := ecies.EncryptBIE1([]byte("hello"))
```
