/*
Package script 提供 TBC 脚本的构建、解析与验证功能。

对应 tbc-lib-js 的 Script 与 Interpreter 模块，详见 tbc 包 docs/script.md。

主要类型：

  - Script: 脚本本体，支持 P2PKH、P2PK、P2MS、OP_RETURN 等常见类型
  - Chunk: 解析后的脚本片段
  - interpreter.Engine: 脚本解释与验证引擎

常用方法：

  - NewFromHexString/NewFromBytes: 从 hex 或字节创建脚本
  - NewP2PKHFromAddress: 创建 P2PKH 输出脚本
  - BuildDataOut: 创建 OP_RETURN 数据输出
  - IsP2PKH/IsP2MS/IsNullData: 判断脚本类型
*/
package script
