package tbc

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"

	"github.com/libsv/go-bk/crypto"

	"github.com/LoongYearMeta/tbc-lib-go/bscript"
	"github.com/LoongYearMeta/tbc-lib-go/encoding"
)

/*
General format of a Bitcoin transaction (inside a block)
--------------------------------------------------------
Field            Description                                                               Size

Version no	     currently 1	                                                           4 bytes

In-counter  	 positive integer VI = encoding.VarInt                                              1 - 9 bytes

list of Inputs	 the first input of the first transaction is also called "coinbase"        <in-counter>-many Inputs
                 (its content was ignored in earlier versions)

Out-counter    	 positive integer VI = encoding.VarInt                                              1 - 9 bytes

list of Outputs  the Outputs of the first transaction spend the mined                      <out-counter>-many Outputs
								 bitcoins for the block

lock_time        if non-zero and sequence numbers are < 0xFFFFFFFF: block height or        4 bytes
                 timestamp when transaction is final
--------------------------------------------------------
*/

// Tx wraps a bitcoin transaction
//
// DO NOT CHANGE ORDER - Optimised memory via malign
type Tx struct {
	Inputs   []*Input
	Outputs  []*Output
	Version  uint32
	LockTime uint32
}

// Txs a collection of *tbc.Tx.
type Txs []*Tx

// NewTx creates a new transaction object with default values.
// This matches the behavior of JavaScript's new Transaction() constructor.
func NewTx() *Tx {
	return &Tx{
		Version:  1,                  // Default transaction version for this library
		LockTime: 0,                  // Match JavaScript DEFAULT_NLOCKTIME
		Inputs:   make([]*Input, 0),  // Match JavaScript inputs = []
		Outputs:  make([]*Output, 0), // Match JavaScript outputs = []
	}
}

// NewTxFromString takes a toBytesHelper string representation of a bitcoin transaction
// and returns a Tx object.
func NewTxFromString(str string) (*Tx, error) {
	bb, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return NewTxFromBytes(bb)
}

