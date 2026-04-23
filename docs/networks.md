# Networks 网络配置

**参考：** [tbc-lib-js/docs/networks.md](../../tbc-lib-js/docs/networks.md)

`tbc-lib-go` 使用包级常量 `bt.Livenet`、`bt.Testnet`、`bt.Regtest`、`bt.STN` 描述网络参数（地址版本字节、P2P 魔数、端口等），与 JS 侧 `Networks.livenet` / `Networks.testnet` 等常量对象一一对应。

## 预定义网络

```go
import bt "github.com/sCrypt-Inc/go-bt/v2"

_ = bt.Livenet
_ = bt.Testnet
_ = bt.Regtest
_ = bt.STN
```

## 主要字段

| 字段 | 说明 |
|------|------|
| `Name` / `Alias` | 如 `livenet`、`mainnet`、`testnet` |
| `PubKeyHash` | P2PKH 地址版本字节 |
| `PrivateKey` | WIF 版本字节 |
| `ScriptHash` | P2SH 版本字节（TBC 上 P2SH 不推荐） |
| `XPubKey` / `XPrivKey` | BIP32 扩展键魔数 |
| `NetworkMagic` | P2P 魔数（4 字节） |
| `Port` | 默认 P2P 端口 |
| `DNSServers` | DNS 种子节点列表 |
| `CashAddrPref` | CashAddr 前缀（若使用） |

## 常量速览

```go
bt.Livenet.PubKeyHash  // 主网 P2PKH 前缀，如 0x00
bt.Livenet.Port        // 8333
bt.Testnet.PubKeyHash  // 测试网，如 0x6f
bt.Testnet.Port        // 18333
```

## 默认网络

```go
_ = bt.DefaultNetwork // 默认 livenet
bt.DefaultNetwork = bt.Testnet
```

## Regtest

JS 文档中的 `Networks.enableRegtest()` 会切换 testnet 的魔数与端口；Go 侧直接使用 **`bt.Regtest`** 常量即可用于本地回归环境（端口、魔数与 testnet 主配置不同，见 `network.go`）。

## 与 tbc-lib-js 的对应关系

| tbc-lib-js | tbc-lib-go |
|------------|------------|
| `Networks.livenet` | `bt.Livenet` |
| `Networks.testnet` | `bt.Testnet` |
| `Networks.defaultNetwork` | `bt.DefaultNetwork` |
| `Networks.enableRegtest()` | 使用 `bt.Regtest` |
