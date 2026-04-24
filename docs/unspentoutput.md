# UnspentOutput (UTXO)

**参考：** [tbc-lib-js/docs/unspentoutput.md](../../tbc-lib-js/docs/unspentoutput.md)

`tbc.UTXO` 表示未花费输出，用于 `FromUTXOs` / `FromChain`，对应 JS 中 `Transaction.UnspentOutput` 或 `from({ txId, outputIndex, script, satoshis })` 所携带的信息。

> 两种 import 风格任选：
> - **推荐（门面）**：`import tbc "github.com/LoongYearMeta/tbc-lib-go"` 直接用 `tbc.X`，下游零改动。
> - **进阶（子包）**：按需 import 对应子包（本主题对应 `.../transaction`），享受更细粒度的依赖控制。
>
> 本文档示例以**门面风格**为主。

## 字段

| 含义 | Go | JS 常见字段名 |
|------|-----|----------------|
| 前序交易 ID（32 字节，库内小端存储） | `TxID []byte` | `txId` / `txid` |
| 输出索引 | `Vout uint32` | `outputIndex` / `vout` |
| 锁定脚本 | `LockingScript *script.Script` | `script` / `scriptPubKey` |
| 金额（聪） | `Satoshis uint64` | `satoshis`；或 `amount`（TBC） |
| 序列号 | `SequenceNumber uint32` | 默认 `0xffffffff` |

## 节点 JSON（listunspent 等）

与 bitcoind 风格字段兼容，通过 `NodeJSON()` 做 `json.Marshal` / `Unmarshal`：

```go
import (
	"encoding/json"

	tbc "github.com/LoongYearMeta/tbc-lib-go"
)

nodeJSON := `[{
	"txid": "a0a08e397203df68392ee95b3f08b0b3b3e2401410a38d46ae0874f74846f2e9",
	"vout": 0,
	"scriptPubKey": "76a914089acaba6af8b2b4fb4bed3b747ab1e4e60b496588ac",
	"amount": 0.0007
}]`

var utxos tbc.UTXOs
if err := json.Unmarshal([]byte(nodeJSON), utxos.NodeJSON()); err != nil {
	panic(err)
}
```

## 手动构造

```go
import (
	"encoding/hex"

	tbc "github.com/LoongYearMeta/tbc-lib-go"
	"github.com/LoongYearMeta/tbc-lib-go/script"
)

txID, _ := hex.DecodeString("a0a08e397203df68392ee95b3f08b0b3b3e2401410a38d46ae0874f74846f2e9")
lockingScript, _ := script.NewFromHexString("76a914089acaba6af8b2b4fb4bed3b747ab1e4e60b496588ac")

utxo := &tbc.UTXO{
	TxID:          txID,
	Vout:          0,
	LockingScript: lockingScript,
	Satoshis:      70000,
}
```

## 常用方法

```go
txIDStr := utxo.TxIDStr()
scriptHex := utxo.LockingScriptHexString()
bb, _ := json.Marshal(utxo.NodeJSON())
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js | tbc-lib-go |
|------------|------------|
| `UnspentOutput({...})` | `&tbc.UTXO{...}` 或 JSON → `NodeJSON()` |
| `utxo.txId` | `utxo.TxID`（字节） / `TxIDStr()` |
| `utxo.outputIndex` | `utxo.Vout` |
| `utxo.script` / `scriptPubKey` | `utxo.LockingScript` |
| `utxo.satoshis` / `amount` | `utxo.Satoshis` |
