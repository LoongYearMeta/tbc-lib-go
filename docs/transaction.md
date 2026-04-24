# Transaction 交易

**参考：** [tbc-lib-js/docs/transaction.md](../../tbc-lib-js/docs/transaction.md)

`tbc.Tx` 对应官方库中的 `Transaction`：管理输入、输出、版本号与 `LockTime`（`uint32`），支持链式构造与序列化。

## 交易结构

- **Inputs**：对前序输出的引用与解锁脚本
- **Outputs**：本交易创建的输出
- **Version**：通常为 `1`（TBC 高版本交易另有约定，见源码与 JS 侧一致部分）
- **LockTime**：锁定时间；可直接赋值 `tx.LockTime`，语义与 [Bitcoin locktime](https://bitcoin.org/en/developer-guide#locktime-and-sequence-number) 一致

## 创建与反序列化

```go
tx := tbc.NewTx()

tx, err := tbc.NewTxFromString(hexTx)
tx, err = tbc.NewTxFromBytes(txBytes)
```

## 添加输入

底层 API（返回 `error`）：

```go
if err := tx.FromUTXOs(utxo1, utxo2); err != nil { /* ... */ }
// 或 From(prevTxIDHex, vout, lockingScriptHex, satoshis)
```

与 JS `transaction.from(utxos)` 相同的**链式**写法（失败时 `panic`，便于一行写完）：

```go
tx := tbc.NewTx().FromChain(utxo1, utxo2)

tx = tbc.NewTx().FromStringChain(
	"a0a08e397203df68392ee95b3f08b0b3b3e2401410a38d46ae0874f74846f2e9",
	0,
	"76a914089acaba6af8b2b4fb4bed3b747ab1e4e60b496588ac",
	70000,
)
```

## 添加输出

```go
tx.To(address, satoshis) // PayToAddress，与 JS .to 对应

outputs := []tbc.ToOutput{{Address: "1A...", Amount: 1000}, {Address: "1B...", Amount: 2000}}
tx.ToMultiple(outputs)

// 任意锁定脚本 + 金额
_ = tx.PayTo(lockingScript, amountSat)
_ = tx.AddOutput(&tbc.Output{LockingScript: script, Satoshis: n})
```

## 找零与费率

与 JS 的 `.change(addr)` + `.fee(sat)` 组合对应：Go 使用 **`Change(address string, feeQuote *tbc.FeeQuote)`**。`feeQuote == nil` 时使用 `tbc.NewFeeQuote()` 的默认报价（内部再区分 standard / data 等，见 `fees.go`）。

```go
tx.Change(changeAddress, nil)
```

单笔矿工费由找零逻辑与 `FeeQuote` 共同决定；若需与 JS 一样显式指定「固定手续费」，请在设置输出后、调用 `Change` 前阅读 `ChangeToAddress` / `FeeQuote` 相关源码或合约层封装。

## 签名

与 JS `.sign(privateKey, sighashType)` 不同，Go 使用 **`context.Context` + `UnlockerGetter`**（常用 `unlocker.Getter` 携带 `*bec.PrivateKey`）。链式方法：

```go
import (
	"context"
	"github.com/LoongYearMeta/tbc-lib-go/unlocker"
)

ctx := context.Background()
tx.Sign(ctx, &unlocker.Getter{PrivateKey: privKey})
```

等价于对全部输入调用 `FillAllInputs`；若需按输入处理错误，请直接使用 `FillAllInputs` / `FillInput` 而非链式 `Sign`。

## 序列化

```go
hexTx := tx.String()
raw := tx.Bytes()
hexChecked, err := tx.Serialize() // 带检查
```

## 交易 ID

```go
idHex := tx.TxID()       // hex 字符串
idBytes := tx.TxIDBytes() // 32 字节（内部哈希字节序与链上展示一致）
```

## 费用与检查（对照 JS 文档）

JS 文档中的 `serialize({ disableLargeFees: ... })` 等选项，在 Go 中由 `Serialize` / `CanBeDeep` 等路径体现；具体字段名与默认值以 Go 源码为准。常见常量：`FeeQuote`、`FeeTypeStandard` / `FeeTypeData` 等与 `fees.go` 中定义一致。

## 多签与部分签名

JS 文档描述了 `getSignatures` / `applySignature` 等流程。Go 库在输入层通过 `Unlocker` 与签名哈希类型组合完成解锁；多签场景通常需要自定义 `UnlockerGetter` 或在合约库中组装解锁脚本。脚本层识别见 [script.md](script.md)。

## 与 tbc-lib-js 的对应关系（速查）

| tbc-lib-js | tbc-lib-go |
|------------|------------|
| `new Transaction()` | `tbc.NewTx()` |
| `.from(utxo)` | `.FromChain(utxo)` 或 `FromUTXOs` |
| `.to(addr, amount)` | `.To(addr, amount)` |
| `.change(addr)` | `.Change(addr, feeQuote)`，`feeQuote` 可为 `nil` |
| `.fee(sat)` | 由 `Change` + `FeeQuote` 路径体现（非同名方法） |
| `.sign(priv, sighash)` | `.Sign(ctx, unlockerGetter)` |
| `.serialize()` | `.Serialize()` / `.String()` |
