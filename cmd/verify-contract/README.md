# 合约验证指南

## 📋 概述

本文档详细说明如何使用脚本解释器验证 TBC 链上的各种交易类型，包括普通交易和智能合约交易。

## 🔍 验证流程

### 基本验证流程

```
1. 获取交易数据
   ↓
2. 遍历所有输入
   ↓
3. 获取前序交易和输出
   ↓
4. 使用解释器执行脚本验证
   ↓
5. 检查验证结果
```

## 📝 验证类型

### 1. 普通P2PKH交易验证

**特点：**
- 单个输入和输出
- 使用标准锁定脚本（OP_DUP OP_HASH160 ... OP_CHECKSIG）
- 验证签名和公钥哈希

**代码示例：**

```go
func verifyStandardTransaction(txid string, network string) error {
    // 1. 获取交易
    tx, err := bt.FetchTXRaw(txid, network)
    if err != nil {
        return err
    }

    // 2. 获取输入和前序输出
    input := tx.InputIdx(0)
    prevTxID := input.PreviousTxIDStr()
    vout := input.PreviousTxOutIndex

    prevTx, _ := bt.FetchTXRaw(prevTxID, network)
    prevOutput := prevTx.OutputIdx(int(vout))

    // 3. 执行验证
    return interpreter.NewEngine().Execute(
        interpreter.WithTx(tx, 0, prevOutput),
        interpreter.WithForkID(),
        interpreter.WithAfterGenesis(),
    )
}
```

### 2. FT代币交易验证

**特点：**
- **Code Script**：包含代币转移验证逻辑
- **Tape Script**：存储代币余额和元数据
- 输出结构：`[code script, tape script]` 成对出现

**验证要点：**
1. 验证输入UTXO的code script
2. 验证tape script中的余额计算
3. 验证输出结构是否正确

**输出结构：**
```
输出0: Code Script (500 satoshis) - 验证逻辑
输出1: Tape Script (0 satoshis) - 元数据
输出2: Code Script (找零) - 可选
输出3: Tape Script (找零) - 可选
```

### 3. NFT交易验证

**特点：**
- **Code Script**：包含NFT转移验证逻辑
- **Hold Script**：持有脚本（P2PKH）
- **Tape Script**：存储NFT元数据

**输出结构：**
```
输出0: Code Script (200 satoshis) - 验证逻辑
输出1: Hold Script (100 satoshis) - 持有脚本
输出2: Tape Script (0 satoshis) - 元数据
```

**验证要点：**
1. 验证code script中的NFT转移逻辑
2. 验证hold script的地址匹配
3. 验证tape script的元数据完整性

### 4. PoolNFT交易验证

**特点：**
- 复杂的流动性管理逻辑
- 使用 `OP_PUSH_META` 获取交易元数据
- 使用 `OP_PARTIAL_HASH` 进行部分哈希计算
- 验证恒定乘积公式：`x * y = k`

**验证要点：**
1. 验证输入/输出的流动性代币数量
2. 验证交换比例计算
3. 验证手续费计算

## 🛠️ 关键操作码

### OP_PUSH_META
从交易中获取元数据：
- `1`: 交易版本
- `2`: 交易锁定时间
- `3`: 输入数量
- `4`: 输出数量
- `5`: 所有输入的哈希
- `6`: 当前输入数据
- `7`: 所有输出的哈希

### OP_PARTIAL_HASH
执行部分SHA256哈希计算，用于验证大脚本的完整性。

## 📊 验证步骤详解

### 步骤1: 获取交易数据

```go
tx, err := bt.FetchTXRaw(txid, network)
```

### 步骤2: 遍历输入

```go
for i := 0; i < tx.InputCount(); i++ {
    input := tx.InputIdx(i)
    
    // 跳过Coinbase输入
    if isCoinbaseInput(input) {
        continue
    }
    
    // 处理输入...
}
```

### 步骤3: 获取前序交易

```go
prevTxID := input.PreviousTxIDStr()
vout := input.PreviousTxOutIndex

prevTx, err := bt.FetchTXRaw(prevTxID, network)
prevOutput := prevTx.OutputIdx(int(vout))
```

### 步骤4: 执行脚本验证

```go
err := interpreter.NewEngine().Execute(
    interpreter.WithTx(tx, inputIdx, prevOutput),
    interpreter.WithForkID(),        // 启用ForkID
    interpreter.WithAfterGenesis(),  // 启用新脚本规则
)
```

## ⚠️ 注意事项

### 1. Coinbase交易
Coinbase交易（挖矿奖励）无法使用标准脚本验证，需要特殊处理：

```go
func isCoinbaseInput(input *bt.Input) bool {
    prevTxID := input.PreviousTxID()
    vout := input.PreviousTxOutIndex
    
    allZero := true
    for _, b := range prevTxID {
        if b != 0 {
            allZero = false
            break
        }
    }
    
    return allZero && vout == 0xFFFFFFFF
}
```

### 2. 合约输出结构
合约交易通常有多个输出，需要验证输出结构：
- Code Script 和 Tape Script 成对出现
- 金额要求：Code Script通常有最小金额（如500 satoshis）
- Tape Script通常为0 satoshis

### 3. 前序交易依赖
某些合约（如NFT）需要验证前序交易（pre_tx）和前前序交易（pre_pre_tx）：
- 需要获取完整的交易链
- 验证交易之间的关联性

### 4. 脚本标志
根据网络和交易类型选择合适的脚本标志：
- `WithForkID()`: 用于BCH等分叉链
- `WithAfterGenesis()`: 启用新脚本规则（支持新操作码）

## 🔧 调试技巧

### 1. 查看脚本内容

```go
scriptASM, _ := output.LockingScript.ToASM()
fmt.Printf("脚本: %s\n", scriptASM)
```

### 2. 查看交易结构

```go
fmt.Printf("输入数量: %d\n", tx.InputCount())
fmt.Printf("输出数量: %d\n", tx.OutputCount())
for i := 0; i < tx.OutputCount(); i++ {
    output := tx.OutputIdx(i)
    fmt.Printf("输出 %d: 金额=%d\n", i, output.Satoshis)
}
```

### 3. 错误处理

```go
if err := interpreter.NewEngine().Execute(...); err != nil {
    // 错误信息会包含具体的失败原因
    // 例如: "OP_EQUALVERIFY failed" 表示公钥哈希不匹配
    //      "OP_CHECKSIG failed" 表示签名验证失败
    log.Printf("验证失败: %v", err)
}
```

## 📚 相关资源

- 脚本解释器文档: `bscript/interpreter/`
- 操作码实现: `bscript/interpreter/operations.go`
- 示例代码: `examples/sample_transaction/sample_transaction.go`

## 🚀 执行命令

```bash
# 验证普通交易
go run ./cmd/verify-interpreter

# 验证合约交易
go run ./cmd/verify-contract
```
