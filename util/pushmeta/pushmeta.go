// Package pushmeta implements TBC's OP_PUSH_META helpers used by script
// interpreter and covenant templates. Matches tbc-lib-js
// lib/script/interpreter.js case 6 (OP_PUSH_META).
package pushmeta

import (
	"encoding/binary"

	"github.com/LoongYearMeta/tbc-lib-go/transaction"
)

// CurrentInputOutpointBytes returns the 40 bytes pushed by OP_6 OP_PUSH_META:
// previous txid in wire (little-endian) order, then vout and sequence as uint32 LE.
// Matches tbc-lib-js lib/script/interpreter.js case 6 (OP_PUSH_META).
func CurrentInputOutpointBytes(tx *transaction.Tx, inputIdx int) []byte {
	if tx == nil || inputIdx < 0 || inputIdx >= len(tx.Inputs) {
		return nil
	}
	in := tx.Inputs[inputIdx]
	prevTxID := make([]byte, len(in.PreviousTxID()))
	copy(prevTxID, in.PreviousTxID())
	for i := 0; i < len(prevTxID)/2; i++ {
		prevTxID[i], prevTxID[len(prevTxID)-1-i] = prevTxID[len(prevTxID)-1-i], prevTxID[i]
	}
	out := make([]byte, 0, 40)
	out = append(out, prevTxID...)
	oi := make([]byte, 4)
	binary.LittleEndian.PutUint32(oi, in.PreviousTxOutIndex)
	out = append(out, oi...)
	seq := make([]byte, 4)
	binary.LittleEndian.PutUint32(seq, in.SequenceNumber)
	out = append(out, seq...)
	return out
}
