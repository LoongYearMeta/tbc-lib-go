/*
Package bt 提供创建和操作 Turing BC (TBC) 交易所需的核心功能。

本库与官方 JavaScript 库 [tbc-lib-js](https://github.com/TuringBitChain/tbc-lib-js)（npm 包名 `tbc-lib-js`）在交易与脚本等核心概念上保持对应，便于跨语言开发；人类可读说明见仓库内 docs/ 目录。

主要类型与功能：

  - UTXO: 未花费输出，用于构造交易输入
  - Tx: 交易构建、签名、序列化
  - Block: 区块解析与验证
  - Network: 网络配置（livenet/testnet/regtest/stn）
  - ECIES: 与 Electrum 兼容的消息加密
  - bscript: 脚本构建与解析

详细文档参见 docs/ 目录。
*/
package bt
