# Script 脚本

`bscript.Script` 用于构建和解析 TBC 交易中使用的脚本。所有交易的输入和输出都包含脚本，验证时会将输入脚本与输出脚本拼接执行。

## 脚本类型常量

```go
bscript.ScriptTypePubKeyHash   // "pubkeyhash"
bscript.ScriptTypePubKey      // "pubkey"
bscript.ScriptTypeMultiSig    // "multisig"
bscript.ScriptTypeNullData    // "nulldata"
bscript.ScriptTypeSecureHash  // "securehash"
```

## 创建脚本

```go
// 空脚本
script := bscript.NewScript()

// 从 hex 创建
script, err := bscript.NewFromHexString("76a914...")

// 从字节创建
script := bscript.NewFromBytes(scriptBytes)
```

## 常见输出脚本

### Pay to Public Key Hash (P2PKH)

最常用的支付脚本，支付到地址（公钥哈希的 Base58Check 编码）：

```go
script := bscript.NewP2PKHFromAddress(address)
// 或
script := bscript.NewP2PKHFromPubKeyHash(pubKeyHash)
```

### Pay to Public Key (P2PK)

```go
script := bscript.NewP2PKFromPubKey(pubKey)
```

### Pay to Multisig (P2MS)

```go
script, err := bscript.NewP2MultiSig(pubKeys, threshold)
```

### Pay to Script Hash (P2SH)

**注意**：TBC 上不推荐使用 P2SH。

### 数据输出 (OP_RETURN)

```go
script, err := bscript.BuildDataOut(data, "hex")
// 或使用 tx.AddOpReturnOutput(data) 直接添加 OP_RETURN 输出
```

## 解析与识别

```go
// 解析原始脚本
script := bscript.NewFromBytes(rawScript)

// 判断脚本类型
script.IsP2PKH()    // 是否为 P2PKH 输出
script.IsP2PK()     // 是否为 P2PK 输出
script.IsP2MS()     // 是否为多签输出
script.IsP2SH()     // 是否为 P2SH 输出
script.IsNullData() // 是否为 OP_RETURN 数据输出
```

## 脚本解释与验证

使用 `bscript/interpreter` 包进行脚本验证：

```go
import "github.com/sCrypt-Inc/go-bt/v2/bscript/interpreter"

engine := interpreter.NewEngine()

// 方式一：使用锁定脚本和解锁脚本验证
err := engine.Execute(
    interpreter.WithScripts(lockingScript, unlockingScript),
    interpreter.WithAfterGenesis(),
    interpreter.WithForkID(),
)
if err == nil {
    // 验证通过
}

// 方式二：使用交易和输入索引验证
err = engine.Execute(
    interpreter.WithTx(tx, inputIndex, previousOutput),
    interpreter.WithAfterGenesis(),
    interpreter.WithForkID(),
)
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js                      | tbc-lib-go                           |
|---------------------------------|--------------------------------------|
| Script.buildPublicKeyHashOut(addr) | bscript.NewP2PKHFromAddress(addr) |
| Script.buildDataOut(data)       | bscript.BuildDataOut(data, "hex")    |
| script.isPublicKeyHashOut()     | script.IsP2PKH()                     |
| script.isMultisigOut()          | script.IsP2MS()                      |
| Interpreter().verify(...)       | interpreter.NewEngine().Execute(WithScripts/WithTx...) |
