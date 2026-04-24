package tbc

import "github.com/LoongYearMeta/tbc-lib-go/transaction"

// CurrentInputOutpointBytes returns the 40 bytes pushed by OP_6 OP_PUSH_META:
// previous txid in wire (little-endian) order, then vout and sequence as uint32 LE.
// Matches tbc-lib-js lib/script/interpreter.js case 6 (OP_PUSH_META).
//
// Delegates to transaction.CurrentInputOutpointBytes; this root-package wrapper
// preserves the historical tbc.CurrentInputOutpointBytes facade.
func CurrentInputOutpointBytes(tx *Tx, inputIdx int) []byte {
	return transaction.CurrentInputOutpointBytes(tx, inputIdx)
}
