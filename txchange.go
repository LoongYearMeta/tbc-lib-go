package tbc

import (
	"github.com/LoongYearMeta/tbc-lib-go/bscript"
)

const (
	// DustLimit is the current minimum txo output accepted by miners.
	DustLimit = 1
)

// ChangeToAddress calculates the amount of fees needed to cover the transaction
// and adds the leftover change in a new P2PKH output using the address provided.
func (tx *Tx) ChangeToAddress(addr string, f *FeeQuote) error {
	s, err := bscript.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	return tx.ChangeWithScript(s, f)
}

// ChangeWithScript calculates the amount of fees needed to cover the transaction
//  and adds the leftover change in a new output using the script provided.
func (tx *Tx) ChangeWithScript(s *bscript.Script, f *FeeQuote) error {
	if _, _, err := tx.change(f, &changeOutput{
		lockingScript: s,
		newOutput:     true,
	}); err != nil {
		return err
	}
	return nil
}

// ChangeToExistingOutput will calculate fees and add them to an output at the index specified (0 based).
// If an invalid index is supplied and error is returned.
func (tx *Tx) ChangeToExistingOutput(index uint, f *FeeQuote) error {
	if int(index) > tx.OutputCount()-1 {
		return ErrOutputNoExist
	}
	available, hasChange, err := tx.change(f, nil)
	if err != nil {
		return err
	}
	if hasChange {
		tx.Outputs[index].Satoshis += available
	}
	return nil
}

type changeOutput struct {
	lockingScript *bscript.Script
	newOutput     bool
}

// change 手续费与 tbc-lib-js getFee/_estimateFee 一致：
// txFees = ceil(estimateSizeLikeJS * feePerKb / 1000)，feePerKb = MiningFee.Satoshis*1000/MiningFee.Bytes；
// 新建找零时 estimate 含即将写入的找零输出脚本（与 JS 临时 0 sat 找零等价）。
func (tx *Tx) change(f *FeeQuote, output *changeOutput) (uint64, bool, error) {
	inputAmount := tx.TotalInputSatoshis()
	outputAmount := tx.TotalOutputSatoshis()
	if inputAmount < outputAmount {
		return 0, false, ErrInsufficientInputs
	}

	available := inputAmount - outputAmount

	stdFee, err := f.Fee(FeeTypeStandard)
	if err != nil {
		return 0, false, err
	}

	var est int
	if output != nil && output.newOutput {
		est = estimateSizeLikeJS(tx, output.lockingScript)
	} else {
		est = estimateSizeLikeJS(tx, nil)
	}
	txFees := CeilMiningFeeFromEstimatedBytes(est, stdFee.MiningFee)

	if available <= txFees || available-txFees <= uint64(DustLimit) {
		return 0, false, nil
	}

	available -= txFees
	if output != nil && output.newOutput {
		tx.AddOutput(&Output{Satoshis: available, LockingScript: output.lockingScript})
	}
	return available, true, nil
}