// NewTxFromBytes takes an array of bytes, constructs a Tx and returns it.
// It parses the first complete transaction in the buffer. Trailing bytes (for example
// extra suffixes from explorers, tooling, or concatenated payloads) are ignored, matching
// tbc-lib-js Transaction / bitcore-style deserialization.
func NewTxFromBytes(b []byte) (*Tx, error) {
	tx, _, err := NewTxFromStream(b)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// NewTxFromStream takes an array of bytes and constructs a Tx from it, returning the Tx and the bytes used.
// Despite the name, this is not actually reading a stream in the true sense: it is a byte slice that contains
// many transactions one after another.
func NewTxFromStream(b []byte) (*Tx, int, error) {
	if len(b) < 10 {
		return nil, 0, ErrTxTooShort
	}

	var offset int
	t := Tx{
		Version: binary.LittleEndian.Uint32(b[offset:4]),
	}
	offset += 4

	inputCount, size := encoding.NewVarIntFromBytes(b[offset:])
	offset += size

	// create Inputs
	var i uint64
	var err error
	var input *Input
	for ; i < uint64(inputCount); i++ {
		input, size, err = newInputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += size
		t.addInput(input)
	}

	// create Outputs
	var outputCount encoding.VarInt
	var output *Output
	outputCount, size = encoding.NewVarIntFromBytes(b[offset:])
	offset += size
	for i = 0; i < uint64(outputCount); i++ {
		output, size, err = newOutputFromBytes(b[offset:])
		if err != nil {
			return nil, 0, err
		}
		offset += size
		t.AddOutput(output)
	}

	t.LockTime = binary.LittleEndian.Uint32(b[offset:])
	offset += 4

	return &t, offset, nil
}

// ReadFrom reads from the `io.Reader` into the `tbc.Tx`.
func (tx *Tx) ReadFrom(r io.Reader) (int64, error) {
	*tx = Tx{}
	var bytesRead int64

	version := make([]byte, 4)
	n, err := io.ReadFull(r, version)
	bytesRead += int64(n)
	if err != nil {
		return bytesRead, err
	}

	tx.Version = binary.LittleEndian.Uint32(version)

	var inputCount encoding.VarInt
	n64, err := inputCount.ReadFrom(r)
	bytesRead += n64
	if err != nil {
		return bytesRead, err
	}

	// create Inputs
	for i := uint64(0); i < uint64(inputCount); i++ {
		input := new(Input)
		n64, err = input.ReadFrom(r)
		bytesRead += n64
		if err != nil {
			return bytesRead, err
		}
		tx.Inputs = append(tx.Inputs, input)
	}

	var outputCount encoding.VarInt
	n64, err = outputCount.ReadFrom(r)
	bytesRead += n64
	if err != nil {
		return bytesRead, err
	}

	for i := uint64(0); i < uint64(outputCount); i++ {
		output := new(Output)
		n64, err = output.ReadFrom(r)
		bytesRead += n64
		if err != nil {
			return bytesRead, err
		}

		tx.Outputs = append(tx.Outputs, output)
	}

	locktime := make([]byte, 4)
	n, err = io.ReadFull(r, locktime)
	bytesRead += int64(n)
	if err != nil {
		return bytesRead, err
	}
	tx.LockTime = binary.LittleEndian.Uint32(locktime)

	return bytesRead, nil
}

// ReadFrom txs from a block in a `tbc.Txs`. This assumes a preceding varint detailing
// the total number of txs that the reader will provide.
func (tt *Txs) ReadFrom(r io.Reader) (int64, error) {
	var bytesRead int64

	var txCount encoding.VarInt
	n, err := txCount.ReadFrom(r)
	bytesRead += n
	if err != nil {
		return bytesRead, err
	}

	*tt = make([]*Tx, txCount)

	for i := uint64(0); i < uint64(txCount); i++ {
		tx := new(Tx)
		n, err := tx.ReadFrom(r)
		bytesRead += n
		if err != nil {
			return bytesRead, err
		}

		(*tt)[i] = tx
	}

	return bytesRead, nil
}

// HasDataOutputs returns true if the transaction has
// at least one data (OP_RETURN) output in it.
func (tx *Tx) HasDataOutputs() bool {
	for _, out := range tx.Outputs {
		if out.LockingScript.IsData() {
			return true
		}
	}
	return false
}

// InputIdx will return the input at the specified index.
//
// This will consume an overflow error and simply return nil if the input
// isn't found at the index.
func (tx *Tx) InputIdx(i int) *Input {
	if i > tx.InputCount()-1 {
		return nil
	}
	return tx.Inputs[i]
}

// OutputIdx will return the output at the specified index.
//
// This will consume an overflow error and simply return nil if the output
// isn't found at the index.
func (tx *Tx) OutputIdx(i int) *Output {
	if i > tx.OutputCount()-1 {
		return nil
	}
	return tx.Outputs[i]
}

// IsCoinbase determines if this transaction is a coinbase by
// checking if the tx input is a standard coinbase input.
func (tx *Tx) IsCoinbase() bool {
	if len(tx.Inputs) != 1 {
		return false
	}

	cbi := make([]byte, 32)

	if !bytes.Equal(tx.Inputs[0].PreviousTxID(), cbi) {
		return false
	}

	if tx.Inputs[0].PreviousTxOutIndex == DefaultSequenceNumber || tx.Inputs[0].SequenceNumber == DefaultSequenceNumber {
		return true
	}

	return false
}

// TxIDBytes returns the transaction ID of the transaction as bytes
// (which is also the transaction hash).
func (tx *Tx) TxIDBytes() []byte {
	if tx.Version >= 10 {
		return encoding.ReverseBytes(crypto.Sha256d(tx.newTxHeader()))
	}
	return encoding.ReverseBytes(crypto.Sha256d(tx.Bytes()))
}

// TxID returns the transaction ID of the transaction
// (which is also the transaction hash).
func (tx *Tx) TxID() string {
	return hex.EncodeToString(tx.TxIDBytes())
}

// newTxHeader builds the TBC v10+ transaction header used for txid computation.
// Mirrors tbc-lib-js Transaction.prototype.newTxHeader.
func (tx *Tx) newTxHeader() []byte {
	h := make([]byte, 0, 4+4+4+4+32+32+32)

	h = append(h, encoding.LittleEndianBytes(tx.Version, 4)...)

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)
	h = append(h, lt...)

	ic := make([]byte, 4)
	binary.LittleEndian.PutUint32(ic, uint32(len(tx.Inputs)))
	h = append(h, ic...)

	oc := make([]byte, 4)
	binary.LittleEndian.PutUint32(oc, uint32(len(tx.Outputs)))
	h = append(h, oc...)

	var inputBuf []byte
	var inputScriptBuf []byte
	for _, in := range tx.Inputs {
		inputBuf = append(inputBuf, encoding.ReverseBytes(in.previousTxID)...)
		inputBuf = append(inputBuf, encoding.LittleEndianBytes(in.PreviousTxOutIndex, 4)...)
		inputBuf = append(inputBuf, encoding.LittleEndianBytes(in.SequenceNumber, 4)...)

		var scriptBytes []byte
		if in.UnlockingScript != nil {
			scriptBytes = in.UnlockingScript.Bytes()
		}
		scriptHash := sha256.Sum256(scriptBytes)
		inputScriptBuf = append(inputScriptBuf, scriptHash[:]...)
	}
	inputHash := sha256.Sum256(inputBuf)
	h = append(h, inputHash[:]...)
	inputScriptHash := sha256.Sum256(inputScriptBuf)
	h = append(h, inputScriptHash[:]...)

	var outputBuf []byte
	for _, out := range tx.Outputs {
		satBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(satBytes, out.Satoshis)
		outputBuf = append(outputBuf, satBytes...)
		var lockBytes []byte
		if out.LockingScript != nil {
			lockBytes = out.LockingScript.Bytes()
		}
		outScriptHash := sha256.Sum256(lockBytes)
		outputBuf = append(outputBuf, outScriptHash[:]...)
	}
	outputHash := sha256.Sum256(outputBuf)
	h = append(h, outputHash[:]...)

	return h
}

