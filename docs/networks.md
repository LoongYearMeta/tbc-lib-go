# Networks 网络配置

tbc-lib-go 支持 TBC 主网（livenet）、测试网（testnet）、回归测试网（regtest）以及 STN 网络。设计参考 tbc-lib-js 的 `Networks` 模块。

## 预定义网络

```go
import "github.com/sCrypt-Inc/go-bt/v2"

// 主网
bt.Livenet

// 测试网
bt.Testnet

// 回归测试网（本地开发）
bt.Regtest

// STN（Scaling Test Network）
bt.STN
```

## 网络结构

每个 `Network` 包含：

| 字段         | 说明                              |
|--------------|-----------------------------------|
| Name         | 网络名称，如 "livenet", "testnet" |
| Alias        | 别名，如 "mainnet"                |
| PubKeyHash   | P2PKH 地址版本字节（如 livenet 为 0x00） |
| PrivateKey   | 私钥 WIF 版本字节                 |
| ScriptHash   | P2SH 地址版本字节                 |
| XPubKey      | 扩展公钥魔数                      |
| XPrivKey     | 扩展私钥魔数                      |
| NetworkMagic | P2P 网络魔数（4 字节，小端序）    |
| Port         | P2P 端口                          |
| DNSServers   | DNS 种子节点列表                  |
| CashAddrPref | CashAddr 前缀（如 "bitcoincash"） |

## 网络常量示例

```go
// Livenet 典型配置
bt.Livenet.PubKeyHash   // 0x00 - 地址以 '1' 开头
bt.Livenet.PrivateKey   // 0x80
bt.Livenet.ScriptHash   // 0x05
bt.Livenet.Port         // 8333

// Testnet 典型配置
bt.Testnet.PubKeyHash   // 0x6f - 地址以 'm' 或 'n' 开头
bt.Testnet.PrivateKey   // 0xef
bt.Testnet.Port         // 18333
```

## 默认网络

```go
// 获取当前默认网络
current := bt.DefaultNetwork  // 默认为 bt.Livenet

// 如需使用测试网作为默认，可直接赋值
bt.DefaultNetwork = bt.Testnet
```

## Regtest 说明

Regtest 用于本地开发，可以程序化地即时生成区块。端口为 18444，魔数与 testnet 不同。

## 与 tbc-lib-js 的对应关系

| tbc-lib-js                    | tbc-lib-go                     |
|------------------------------|--------------------------------|
| Networks.livenet             | bt.Livenet                     |
| Networks.testnet             | bt.Testnet                     |
| Networks.defaultNetwork      | bt.DefaultNetwork              |
| Networks.enableRegtest()     | 使用 bt.Regtest                |
