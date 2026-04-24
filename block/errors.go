package block

import "github.com/pkg/errors"

// General errors reported by block / merkleblock parsing.
var (
	ErrBlockNil             = errors.New("block is nil")
	ErrBlockTooShort        = errors.New("too short to be a block")
	ErrBlockTxCountTooLarge = errors.New("block tx count too large")
	ErrBlockHeaderNil       = errors.New("block header is nil")
	ErrBlockHeaderTooShort  = errors.New("too short to be a block header")
	ErrMerkleBlockNil       = errors.New("merkleblock is nil")
	ErrInvalidMerkleTree    = errors.New("invalid merkle tree")
)