// String encodes the transaction into a hex string.
func (tx *Tx) String() string {
	return hex.EncodeToString(tx.Bytes())
}

// IsValidTxID will check that the txid bytes are valid.
//
// A txid should be of 32 bytes length.
func IsValidTxID(txid []byte) bool {
	return len(txid) == 32
}

// Bytes encodes the transaction into a byte array.
// See https://chainquery.com/bitcoin-cli/decoderawtransaction
func (tx *Tx) Bytes() []byte {
	return tx.toBytesHelper(0, nil)
}

// BytesWithClearedInputs encodes the transaction into a byte array but clears its Inputs first.
// This is used when signing transactions.
func (tx *Tx) BytesWithClearedInputs(index int, lockingScript []byte) []byte {
	return tx.toBytesHelper(index, lockingScript)
}

// Clone returns a clone of the tx
func (tx *Tx) Clone() *Tx {
	// Ignore err as byte slice passed in is created from valid tx
	clone, _ := NewTxFromBytes(tx.Bytes())

	for i, input := range tx.Inputs {
		clone.Inputs[i].PreviousTxSatoshis = input.PreviousTxSatoshis
		clone.Inputs[i].PreviousTxScript = input.PreviousTxScript
	}

	return clone
}

// NodeJSON returns a wrapped *tbc.Tx for marshalling/unmarshalling into a node tx format.
//
// Marshalling usage example:
//
//	bb, err := json.Marshal(tx.NodeJSON())
//
// Unmarshalling usage example:
//
//	tx := tbc.NewTx()
//	if err := json.Unmarshal(bb, tx.NodeJSON()); err != nil {}
func (tx *Tx) NodeJSON() interface{} {
	return &nodeTxWrapper{Tx: tx}
}

