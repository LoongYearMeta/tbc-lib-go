package tbc_test

// Auto-generated compile-time completeness gate for the tbc package facade.
// Each historical `tbc.X` symbol is referenced here. If any is deleted from the
// facade (or a subpackage sub-dependency drops one), this file fails to compile.
// The test body is deliberately a `var _ = func() {...}` — no runtime behavior.

import tbc "github.com/LoongYearMeta/tbc-lib-go"

var _ = func() {
	var _ tbc.Block
	var _ tbc.BlockHeader
	var _ tbc.ECIES
	var _ tbc.ECIESOptions
	var _ tbc.Fee
	var _ tbc.FeeQuote
	var _ tbc.FeeQuotes
	var _ tbc.FeeType
	var _ tbc.FeeUnit
	var _ tbc.Input
	var _ tbc.MerkleBlock
	var _ tbc.Message
	var _ tbc.Network
	var _ tbc.Output
	var _ tbc.ToOutput
	var _ tbc.Tx
	var _ tbc.TxFees
	var _ tbc.TxSize
	var _ tbc.Txs
	var _ tbc.UTXO
	var _ tbc.UTXOGetterFunc
	var _ tbc.UTXOs
	var _ tbc.Unlocker
	var _ tbc.UnlockerGetter
	var _ tbc.UnlockerParams
	var _ tbc.VarInt
	_ = tbc.AddNetwork
	_ = tbc.CeilMiningFeeFromEstimatedBytes
	_ = tbc.ClampDifficulty
	_ = tbc.CurrentInputOutpointBytes
	_ = tbc.DecryptWith
	_ = tbc.EncryptFor
	_ = tbc.FromJSON
	_ = tbc.FromObject
	_ = tbc.FromString
	_ = tbc.GetNetwork
	_ = tbc.IsHexString
	_ = tbc.IsValidTxID
	_ = tbc.LittleEndianBytes
	_ = tbc.NewBlock
	_ = tbc.NewBlockFromBytes
	_ = tbc.NewBlockFromRawBlock
	_ = tbc.NewBlockFromString
	_ = tbc.NewBlockHeader
	_ = tbc.NewBlockHeaderFromBytes
	_ = tbc.NewBlockHeaderFromRawBlock
	_ = tbc.NewBlockHeaderFromString
	_ = tbc.NewECIES
	_ = tbc.NewFeeQuote
	_ = tbc.NewFeeQuotes
	_ = tbc.NewMerkleBlock
	_ = tbc.NewMerkleBlockFromBytes
	_ = tbc.NewMerkleBlockFromString
	_ = tbc.NewMessageFromBytes
	_ = tbc.NewMessageFromString
	_ = tbc.NewTx
	_ = tbc.NewTxFromBytes
	_ = tbc.NewTxFromStream
	_ = tbc.NewTxFromString
	_ = tbc.NewVarIntFromBytes
	_ = tbc.RandomBytes
	_ = tbc.ReadVarBytes
	_ = tbc.RemoveNetwork
	_ = tbc.ReverseBytes
	_ = tbc.SignMessage
	_ = tbc.VerifyMessage
	_ = tbc.VerifyMessageWithAddress
	_ = tbc.VerifyMessageWithPubKey
	_ = tbc.WriteVarBytes
	_ = tbc.DefaultSequenceNumber

	// ---- var (...) / const (...) block completeness gate ----

	// networks vars
	_ = tbc.Livenet
	_ = tbc.Testnet
	_ = tbc.Regtest
	_ = tbc.STN
	_ = tbc.AddNetwork
	_ = tbc.GetNetwork
	_ = tbc.RemoveNetwork
	_ = tbc.DefaultNetwork

	// encoding vars
	_ = tbc.NewVarIntFromBytes
	_ = tbc.ReverseBytes
	_ = tbc.LittleEndianBytes
	_ = tbc.ReadVarBytes
	_ = tbc.WriteVarBytes
	_ = tbc.IsHexString

	// transaction constructors
	_ = tbc.NewTx
	_ = tbc.NewTxFromString
	_ = tbc.NewTxFromBytes
	_ = tbc.NewTxFromStream
	_ = tbc.NewFeeQuote
	_ = tbc.NewFeeQuotes
	_ = tbc.CeilMiningFeeFromEstimatedBytes
	_ = tbc.IsValidTxID

	// transaction constants
	_ = tbc.DustLimit
	_ = tbc.MaxTxInSequenceNum
	_ = tbc.MaxPrevOutIndex
	_ = tbc.SequenceLockTimeDisabled
	_ = tbc.SequenceLockTimeIsSeconds
	_ = tbc.SequenceLockTimeMask
	_ = tbc.FeeTypeStandard
	_ = tbc.FeeTypeData
	_ = tbc.DefaultSequenceNumber

	// transaction sentinel errors
	_ = tbc.ErrInvalidTxID
	_ = tbc.ErrTxNil
	_ = tbc.ErrTxTooShort
	_ = tbc.ErrNLockTimeLength
	_ = tbc.ErrEmptyValues
	_ = tbc.ErrUnsupportedScript
	_ = tbc.ErrInvalidScriptType
	_ = tbc.ErrNoUnlocker
	_ = tbc.ErrInputNoExist
	_ = tbc.ErrInputTooShort
	_ = tbc.ErrOutputNoExist
	_ = tbc.ErrOutputTooShort
	_ = tbc.ErrInsufficientInputs
	_ = tbc.ErrEmptyPreviousTxID
	_ = tbc.ErrEmptyPreviousTxScript
	_ = tbc.ErrFeeQuotesNotInit
	_ = tbc.ErrMinerNoQuotes
	_ = tbc.ErrFeeTypeNotFound
	_ = tbc.ErrFeeQuoteNotInit
	_ = tbc.ErrUnknownFeeType
	_ = tbc.ErrNoUTXO
	_ = tbc.ErrInsufficientFunds

	// block constructors
	_ = tbc.NewBlock
	_ = tbc.NewBlockFromBytes
	_ = tbc.NewBlockFromString
	_ = tbc.NewBlockFromRawBlock
	_ = tbc.NewBlockHeader
	_ = tbc.NewBlockHeaderFromBytes
	_ = tbc.NewBlockHeaderFromString
	_ = tbc.NewBlockHeaderFromRawBlock
	_ = tbc.NewMerkleBlock
	_ = tbc.NewMerkleBlockFromBytes
	_ = tbc.NewMerkleBlockFromString
	_ = tbc.ClampDifficulty

	// block constants
	_ = tbc.MaxBlockSize
	_ = tbc.BlockStartOffset
	_ = tbc.GenesisBits
	_ = tbc.BlockHeaderSize
	_ = tbc.BlockHeaderStartOffset
	_ = tbc.MaxBlockTimeOffsetSeconds

	// block vars
	_ = tbc.NullHash
	_ = tbc.LargestHash

	// block sentinel errors
	_ = tbc.ErrBlockNil
	_ = tbc.ErrBlockTooShort
	_ = tbc.ErrBlockTxCountTooLarge
	_ = tbc.ErrBlockHeaderNil
	_ = tbc.ErrBlockHeaderTooShort
	_ = tbc.ErrMerkleBlockNil
	_ = tbc.ErrInvalidMerkleTree

	// message vars
	_ = tbc.NewMessageFromString
	_ = tbc.NewMessageFromBytes
	_ = tbc.FromString
	_ = tbc.FromJSON
	_ = tbc.FromObject
	_ = tbc.SignMessage
	_ = tbc.VerifyMessage
	_ = tbc.VerifyMessageWithPubKey
	_ = tbc.VerifyMessageWithAddress

	// ecies vars
	_ = tbc.NewECIES
	_ = tbc.EncryptFor
	_ = tbc.DecryptWith
	_ = tbc.RandomBytes

	// util/pushmeta vars
	_ = tbc.CurrentInputOutpointBytes
}
