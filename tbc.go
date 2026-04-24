// Package tbc is the facade for the tbc-lib-go library. All historical tbc.X
// symbols are re-exported here from their new home in subpackages. Prefer
// importing the subpackage directly for new code, but this facade preserves
// backwards compatibility for every exported symbol that existed prior to
// the JS-style layout refactor.
//
// Design reference: docs/superpowers/specs/2026-04-24-js-style-layout-design.md
package tbc

import (
	"github.com/LoongYearMeta/tbc-lib-go/encoding"
	"github.com/LoongYearMeta/tbc-lib-go/networks"
	"github.com/LoongYearMeta/tbc-lib-go/transaction"
)

// ======== networks ========

type Network = networks.Network

var (
	Livenet = networks.Livenet
	Testnet = networks.Testnet
	Regtest = networks.Regtest
	STN     = networks.STN

	AddNetwork    = networks.AddNetwork
	GetNetwork    = networks.GetNetwork
	RemoveNetwork = networks.RemoveNetwork
)

// DefaultNetwork is intentionally defined HERE rather than re-exported from
// networks/. A "var = pkg.Var" re-export would copy the value, so
// `tbc.DefaultNetwork = tbc.Testnet` would only mutate the root-package copy
// and silently drift from networks/. Keeping it in the root package preserves
// the original single-source-of-truth semantics.
var DefaultNetwork = Livenet

// ======== encoding ========

type VarInt = encoding.VarInt

var (
	NewVarIntFromBytes = encoding.NewVarIntFromBytes
	ReverseBytes       = encoding.ReverseBytes
	LittleEndianBytes  = encoding.LittleEndianBytes
	ReadVarBytes       = encoding.ReadVarBytes
	WriteVarBytes      = encoding.WriteVarBytes
	IsHexString        = encoding.IsHexString
)

// ======== transaction ========

// types
type (
	Tx             = transaction.Tx
	Txs            = transaction.Txs
	Input          = transaction.Input
	Output         = transaction.Output
	UTXO           = transaction.UTXO
	UTXOs          = transaction.UTXOs
	FeeQuote       = transaction.FeeQuote
	FeeQuotes      = transaction.FeeQuotes
	Fee            = transaction.Fee
	FeeUnit        = transaction.FeeUnit
	FeeType        = transaction.FeeType
	TxFees         = transaction.TxFees
	TxSize         = transaction.TxSize
	Unlocker       = transaction.Unlocker
	UnlockerGetter = transaction.UnlockerGetter
	UnlockerParams = transaction.UnlockerParams
	ToOutput       = transaction.ToOutput
	UTXOGetterFunc = transaction.UTXOGetterFunc
)

// constructors
var (
	NewTx                           = transaction.NewTx
	NewTxFromString                 = transaction.NewTxFromString
	NewTxFromBytes                  = transaction.NewTxFromBytes
	NewTxFromStream                 = transaction.NewTxFromStream
	NewFeeQuote                     = transaction.NewFeeQuote
	NewFeeQuotes                    = transaction.NewFeeQuotes
	CeilMiningFeeFromEstimatedBytes = transaction.CeilMiningFeeFromEstimatedBytes
	IsValidTxID                     = transaction.IsValidTxID
)

// constants
const (
	DustLimit                 = transaction.DustLimit
	MaxTxInSequenceNum        = transaction.MaxTxInSequenceNum
	MaxPrevOutIndex           = transaction.MaxPrevOutIndex
	SequenceLockTimeDisabled  = transaction.SequenceLockTimeDisabled
	SequenceLockTimeIsSeconds = transaction.SequenceLockTimeIsSeconds
	SequenceLockTimeMask      = transaction.SequenceLockTimeMask
	FeeTypeStandard           = transaction.FeeTypeStandard
	FeeTypeData               = transaction.FeeTypeData
	DefaultSequenceNumber     = transaction.DefaultSequenceNumber
)

// sentinels
var (
	ErrInvalidTxID           = transaction.ErrInvalidTxID
	ErrTxNil                 = transaction.ErrTxNil
	ErrTxTooShort            = transaction.ErrTxTooShort
	ErrNLockTimeLength       = transaction.ErrNLockTimeLength
	ErrEmptyValues           = transaction.ErrEmptyValues
	ErrUnsupportedScript     = transaction.ErrUnsupportedScript
	ErrInvalidScriptType     = transaction.ErrInvalidScriptType
	ErrNoUnlocker            = transaction.ErrNoUnlocker
	ErrInputNoExist          = transaction.ErrInputNoExist
	ErrInputTooShort         = transaction.ErrInputTooShort
	ErrOutputNoExist         = transaction.ErrOutputNoExist
	ErrOutputTooShort        = transaction.ErrOutputTooShort
	ErrInsufficientInputs    = transaction.ErrInsufficientInputs
	ErrEmptyPreviousTxID     = transaction.ErrEmptyPreviousTxID
	ErrEmptyPreviousTxScript = transaction.ErrEmptyPreviousTxScript
	ErrFeeQuotesNotInit      = transaction.ErrFeeQuotesNotInit
	ErrMinerNoQuotes         = transaction.ErrMinerNoQuotes
	ErrFeeTypeNotFound       = transaction.ErrFeeTypeNotFound
	ErrFeeQuoteNotInit       = transaction.ErrFeeQuoteNotInit
	ErrUnknownFeeType        = transaction.ErrUnknownFeeType
	ErrNoUTXO                = transaction.ErrNoUTXO
	ErrInsufficientFunds     = transaction.ErrInsufficientFunds
	// NOTE: Block-related sentinel errors (ErrBlockNil, ErrMerkleBlockNil, etc.)
	// are currently declared in transaction/errors.go alongside these tx errors.
	// They will be moved to block/errors.go in Phase 5. For now, re-export via transaction.
	ErrBlockNil             = transaction.ErrBlockNil
	ErrBlockTooShort        = transaction.ErrBlockTooShort
	ErrBlockTxCountTooLarge = transaction.ErrBlockTxCountTooLarge
	ErrBlockHeaderNil       = transaction.ErrBlockHeaderNil
	ErrBlockHeaderTooShort  = transaction.ErrBlockHeaderTooShort
	ErrMerkleBlockNil       = transaction.ErrMerkleBlockNil
	ErrInvalidMerkleTree    = transaction.ErrInvalidMerkleTree
)