// NodeJSON returns a wrapped tbc.Txs for marshalling/unmarshalling into a node tx format.
//
// Marshalling usage example:
//
//	bb, err := json.Marshal(txs.NodeJSON())
//
// Unmarshalling usage example:
//
//	var txs tbc.Txs
//	if err := json.Unmarshal(bb, txs.NodeJSON()); err != nil {}
func (tt *Txs) NodeJSON() interface{} {
	return (*nodeTxsWrapper)(tt)
}

func (tx *Tx) toBytesHelper(index int, lockingScript []byte) []byte {
	h := make([]byte, 0)

	h = append(h, encoding.LittleEndianBytes(tx.Version, 4)...)

	h = append(h, encoding.VarInt(uint64(len(tx.Inputs))).Bytes()...)

	for i, in := range tx.Inputs {
		s := in.Bytes(lockingScript != nil)
		if i == index && lockingScript != nil {
			h = append(h, encoding.VarInt(uint64(len(lockingScript))).Bytes()...)
			h = append(h, lockingScript...)
		} else {
			h = append(h, s...)
		}
	}

	h = append(h, encoding.VarInt(uint64(len(tx.Outputs))).Bytes()...)
	for _, out := range tx.Outputs {
		h = append(h, out.Bytes()...)
	}

	lt := make([]byte, 4)
	binary.LittleEndian.PutUint32(lt, tx.LockTime)

	return append(h, lt...)
}

// TxSize contains the size breakdown of a transaction
// including the breakdown of data bytes vs standard bytes.
// This information can be used when calculating fees.
type TxSize struct {
	// TotalBytes are the amount of bytes for the entire tx.
	TotalBytes uint64
	// TotalStdBytes are the amount of bytes for the tx minus the data bytes.
	TotalStdBytes uint64
	// TotalDataBytes is the size in bytes of the op_return / data outputs.
	TotalDataBytes uint64
}

// Size will return the size of tx in bytes.
func (tx *Tx) Size() int {
	return len(tx.Bytes())
}

// SizeWithTypes will return the size of tx in bytes
// and include the different data types (std/data/etc.).
func (tx *Tx) SizeWithTypes() *TxSize {
	totBytes := tx.Size()

	// calculate data outputs
	dataLen := 0
	for _, d := range tx.Outputs {
		if d.LockingScript.IsData() {
			dataLen += d.LockingScript.Len()
		}
	}
	return &TxSize{
		TotalBytes:     uint64(totBytes),
		TotalStdBytes:  uint64(totBytes - dataLen),
		TotalDataBytes: uint64(dataLen),
	}
}

// EstimateSize will return the size of tx in bytes and will add 107 bytes
// to the unlocking script of any unsigned inputs (only P2PKH for now) found
// to give a final size estimate of the tx size.
func (tx *Tx) EstimateSize() (int, error) {
	tempTx, err := tx.estimatedFinalTx()
	if err != nil {
		return 0, err
	}

	return tempTx.Size(), nil
}

// EstimateSizeWithTypes will return the size of tx in bytes, including the
// different data types (std/data/etc.), and will add 107 bytes to the unlocking
// script of any unsigned inputs (only P2PKH for now) found to give a final size
// estimate of the tx size.
func (tx *Tx) EstimateSizeWithTypes() (*TxSize, error) {
	tempTx, err := tx.estimatedFinalTx()
	if err != nil {
		return nil, err
	}

	return tempTx.SizeWithTypes(), nil
}

