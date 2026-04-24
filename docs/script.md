# Script 脚本

**参考：** [tbc-lib-js/docs/script.md](../../tbc-lib-js/docs/script.md)

`bscript.Script` 用于构造、解析与识别常见锁定脚本；验证时由 `bscript/interpreter` 执行拼接后的脚本。

## 类型常量（字符串标签）

```go
import "github.com/LoongYearMeta/tbc-lib-go/bscript"

_ = bscript.ScriptTypePubKeyHash
_ = bscript.ScriptTypeMultiSig
_ = bscript.ScriptTypeNullData
```

## 创建与解析

```go
s := bscript.NewScript()
s, err := bscript.NewFromHexString("76a914...")
s = bscript.NewFromBytes(raw)
s, err = bscript.NewFromASM("OP_DUP OP_HASH160 ...")
```

## 常见输出

### P2PKH

```go
s, err := bscript.NewP2PKHFromAddress(addressString)
// 或 NewP2PKHFromPubKeyHash、NewP2PKHFromPubKeyEC 等
```

### P2PK

```go
s, err := bscript.NewP2PKFromPubKey(pubKey)
```

### P2MS（多签）

本库未暴露与 JS `Script.buildMultisigOut(pubkeys, m)` 同名的单函数；通常使用 **`NewFromASM`** 或 **`NewFromHexString`** 载入与 JS 文档相同的裸脚本，再用 **`IsMultisigOut()`** 校验。JS 文档中的 2-of-3 示例 hex 可直接用于对齐测试。

### P2SH

**TBC 上不推荐使用 P2SH**（与 JS 文档一致）。

### OP_RETURN

```go
s, err := bscript.BuildDataOut(dataBytes, "") // 或按 `script_doc_test` 使用的 encoding 参数
```

也可在交易上使用 `tx.AddOpReturnOutput(data)`。

## 类型识别

```go
script.IsP2PKH()
script.IsP2PK()
script.IsMultisigOut()
script.IsP2SH()
script.IsDataOut() // OP_RETURN 数据输出
```

## 解释器

`tbc-lib-js` 的 `Interpreter#verify(inputScript, outputScript)` 参数顺序为 **先解锁脚本、后锁定脚本**。Go 的 `WithScripts` 为 **`WithScripts(lockingScript, unlockingScript)`**，顺序与 JS **相反**，请注意。

```go
import "github.com/LoongYearMeta/tbc-lib-go/bscript/interpreter"

err := interpreter.NewEngine().Execute(
	interpreter.WithScripts(lockingScript, unlockingScript),
	interpreter.WithAfterGenesis(),
	interpreter.WithForkID(),
)

// 或在完整交易上下文中
err = interpreter.NewEngine().Execute(
	interpreter.WithTx(tx, inputIndex, previousOutput),
	interpreter.WithAfterGenesis(),
	interpreter.WithForkID(),
)
```

## 与 tbc-lib-js 的对应关系

| tbc-lib-js | tbc-lib-go |
|------------|------------|
| `Script.buildPublicKeyHashOut` | `bscript.NewP2PKHFromAddress` 等 |
| `Script.buildPublicKeyOut` | `bscript.NewP2PKFromPubKey` |
| `Script.buildMultisigOut` | `NewFromASM` / `NewFromHexString` + `IsMultisigOut` |
| `Script.buildDataOut` | `bscript.BuildDataOut` |
| `Interpreter().verify(in, out)` | `Execute(WithScripts(lock, unlock), ...)` |
