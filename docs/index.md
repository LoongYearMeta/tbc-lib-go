# tbc-lib-go 基础库文档

## 简介

`tbc-lib-go`（Go 模块：`github.com/LoongYearMeta/tbc-lib-go`）提供 Turing BC (TBC) 交易的构造、解析、签名、脚本与区块处理等**链上核心**能力，在类型与调用习惯上与官方 JavaScript 库 **[tbc-lib-js](https://github.com/TuringBitChain/tbc-lib-js)**（当前工作区同步版本 **npm 1.0.30**）的文档与 README 示例对齐，便于跨语言对照。

官方 JS 文档入口与源码中的 `docs/` 目录一致，见仓库内 [docs/index.md](../../tbc-lib-js/docs/index.md)。

## 快速开始

```bash
go get github.com/LoongYearMeta/tbc-lib-go
```

## 文档索引（与 tbc-lib-js/docs 主题对应）

| 主题 | Go 文档 | JS 参考 |
|------|---------|---------|
| 网络 | [networks.md](networks.md) | [networks.md](../../tbc-lib-js/docs/networks.md) |
| 区块 | [block.md](block.md) | [block.md](../../tbc-lib-js/docs/block.md) |
| 脚本 | [script.md](script.md) | [script.md](../../tbc-lib-js/docs/script.md) |
| 交易 | [transaction.md](transaction.md) | [transaction.md](../../tbc-lib-js/docs/transaction.md) |
| 未花费输出 | [unspentoutput.md](unspentoutput.md) | [unspentoutput.md](../../tbc-lib-js/docs/unspentoutput.md) |
| ECIES | [ecies.md](ecies.md) | [ecies.md](../../tbc-lib-js/docs/ecies.md) |

JS 文档中还列有 `address.md`、`privatekey.md` 等（Bitcore 风格索引）；Go 侧对应能力分散在 `script`（地址）、`github.com/libsv/go-bk/bec`（密钥）等包中，本库以交易与脚本为主线，密钥示例见下文。

## 链上 HTTP（索引器 / 广播）

本仓库根包**不再内置** HTTP 索引器客户端。若需与节点交互（拉取 UTXO、广播 `txraw` 等），请在应用层使用 **`github.com/sCrypt-Inc/tbc-contract-go/lib/api`**，或与官方生态中其它 API 封装组合，再与本文档中的 `tbc.Tx` / `tbc.UTXO` 衔接。

## 示例：创建并签名（链式 API）

与官方 README 中 `Transaction` 的 `.from().to().change().fee().sign()` 思路一致；Go 使用 `FromChain` / `To` / `Change` / `Sign`，其中 `Change` 需传入 `*tbc.FeeQuote`（可为 `nil` 使用默认），`Sign` 需 `context` 与 `UnlockerGetter`。

```go
package main

import (
	"context"

	"github.com/libsv/go-bk/bec"
	tbc "github.com/LoongYearMeta/tbc-lib-go"
	"github.com/LoongYearMeta/tbc-lib-go/script"
	"github.com/LoongYearMeta/tbc-lib-go/unlocker"
)

func example(utxo *tbc.UTXO, priv *bec.PrivateKey, toAddr, changeAddr string) {
	ctx := context.Background()
	tx := tbc.NewTx().
		FromChain(utxo).
		To(toAddr, 50_000).
		Change(changeAddr, nil)
	tx.Sign(ctx, &unlocker.Getter{PrivateKey: priv})
	_ = tx.String()
}
```

## 示例：解析区块

```go
block, err := tbc.NewBlockFromString(hexEncodedBlock)
if err != nil {
	return
}
for _, tx := range block.Transactions {
	_ = tx.TxID()
}
```

## 示例：ECIES

```go
ec := tbc.NewECIES(nil)
ec.PublicKey(recipientPubKey)
cipher, err := ec.EncryptBIE1([]byte("hello"))
_ = cipher
_ = err
```
