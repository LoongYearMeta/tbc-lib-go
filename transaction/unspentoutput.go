package transaction

import (
	"encoding/hex"

	"github.com/LoongYearMeta/tbc-lib-go/script"
)

// UTXO 表示未花费交易输出（Unspent Transaction Output），用于创建交易输入。
//
// 对应 tbc-lib-js 的 Transaction.UnspentOutput，字段含义：
//   - TxID: 引用的交易 ID（32 字节，小端序存储）
//   - Vout: 该输出在交易中的索引（对应 outputIndex）
//   - LockingScript: 锁定脚本 scriptPubKey，定义花费条件
//   - Satoshis: 该输出包含的聪数量（对应 amount 或 satoshis）
//   - SequenceNumber: 序列号，默认 0xFFFFFFFF 表示可立即花费
type UTXO struct {
	TxID           []byte
	Vout           uint32
	LockingScript  *script.Script
	Satoshis       uint64
	SequenceNumber uint32
}

// UTXOs 表示 *tbc.UTXO 的切片，用于批量处理 UTXO。
type UTXOs []*UTXO

// NodeJSON 返回用于 JSON 序列化/反序列化的包装类型，兼容节点格式（txid, vout, scriptPubKey, amount）。
//
// 与 bitcoind listunspent 等 RPC 返回格式兼容。
//
// Marshalling 示例:
//
//	bb, err := json.Marshal(utxo.NodeJSON())
//
// Unmarshalling 示例:
//
//	utxo := &tbc.UTXO{}
//	if err := json.Unmarshal(bb, utxo.NodeJSON()); err != nil {}
func (u *UTXO) NodeJSON() interface{} {
	return &nodeUTXOWrapper{UTXO: u}
}

// NodeJSON 返回 UTXOs 的节点格式包装，用于批量 JSON 序列化/反序列化。
func (u *UTXOs) NodeJSON() interface{} {
	return (*nodeUTXOsWrapper)(u)
}

// TxIDStr 将 TxID 编码为 hex 字符串返回。
func (u *UTXO) TxIDStr() string {
	return hex.EncodeToString(u.TxID)
}

// LockingScriptHexString 返回锁定脚本的 hex 字符串表示。
func (u *UTXO) LockingScriptHexString() string {
	return u.LockingScript.String()
}
