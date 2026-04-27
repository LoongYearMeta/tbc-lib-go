# tbc-lib-go

> TuringBitChain (TBC) 的 Go SDK：交易构造、脚本、区块、签名、ECIES、消息签名等链上核心能力。

`tbc-lib-go` 与官方 JavaScript 库 [`tbc-lib-js`](https://github.com/TuringBitChain/tbc-lib-js)
**保持 API 与序列化结果对齐**（fee 估算、sighash、网络参数局等都按 JS 校准）。
许多源文件以 `// 与 tbc-lib-js X.Y 一致` 形式注明对齐位置，便于跨语言对照。

- Module path：`github.com/LoongYearMeta/tbc-lib-go`
- 默认网络：Livenet；DustLimit = 1（不是 BTC 的 546）；默认 fee quote = 100 sat / 1 KB（与 tbc-lib-js `Transaction.FEE_PER_KB = 100` 对齐）

## 安装

```bash
go get github.com/LoongYearMeta/tbc-lib-go
```

要求一个 [当前受支持的 Go 版本](https://go.dev/doc/devel/release)；模块声明 `go 1.17`。

## 快速开始

链式 API 与 `tbc-lib-js` 中 `Transaction().from().to().change().sign()` 思路一致：

```go
package main

import (
    "context"

    tbc "github.com/LoongYearMeta/tbc-lib-go"
    "github.com/LoongYearMeta/tbc-lib-go/bec"
    "github.com/LoongYearMeta/tbc-lib-go/unlocker"
)

func example(utxo *tbc.UTXO, priv *bec.PrivateKey, toAddr, changeAddr string) {
    ctx := context.Background()
    tx := tbc.NewTx().
        FromChain(utxo).
        To(toAddr, 50_000).
        Change(changeAddr, nil) // nil → 使用默认 FeeQuote

    tx.Sign(ctx, &unlocker.Getter{PrivateKey: priv})
    _ = tx.String()
}
```

链式方法（`FromChain` / `To` / `Change` / `Sign`）**遇错 panic**，对应 JS 的 throw。
若需要返回 error，请使用同名前缀的非链式版本：`From` / `FromUTXOs` / `PayToAddress`
/ `ChangeToAddress` / `FillAllInputs`。

更多主题示例见 [`docs/index.md`](docs/index.md) 及其下的分主题文档（transaction / block /
script / networks / unspentoutput / ecies）。

## 仓库布局

自 JS-style layout 重构起，仓库按 `tbc-lib-js/lib/` 拓扑同构组织，根包是**门面（facade）**：

```
github.com/LoongYearMeta/tbc-lib-go
├── tbc.go                 门面：re-export 历史 tbc.X 符号
├── transaction/           交易、输入/输出、费率、签名哈希、JSON
│   └── sighash/           SIGHASH flag 常量
├── script/                脚本构造、地址、BIP276；含 interpreter/
├── block/                 Block / BlockHeader / MerkleBlock
├── networks/              Livenet / Testnet / Regtest / STN
├── message/               消息签名 / 验签
├── ecies/                 ECIES (BIE1)
├── crypto/                Hash / ECDSA / Signature / BN / 随机数
├── encoding/              VarInt / Base58 / Base58Check / BufferReader/Writer / Hex
├── util/
│   ├── pushmeta/          OP_PUSH_META outpoint 辅助（TBC 独有）
│   └── partialsha256/     部分 SHA256 状态
├── taproot/               Taproot 数据结构（JS 无对应）
├── unlocker/              基于私钥的 Unlocker 实现
│
├── base58/ bec/ bip32/ chaincfg/ wif/   
│
├── bscript/               forwarder：转发到 script/
└── sighash/               forwarder：转发到 transaction/sighash/
```

详见 [`文件结构说明.md`](文件结构说明.md)。

### 两种 import 风格

- **门面（推荐用于历史代码）**：`tbc.Tx`、`tbc.NewTx()`、`tbc.Livenet`、`tbc.ErrNoUTXO` 等
  全部仍可用，门面层用类型别名做 zero-cost re-export，方法集与子包完全一致。
- **直接按子包**（推荐用于新代码）：`transaction.Tx == tbc.Tx`，`script.Script`、
  `block.Block`、`networks.Livenet`、`encoding.VarInt` 等可分别 import，依赖范围更小。

## 开发

```bash
make test            # lint + go test ./... -v（完整套件）
make test-short      # lint + go test ./... -v -test.short
make test-unit       # go test ./... -race -cover（CI 跑的就是这个）
make test-ci-no-race # CI 无 race 变体
make lint            # golangci-lint v1.45.2（已在 .golangci.yml 中钉版本）
make vet             # go vet ./...
make bench           # go test -bench=. -benchmem
make coverage        # 生成 coverage 报告
```

跑单个测试：

```bash
go test ./transaction/ -run TestJSEstimateSize_oneP2PKHInputOneP2PKHOutput -v
go test ./script/ -run TestScript -v
```

## 文档

- [`docs/index.md`](docs/index.md) — 文档入口与分主题索引
- 主题文档：[`docs/transaction.md`](docs/transaction.md)、[`docs/block.md`](docs/block.md)、
  [`docs/script.md`](docs/script.md)、[`docs/networks.md`](docs/networks.md)、
  [`docs/unspentoutput.md`](docs/unspentoutput.md)、[`docs/ecies.md`](docs/ecies.md)
- [`文件结构说明.md`](文件结构说明.md) — 目录与子包详解

## License

[ISC](LICENSE)。
