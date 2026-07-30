package transaction

import (
	"fmt"
	"math/bits"

	"github.com/LoongYearMeta/tbc-lib-go/encoding"
	"github.com/LoongYearMeta/tbc-lib-go/script"
)

// estimateSizeLikeJS 对齐 tbc-lib-js Transaction._estimateSize：
// version+locktime(8) + varint(nIn) + varint(nOut) + 各 input + 各 output(8+varint+script)。
// P2PKH 输入按 180 字节（BASE 40 + SCRIPT_MAX 140）；其它输入按 41 字节。
// extraChangeScript 非 nil 时额外计入一笔即将写入的找零输出（与 JS getFee 前带 0 sat 找零输出的估算一致）。
func estimateSizeLikeJS(tx *Tx, extraChangeScript *script.Script) int {
	nOut := len(tx.Outputs)
	if extraChangeScript != nil {
		nOut++
	}
	sz := 8
	sz += encoding.VarInt(uint64(len(tx.Inputs))).Length()
	sz += encoding.VarInt(uint64(nOut)).Length()
	for _, in := range tx.Inputs {
		if in.PreviousTxScript != nil && in.PreviousTxScript.IsP2PKH() {
			sz += 180
		} else {
			sz += 41
		}
	}
	for _, out := range tx.Outputs {
		l := out.LockingScript.Len()
		sz += 8 + encoding.VarInt(uint64(l)).Length() + l
	}
	if extraChangeScript != nil {
		l := extraChangeScript.Len()
		sz += 8 + encoding.VarInt(uint64(l)).Length() + l
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
	if targetFee < 0 {
		return ErrInvalidFee
	}

	in, err := tx.TotalInputSatoshisChecked()
	if err != nil {
		return err
	}
	out, err := tx.TotalOutputSatoshisChecked()
	if err != nil {
		return err
	}
	if in < out {
		return ErrInsufficientInputs
	}

	oldFee := in - out
	target := uint64(targetFee)
	if target == oldFee {
		return nil
	}

	last := len(tx.Outputs) - 1
	current := tx.Outputs[last].Satoshis
	var newSat uint64
	if target > oldFee {
		delta := target - oldFee
		if current < delta {
			return ErrInsufficientInputs
		}
		newSat = current - delta
	} else {
		delta := oldFee - target
		var carry uint64
		newSat, carry = bits.Add64(current, delta, 0)
		if carry != 0 {
			return ErrAmountOverflow
		}
	}
	if newSat < DustLimit {
		return fmt.Errorf("tbc: adjust implicit fee: change would be dust (targetFee=%d, oldFee=%d)", targetFee, oldFee)
	}
	tx.Outputs[last].Satoshis = newSat
	return nil
}
