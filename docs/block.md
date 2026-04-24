# Block 区块

**参考：** [tbc-lib-js/docs/block.md](../../tbc-lib-js/docs/block.md)

`tbc.Block` 表示包含区块头与交易列表的区块，可从 hex 或字节反序列化；在提供完整交易列表时可校验默克尔根，**不保证**每笔交易的花费合法性（与 JS 文档表述一致）。

## 结构

```go
type Block struct {
	Header       *BlockHeader
	Transactions []*Tx
}
```

## 区块头 `BlockHeader`

含 `Version`、`PrevHash`、`MerkleRoot`、`Time`、`Bits`、`Nonce` 等字段，详见 `blockheader.go`。

## 解析

```go
block, err := tbc.NewBlockFromString(hexEncodedBlock)
block, err = tbc.NewBlockFromBytes(blockBytes)
block, err = tbc.NewBlockFromRawBlock(rawPayload) // 含 8 字节前缀的原始载荷时
```

## 校验

```go
ok := block.ValidMerkleRoot()
ok = block.Header.ValidProofOfWork()
ok = block.Header.ValidTimestamp()
```

## 遍历交易

```go
block, _ := tbc.NewBlockFromString(hexBlock)
for _, tx := range block.Transactions {
	_ = tx.TxID() // hex 字符串
}
```

## 常量

```go
_ = tbc.MaxBlockSize
_ = tbc.BlockStartOffset
_ = tbc.NullHash
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js | tbc-lib-go |
|------------|------------|
| `new Block(hex)` | `tbc.NewBlockFromString(hex)` |
| `block.validMerkleRoot()` | `block.ValidMerkleRoot()` |
| `block.header` | `block.Header` |
| `block.transactions` | `block.Transactions` |
