# Block 区块

`bt.Block` 表示 TBC 网络中的一个区块，包含区块头和交易列表。对应 tbc-lib-js 的 `Block` 类。

## 结构说明

```go
type Block struct {
    Header       *BlockHeader  // 区块头
    Transactions []*Tx         // 交易列表
}
```

## 区块头 BlockHeader

区块头包含：

- **Version**：区块版本
- **PrevBlockHash**：前一区块哈希
- **MerkleRoot**：默克尔根
- **Time**：时间戳
- **Bits**：难度目标
- **Nonce**：工作量证明随机数

## 解析区块

```go
// 从 hex 字符串解析
block, err := bt.NewBlockFromString(hexEncodedBlock)

// 从字节解析
block, err := bt.NewBlockFromBytes(blockBytes)

// 从原始区块载荷解析（跳过前 8 字节）
block, err := bt.NewBlockFromRawBlock(rawPayload)
```

## 验证

```go
// 验证默克尔根（若提供完整交易列表）
valid := block.ValidMerkleRoot()

// 验证工作量证明
valid := block.Header.ValidProofOfWork()

// 验证时间戳（未过于超前）
valid := block.Header.ValidTimestamp()
```

## 遍历交易

```go
block, _ := bt.NewBlockFromString(hexBlock)
for _, tx := range block.Transactions {
    fmt.Println(tx.TxIDStr())
}
```

## 常量

```go
bt.MaxBlockSize   // 128000000 字节
bt.BlockStartOffset  // 8（原始区块中区块数据的起始偏移）
bt.NullHash       // 32 字节零哈希
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js           | tbc-lib-go                    |
|----------------------|-------------------------------|
| new Block(hex)       | bt.NewBlockFromString(hex)    |
| block.validMerkleRoot() | block.ValidMerkleRoot()   |
| block.header         | block.Header                  |
| block.transactions   | block.Transactions            |
