// Package tbc is the facade for the tbc-lib-go library. All historical tbc.X
// symbols are re-exported here from their new home in subpackages. Prefer
// importing the subpackage directly for new code, but this facade preserves
// backwards compatibility for every exported symbol that existed prior to
// the JS-style layout refactor.
//
// Design reference: docs/superpowers/specs/2026-04-24-js-style-layout-design.md
package tbc

import (
	"github.com/LoongYearMeta/tbc-lib-go/block"
	"github.com/LoongYearMeta/tbc-lib-go/ecies"
	"github.com/LoongYearMeta/tbc-lib-go/encoding"
	"github.com/LoongYearMeta/tbc-lib-go/message"
	"github.com/LoongYearMeta/tbc-lib-go/networks"
	"github.com/LoongYearMeta/tbc-lib-go/transaction"
	"github.com/LoongYearMeta/tbc-lib-go/util/pushmeta"
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
	CeilFeeForBytes                 = transaction.CeilFeeForBytes
	CeilMiningFeeFromEstimatedBytes = transaction.CeilMiningFeeFromEstimatedBytes
	IsValidTxID                     = transaction.IsValidTxID
)

// constants
const (
	NodeDustLimit             = transaction.NodeDustLimit
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
	ErrAmountOverflow        = transaction.ErrAmountOverflow
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
	ErrInvalidFee            = transaction.ErrInvalidFee
	ErrFeeOverflow           = transaction.ErrFeeOverflow
	ErrNoUTXO                = transaction.ErrNoUTXO
	ErrInsufficientFunds     = transaction.ErrInsufficientFunds
)

// ======== block ========

type (
	Block       = block.Block
	BlockHeader = block.BlockHeader
	MerkleBlock = block.MerkleBlock
)

// constructors
var (
	NewBlock                   = block.NewBlock
	NewBlockFromBytes          = block.NewBlockFromBytes
	NewBlockFromString         = block.NewBlockFromString
	NewBlockFromRawBlock       = block.NewBlockFromRawBlock
	NewBlockHeader             = block.NewBlockHeader
	NewBlockHeaderFromBytes    = block.NewBlockHeaderFromBytes
	NewBlockHeaderFromString   = block.NewBlockHeaderFromString
	NewBlockHeaderFromRawBlock = block.NewBlockHeaderFromRawBlock
	NewMerkleBlock             = block.NewMerkleBlock
	NewMerkleBlockFromBytes    = block.NewMerkleBlockFromBytes
	NewMerkleBlockFromString   = block.NewMerkleBlockFromString
	ClampDifficulty            = block.ClampDifficulty
)

// block constants
const (
	MaxBlockSize              = block.MaxBlockSize
	BlockStartOffset          = block.BlockStartOffset
	GenesisBits               = block.GenesisBits
	BlockHeaderSize           = block.BlockHeaderSize
	BlockHeaderStartOffset    = block.BlockHeaderStartOffset
	MaxBlockTimeOffsetSeconds = block.MaxBlockTimeOffsetSeconds
)

// block vars
var (
	NullHash    = block.NullHash
	LargestHash = block.LargestHash
)

// block sentinels
var (
	ErrBlockNil             = block.ErrBlockNil
	ErrBlockTooShort        = block.ErrBlockTooShort
	ErrBlockTxCountTooLarge = block.ErrBlockTxCountTooLarge
	ErrBlockHeaderNil       = block.ErrBlockHeaderNil
	ErrBlockHeaderTooShort  = block.ErrBlockHeaderTooShort
	ErrMerkleBlockNil       = block.ErrMerkleBlockNil
	ErrInvalidMerkleTree    = block.ErrInvalidMerkleTree
)

// ======== message ========

type Message = message.Message

var (
	NewMessageFromString     = message.NewMessageFromString
	NewMessageFromBytes      = message.NewMessageFromBytes
	FromString               = message.FromString
	FromJSON                 = message.FromJSON
	FromObject               = message.FromObject
	SignMessage              = message.SignMessage
	VerifyMessage            = message.VerifyMessage
	VerifyMessageWithPubKey  = message.VerifyMessageWithPubKey
	VerifyMessageWithAddress = message.VerifyMessageWithAddress
)

// ======== ecies ========

type (
	ECIES        = ecies.ECIES
	ECIESOptions = ecies.ECIESOptions
)

var (
	NewECIES    = ecies.NewECIES
	EncryptFor  = ecies.EncryptFor
	DecryptWith = ecies.DecryptWith
	RandomBytes = ecies.RandomBytes
)

// ======== util/pushmeta ========

var (
	CurrentInputOutpointBytes = pushmeta.CurrentInputOutpointBytes
)
