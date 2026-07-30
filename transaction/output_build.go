package transaction

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/bits"

	"github.com/LoongYearMeta/tbc-lib-go/bip32"
	"github.com/LoongYearMeta/tbc-lib-go/crypto"
	"github.com/pkg/errors"

	"github.com/LoongYearMeta/tbc-lib-go/encoding"
	"github.com/LoongYearMeta/tbc-lib-go/script"
)

// newOutputFromBytes returns a transaction Output from the bytes provided
func newOutputFromBytes(bytes []byte) (*Output, int, error) {
	if len(bytes) < 8 {
		return nil, 0, fmt.Errorf("%w < 8", ErrOutputTooShort)
	}

	offset := 8
	l, size := encoding.NewVarIntFromBytes(bytes[offset:])
	offset += size

	totalLength := offset + int(l)

	if len(bytes) < totalLength {
		return nil, 0, fmt.Errorf("%w < 8 + script", ErrInputTooShort)
	}

	s := script.NewFromBytes(bytes[offset:totalLength])

	return &Output{
		Satoshis:      binary.LittleEndian.Uint64(bytes[0:8]),
		LockingScript: s,
	}, totalLength, nil
}

// TotalOutputSatoshis returns the total Satoshis outputted from the transaction.
func (tx *Tx) TotalOutputSatoshis() (total uint64) {
	for _, o := range tx.Outputs {
		total += o.Satoshis
	}
	return
}

// TotalOutputSatoshisChecked returns the output sum or ErrAmountOverflow.
func (tx *Tx) TotalOutputSatoshisChecked() (uint64, error) {
	var total uint64
	for _, output := range tx.Outputs {
		next, carry := bits.Add64(total, output.Satoshis, 0)
		if carry != 0 {
			return 0, ErrAmountOverflow
		}
		total = next
	}
	return total, nil
}

// AddP2PKHOutputFromPubKeyHashStr makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyHashStr(publicKeyHash string, satoshis uint64) error {
	s, err := script.NewP2PKHFromPubKeyHashStr(publicKeyHash)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromPubKeyBytes makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyBytes(publicKeyBytes []byte, satoshis uint64) error {
	s, err := script.NewP2PKHFromPubKeyBytes(publicKeyBytes)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromPubKeyStr makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromPubKeyStr(publicKey string, satoshis uint64) error {
	s, err := script.NewP2PKHFromPubKeyStr(publicKey)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromAddress makes an output to a PKH with a value.
func (tx *Tx) AddP2PKHOutputFromAddress(addr string, satoshis uint64) error {
	s, err := script.NewP2PKHFromAddress(addr)
	if err != nil {
		return err
	}

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddP2PKHOutputFromScript makes an output to a P2PKH script paid to the provided locking script with a value.
func (tx *Tx) AddP2PKHOutputFromScript(script *script.Script, satoshis uint64) error {
	if !script.IsP2PKH() {
		return errors.Wrapf(ErrInvalidScriptType, "'%s' is not a valid P2PKH script", script.ScriptType())
	}
	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: script,
	})
	return nil
}

// AddP2PKHOutputFromBip32ExtKey generated a random P2PKH output script from a provided *bip32.ExtendedKey,
// and add it to the receiving tx. The derviation path used is returned.
func (tx *Tx) AddP2PKHOutputFromBip32ExtKey(privKey *bip32.ExtendedKey, satoshis uint64) (string, error) {
	script, derivationPath, err := script.NewP2PKHFromBip32ExtKey(privKey)
	if err != nil {
		return "", err
	}

	tx.AddOutput(&Output{
		LockingScript: script,
		Satoshis:      satoshis,
	})
	return derivationPath, nil
}

// AddHashPuzzleOutput makes an output to a hash puzzle + PKH with a value.
func (tx *Tx) AddHashPuzzleOutput(secret, publicKeyHash string, satoshis uint64) error {
	publicKeyHashBytes, err := hex.DecodeString(publicKeyHash)
	if err != nil {
		return err
	}

	s := &script.Script{}

	_ = s.AppendOpcodes(script.OpHASH160)
	secretBytesHash := crypto.Hash160([]byte(secret))

	if err = s.AppendPushData(secretBytesHash); err != nil {
		return err
	}
	_ = s.AppendOpcodes(script.OpEQUALVERIFY, script.OpDUP, script.OpHASH160)

	if err = s.AppendPushData(publicKeyHashBytes); err != nil {
		return err
	}
	_ = s.AppendOpcodes(script.OpEQUALVERIFY, script.OpCHECKSIG)

	tx.AddOutput(&Output{
		Satoshis:      satoshis,
		LockingScript: s,
	})
	return nil
}

// AddOpReturnOutput creates a new Output with OP_FALSE OP_RETURN and then the data
// passed in encoded as hex.
func (tx *Tx) AddOpReturnOutput(data []byte) error {
	o, err := createOpReturnOutput([][]byte{data})
	if err != nil {
		return err
	}

	tx.AddOutput(o)
	return nil
}

// AddOpReturnPartsOutput creates a new Output with OP_FALSE OP_RETURN and then
// uses OP_PUSHDATA format to encode the multiple byte arrays passed in.
func (tx *Tx) AddOpReturnPartsOutput(data [][]byte) error {
	o, err := createOpReturnOutput(data)
	if err != nil {
		return err
	}
	tx.AddOutput(o)
	return nil
}

func createOpReturnOutput(data [][]byte) (*Output, error) {
	s := &script.Script{}

	_ = s.AppendOpcodes(script.OpFALSE, script.OpRETURN)
	if err := s.AppendPushDataArray(data); err != nil {
		return nil, err
	}

	return &Output{LockingScript: s}, nil
}

// OutputCount returns the number of transaction Inputs.
func (tx *Tx) OutputCount() int {
	return len(tx.Outputs)
}

// AddOutput adds a new output to the transaction.
func (tx *Tx) AddOutput(output *Output) {
	tx.Outputs = append(tx.Outputs, output)
}

// PayTo creates a new P2PKH output from a BitCoin address (base58)
// and the satoshis amount and adds that to the transaction.
func (tx *Tx) PayTo(script *script.Script, satoshis uint64) error {
	return tx.AddP2PKHOutputFromScript(script, satoshis)
}

// PayToAddress creates a new P2PKH output from a BitCoin address (base58)
// and the satoshis amount and adds that to the transaction.
func (tx *Tx) PayToAddress(addr string, satoshis uint64) error {
	return tx.AddP2PKHOutputFromAddress(addr, satoshis)
}
