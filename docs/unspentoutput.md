# UnspentOutput (UTXO)

`bt.UTXO` 表示一个未花费的交易输出，用于创建交易输入。其设计对应 tbc-lib-js 的 `Transaction.UnspentOutput`。

## 结构说明

UTXO 包含以下字段：

- **TxID**：引用的交易 ID（32 字节，内部存储为小端序）
- **Vout**：该输出在交易中的索引
- **LockingScript**：锁定脚本（scriptPubKey），定义花费条件
- **Satoshis**：该输出包含的聪（satoshis）数量
- **SequenceNumber**：可选，序列号（默认 0xFFFFFFFF）

## 参数别名说明

为兼容 bitcoind 的 `listunspent` 等 RPC 返回格式，库支持以下 JSON 字段映射：

| 标准字段      | 别名              | 说明                           |
|---------------|-------------------|--------------------------------|
| txid          | TxID              | 交易 ID（hex 字符串）          |
| vout          | outputIndex       | 输出索引                       |
| scriptPubKey  | script / lockingScript | 锁定脚本（hex）        |
| amount        | -                 | TBC 单位（如 0.0007）          |
| satoshis      | -                 | 聪单位（如 70000）             |

通过 `NodeJSON()` 可进行节点格式（txid, vout, scriptPubKey, amount）的序列化/反序列化。

## 示例

### 从节点 JSON 创建 UTXO

```go
package main

import (
    "encoding/json"
    "github.com/sCrypt-Inc/go-bt/v2"
)

func main() {
    // 节点格式的 UTXO JSON（如 listunspent 返回，使用 txid/vout/scriptPubKey/amount）
    nodeJSON := `[{
        "txid": "a0a08e397203df68392ee95b3f08b0b3b3e2401410a38d46ae0874f74846f2e9",
        "vout": 0,
        "scriptPubKey": "76a914089acaba6af8b2b4fb4bed3b747ab1e4e60b496588ac",
        "amount": 0.0007
    }]`

    var utxos bt.UTXOs
    if err := json.Unmarshal([]byte(nodeJSON), utxos.NodeJSON()); err != nil {
        panic(err)
    }
    // utxos 现包含解析出的 UTXO 列表
}
```

### 手动构造 UTXO

```go
txID, _ := hex.DecodeString("a0a08e397203df68392ee95b3f08b0b3b3e2401410a38d46ae0874f74846f2e9")
lockingScript, _ := bscript.NewFromHexString("76a914089acaba6af8b2b4fb4bed3b747ab1e4e60b496588ac")

utxo := &bt.UTXO{
    TxID:          txID,
    Vout:          0,
    LockingScript: lockingScript,
    Satoshis:      70000,
}
```

### 常用方法

```go
// 获取 TxID 字符串
txIDStr := utxo.TxIDStr()

// 获取锁定脚本的 hex 字符串
scriptHex := utxo.LockingScriptHexString()

// 用于 JSON 序列化（节点格式）
bb, _ := json.Marshal(utxo.NodeJSON())
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js               | tbc-lib-go       |
|--------------------------|------------------|
| UnspentOutput({...})     | &bt.UTXO{...}    |
| utxo.txId                | utxo.TxID        |
| utxo.outputIndex         | utxo.Vout        |
| utxo.script / scriptPubKey | utxo.LockingScript |
| utxo.satoshis / amount   | utxo.Satoshis    |