func (tx *Tx) estimatedFinalTx() (*Tx, error) {
	tempTx := tx.Clone()

	// 与 tbc-lib-js Transaction._estimateSize 一致：未签名时 generic Input 为 Script.empty()（仅长度前缀 varint 0），
	// feePerKb/change/seal 内先 _updateChangeOutput 再写入 FT 解锁脚本，故手续费按「空 scriptSig」估算。
	// 曾用数千字节占位会严重高估手续费，导致与 JS 的找零与 raw 不一致。
	dummyP2PKHUnlock, _ := hex.DecodeString("4830450221009c13cbcbb16f2cfedc7abf3a4af1c3fe77df1180c0e7eee30d9bcc53ebda39da02207b258005f1bc3cf9dffa06edb358d6db2bcfc87f50516fac8e3f4686fc2a03df412103107feff22788a1fc8357240bf450fd7bca4bd45d5f8bac63818c5a7b67b03876")

	for _, in := range tempTx.Inputs {
		if in.UnlockingScript != nil && in.UnlockingScript.Len() > 0 {
			continue
		}
		if in.PreviousTxScript != nil && in.PreviousTxScript.IsP2PKH() {
			in.UnlockingScript = bscript.NewFromBytes(dummyP2PKHUnlock)
			continue
		}
		in.UnlockingScript = bscript.NewFromBytes(nil)
	}
	return tempTx, nil
}

// TxFees is returned when CalculateFee is called and contains
// a breakdown of the fees including the total and the size breakdown of
// the tx in bytes.
type TxFees struct {
	// TotalFeePaid is the total amount of fees this tx will pay.
	TotalFeePaid uint64
	// StdFeePaid is the amount of fee to cover the standard inputs and outputs etc.
	StdFeePaid uint64
	// DataFeePaid is the amount of fee to cover the op_return data outputs.
	DataFeePaid uint64
}

// IsFeePaidEnough will calculate the fees that this transaction is paying
// including the individual fee types (std/data/etc.).
func (tx *Tx) IsFeePaidEnough(fees *FeeQuote) (bool, error) {
	expFeesPaid, err := tx.feesPaid(tx.SizeWithTypes(), fees)
	if err != nil {
		return false, err
	}
	totalInputSatoshis := tx.TotalInputSatoshis()
	totalOutputSatoshis := tx.TotalOutputSatoshis()

	if totalInputSatoshis < totalOutputSatoshis {
		return false, nil
	}

	actualFeePaid := totalInputSatoshis - totalOutputSatoshis
	return actualFeePaid >= expFeesPaid.TotalFeePaid, nil
}

// EstimateIsFeePaidEnough will calculate the fees that this transaction is paying
// including the individual fee types (std/data/etc.), and will add 107 bytes to the unlocking
// script of any unsigned inputs (only P2PKH for now) found to give a final size
// estimate of the tx size for fee calculation.
func (tx *Tx) EstimateIsFeePaidEnough(fees *FeeQuote) (bool, error) {
	tempTx, err := tx.estimatedFinalTx()
	if err != nil {
		return false, err
	}
	expFeesPaid, err := tempTx.feesPaid(tempTx.SizeWithTypes(), fees)
	if err != nil {
		return false, err
	}
	totalInputSatoshis := tempTx.TotalInputSatoshis()
	totalOutputSatoshis := tempTx.TotalOutputSatoshis()

	if totalInputSatoshis < totalOutputSatoshis {
		return false, nil
	}

	actualFeePaid := totalInputSatoshis - totalOutputSatoshis
	return actualFeePaid >= expFeesPaid.TotalFeePaid, nil
}

// EstimateFeesPaid will estimate how big the tx will be when finalised
// by estimating input unlocking scripts that have not yet been filled
// including the individual fee types (std/data/etc.).
func (tx *Tx) EstimateFeesPaid(fees *FeeQuote) (*TxFees, error) {
	size, err := tx.EstimateSizeWithTypes()
	if err != nil {
		return nil, err
	}
	return tx.feesPaid(size, fees)
}

func (tx *Tx) feesPaid(size *TxSize, fees *FeeQuote) (*TxFees, error) {
	// get fees
	stdFee, err := fees.Fee(FeeTypeStandard)
	if err != nil {
		return nil, err
	}
	dataFee, err := fees.Fee(FeeTypeData)
	if err != nil {
		return nil, err
	}

	resp := &TxFees{
		StdFeePaid:  size.TotalStdBytes * uint64(stdFee.MiningFee.Satoshis) / uint64(stdFee.MiningFee.Bytes),
		DataFeePaid: size.TotalDataBytes * uint64(dataFee.MiningFee.Satoshis) / uint64(dataFee.MiningFee.Bytes),
	}
	resp.TotalFeePaid = resp.StdFeePaid + resp.DataFeePaid
	return resp, nil

}

