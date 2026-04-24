package tbc

import (
	"fmt"

	"github.com/LoongYearMeta/tbc-lib-go/bscript"
)

// estimateSizeLikeJS 对齐 tbc-lib-js Transaction._estimateSize：
// version+locktime(8) + varint(nIn) + varint(nOut) + 各 input + 各 output(8+varint+script)。
// P2PKH 输入按 180 字节（BASE 40 + SCRIPT_MAX 140）；其它输入按 41 字节。
// extraChangeScript 非 nil 时额外计入一笔即将写入的找零输出（与 JS getFee 前带 0 sat 找零输出的估算一致）。
func estimateSizeLikeJS(tx *Tx, extraChangeScript *bscript.Script) int {
	nOut := len(tx.Outputs)
	if extraChangeScript != nil {
		nOut++
	}
	sz := 8
	sz += VarInt(uint64(len(tx.Inputs))).Length()
	sz += VarInt(uint64(nOut)).Length()
	for _, in := range tx.Inputs {
		if in.PreviousTxScript != nil && in.PreviousTxScript.IsP2PKH() {
			sz += 180
		} else {
			sz += 41
		}
	}
	for _, out := range tx.Outputs {
		l := out.LockingScript.Len()
		sz += 8 + VarInt(uint64(l)).Length() + l
	}
	if extraChangeScript != nil {
		l := extraChangeScript.Len()
		sz += 8 + VarInt(uint64(l)).Length() + l
	}
	return sz
}

// JSEstimateSize 与 tbc-lib-js Transaction._estimateSize 一致（仅当前已存在的 inputs/outputs）。
func (tx *Tx) JSEstimateSize() int {
	return estimateSizeLikeJS(tx, nil)
}

// AdjustImplicitFeeToTarget 通过调整最后一笔输出的 satoshis，使 (inputs−outputs) 等于 targetFee（sat）。
func (tx *Tx) AdjustImplicitFeeToTarget(targetFee int) error {
	if len(tx.Outputs) == 0 {
		return nil
	}
	in := tx.TotalInputSatoshis()
	out := tx.TotalOutputSatoshis()
	oldFee := int(in - out)
	delta := targetFee - oldFee
	if delta == 0 {
		return nil
	}
	last := len(tx.Outputs) - 1
	newSat := int64(tx.Outputs[last].Satoshis) - int64(delta)
	if newSat <= int64(DustLimit) {
		return fmt.Errorf("tbc: adjust implicit fee: change would be dust (delta=%d, targetFee=%d, oldFee=%d)", delta, targetFee, oldFee)
	}
	tx.Outputs[last].Satoshis = uint64(newSat)
	return nil
}
