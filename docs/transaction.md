# Transaction 交易

`bt.Tx` 是 TBC 交易的核心类型，用于构建、签名和序列化交易。API 设计参考 tbc-lib-js 的 `Transaction` 类。

## 交易结构

交易包含：

- **Inputs**：输入列表，每个输入引用一个 UTXO
- **Outputs**：输出列表
- **Version**：交易版本号（通常为 1）
- **LockTime**：锁定时间（区块高度或时间戳）

## 创建交易

```go
// 创建空交易
tx := bt.NewTx()

// 从 hex 字符串解析
tx, err := bt.NewTxFromString(hexTx)

// 从字节解析
tx, err := bt.NewTxFromBytes(txBytes)
```

## 添加输入

```go
// 从单个 UTXO 添加输入
tx.From(&bt.UTXO{
    TxID:          txID,
    Vout:          0,
    LockingScript: script,
    Satoshis:      100000,
})

// 从多个 UTXO 添加输入
tx.From(utxo1).From(utxo2)
```

## 添加输出

```go
// 向地址支付（P2PKH）
tx.To(address, 50000)  // 50000 聪

// 添加自定义输出
tx.AddOutput(&bt.Output{
    LockingScript: script,
    Satoshis:      1000,
})

// 设置找零地址
tx.Change(changeAddress)
```

## 签名

```go
// 使用私钥签名
tx.Sign(privateKey)

// 自定义费用
tx.Fee(546)        // 最小非粉尘费用
tx.FeePerKb(1000)  // 每 KB 费用（聪）
```

## 序列化

```go
// 序列化为 hex
hexTx := tx.String()

// 序列化为字节
txBytes := tx.Bytes()

// 带检查的序列化（校验签名、费用等）
hexTx, err := tx.Serialize()
```

## 获取交易哈希

```go
txID := tx.TxID()        // 返回 []byte
txIDStr := tx.TxIDStr()  // 返回 hex 字符串
```

## 费用计算

当输出总和小于输入总和时，差额作为矿工费。通过 `Change()` 设置找零地址后，库会自动计算并添加找零输出。

## 时间锁定

```go
// 锁定到指定区块高度
tx.LockToBlockHeight(500000)

// 锁定到指定时间
tx.LockToTime(time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC))
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js                | tbc-lib-go             |
|---------------------------|------------------------|
| new Transaction()         | bt.NewTx()             |
| .from(utxo)               | .From(utxo)            |
| .to(address, amount)      | .To(address, amount)   |
| .change(address)          | .Change(address)       |
| .sign(privateKey)         | .Sign(privateKey)      |
| .fee(amount)              | .Fee(amount)           |
| .serialize()              | .Serialize()           |