func (tx *Tx) estimateDeficit(fees *FeeQuote) (uint64, error) {
	totalInputSatoshis := tx.TotalInputSatoshis()
	totalOutputSatoshis := tx.TotalOutputSatoshis()

	expFeesPaid, err := tx.EstimateFeesPaid(fees)
	if err != nil {
		return 0, err
	}

	if totalInputSatoshis > totalOutputSatoshis+expFeesPaid.TotalFeePaid {
		return 0, nil
	}

	return totalOutputSatoshis + expFeesPaid.TotalFeePaid - totalInputSatoshis, nil
}

// Chainable methods for fluent API (matching JavaScript Transaction().from().to().change().sign() pattern)

// FromChain adds input(s) from UTXO(s) and returns the transaction for chaining.
// This matches JavaScript's Transaction.prototype.from() behavior.
// It accepts either a single UTXO or multiple UTXOs.
// Example: tx.FromChain(utxo1, utxo2, utxo3)
// If an error occurs, it panics (similar to JavaScript throwing an exception).
func (tx *Tx) FromChain(utxos ...*UTXO) *Tx {
	if err := tx.FromUTXOs(utxos...); err != nil {
		panic(err)
	}
	return tx
}

// FromStringChain adds an input from string parameters and returns the transaction for chaining.
// This is a convenience method for chaining when you have string-based UTXO information.
// Example: tx.FromStringChain(prevTxID, vout, lockingScript, satoshis)
func (tx *Tx) FromStringChain(prevTxID string, vout uint32, lockingScript string, satoshis uint64) *Tx {
	if err := tx.From(prevTxID, vout, lockingScript, satoshis); err != nil {
		panic(err)
	}
	return tx
}

// To adds an output to the specified address with the given amount and returns the transaction for chaining.
// This matches JavaScript's Transaction.prototype.to() behavior.
// It accepts either a single address/amount pair or a slice of address/amount pairs.
// If an error occurs, it panics (similar to JavaScript throwing an exception).
func (tx *Tx) To(address string, amount uint64) *Tx {
	if err := tx.PayToAddress(address, amount); err != nil {
		panic(err)
	}
	return tx
}

// ToOutput represents an output destination for ToMultiple method.
type ToOutput struct {
	Address string
	Amount  uint64
}

// ToMultiple adds multiple outputs from a slice of ToOutput and returns the transaction for chaining.
// This matches JavaScript's Transaction.prototype.to() behavior when passed an array.
// Example: tx.ToMultiple([]ToOutput{{Address: "addr1", Amount: 1000}, {Address: "addr2", Amount: 2000}})
func (tx *Tx) ToMultiple(outputs []ToOutput) *Tx {
	for _, output := range outputs {
		if err := tx.PayToAddress(output.Address, output.Amount); err != nil {
			panic(err)
		}
	}
	return tx
}

// Change sets the change address and calculates fees, then returns the transaction for chaining.
// This matches JavaScript's Transaction.prototype.change() behavior.
// If an error occurs, it panics (similar to JavaScript throwing an exception).
func (tx *Tx) Change(address string, feeQuote *FeeQuote) *Tx {
	if feeQuote == nil {
		feeQuote = NewFeeQuote()
	}
	if err := tx.ChangeToAddress(address, feeQuote); err != nil {
		panic(err)
	}
	return tx
}

// Sign signs the transaction using the provided UnlockerGetter and returns the transaction for chaining.
// This matches JavaScript's Transaction.prototype.sign() behavior.
// If an error occurs, it panics (similar to JavaScript throwing an exception).
func (tx *Tx) Sign(ctx context.Context, unlockerGetter UnlockerGetter) *Tx {
	if err := tx.FillAllInputs(ctx, unlockerGetter); err != nil {
		panic(err)
	}
	return tx
}
