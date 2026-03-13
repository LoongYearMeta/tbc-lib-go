package bt

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/libsv/go-bk/crypto"
)

const (
	// MaxBlockSize 最大区块体积（字节），与 tbc-lib-js 保持一致。
	MaxBlockSize = 128000000

	// BlockStartOffset 原始区块载荷中，区块数据起始偏移（跳过前 8 字节）。
	BlockStartOffset = 8
)

var (
	// NullHash 32 字节零哈希，用于创世块等场景。
	NullHash = make([]byte, 32)
)

// Block 表示 TBC 区块，包含区块头和交易列表。
//
// 对应 tbc-lib-js 的 Block 类，详见 docs/block.md。
type Block struct {
	Header       *BlockHeader // 区块头
	Transactions []*Tx        // 交易列表
}

// NewBlock returns an empty block.
func NewBlock() *Block {
	return &Block{
		Header:       NewBlockHeader(),
		Transactions: make([]*Tx, 0),
	}
}

// NewBlockFromBytes decodes a block from bytes.
func NewBlockFromBytes(b []byte) (*Block, error) {
	if len(b) < BlockHeaderSize+1 {
		return nil, ErrBlockTooShort
	}
	blk := NewBlock()
	if _, err := blk.ReadFromBytes(b); err != nil {
		return nil, err
	}
	return blk, nil
}

// NewBlockFromString decodes a block from hex.
func NewBlockFromString(str string) (*Block, error) {
	b, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return NewBlockFromBytes(b)
}

// NewBlockFromRawBlock decodes block from raw payload (skip first 8 bytes).
func NewBlockFromRawBlock(raw []byte) (*Block, error) {
	if len(raw) < BlockStartOffset+BlockHeaderSize+1 {
		return nil, ErrBlockTooShort
	}
	return NewBlockFromBytes(raw[BlockStartOffset:])
}

// ReadFrom reads a block from io.Reader.
func (b *Block) ReadFrom(r io.Reader) (int64, error) {
	if b == nil {
		return 0, ErrBlockNil
	}

	*b = Block{}

	header := NewBlockHeader()
	n, err := header.ReadFrom(r)
	if err != nil {
		return n, err
	}

	var txs Txs
	nTx, err := txs.ReadFrom(r)
	if err != nil {
		return n + nTx, err
	}

	b.Header = header
	b.Transactions = txs
	return n + nTx, nil
}

// ReadFromBytes decodes block from bytes and returns consumed size.
func (b *Block) ReadFromBytes(bb []byte) (int, error) {
	if b == nil {
		return 0, ErrBlockNil
	}
	if len(bb) < BlockHeaderSize+1 {
		return 0, ErrBlockTooShort
	}

	header, err := NewBlockHeaderFromBytes(bb)
	if err != nil {
		return 0, err
	}
	offset := BlockHeaderSize

	txCount, size := NewVarIntFromBytes(bb[offset:])
	offset += size
	if txCount > VarInt(MaxBlockSize) {
		return 0, ErrBlockTxCountTooLarge
	}

	transactions := make([]*Tx, 0, txCount)
	for i := uint64(0); i < uint64(txCount); i++ {
		tx, used, txErr := NewTxFromStream(bb[offset:])
		if txErr != nil {
			return 0, txErr
		}
		offset += used
		if offset > len(bb) {
			return 0, ErrBlockTooShort
		}
		transactions = append(transactions, tx)
	}

	b.Header = header
	b.Transactions = transactions
	return offset, nil
}

// WriteTo writes a block to io.Writer.
func (b *Block) WriteTo(w io.Writer) (int64, error) {
	if b == nil {
		return 0, ErrBlockNil
	}
	n, err := w.Write(b.Bytes())
	return int64(n), err
}

// Bytes encodes block into wire format.
func (b *Block) Bytes() []byte {
	if b == nil || b.Header == nil {
		return nil
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(b.Header.Bytes())
	buf.Write(VarInt(uint64(len(b.Transactions))).Bytes())
	for _, tx := range b.Transactions {
		if tx == nil {
			continue
		}
		buf.Write(tx.Bytes())
	}
	return buf.Bytes()
}

// String returns hex-encoded block bytes.
func (b *Block) String() string {
	return hex.EncodeToString(b.Bytes())
}

// HashBytes returns little-endian hash buffer of header.
func (b *Block) HashBytes() []byte {
	if b == nil || b.Header == nil {
		return nil
	}
	return b.Header.HashBytes()
}

// ID returns block hash in big-endian hex.
func (b *Block) ID() string {
	if b == nil || b.Header == nil {
		return ""
	}
	return b.Header.ID()
}

// GetTransactionHashes returns little-endian tx hashes.
func (b *Block) GetTransactionHashes() [][]byte {
	if b == nil {
		return nil
	}
	if len(b.Transactions) == 0 {
		return [][]byte{append([]byte(nil), NullHash...)}
	}

	hashes := make([][]byte, 0, len(b.Transactions))
	for _, tx := range b.Transactions {
		if tx == nil {
			hashes = append(hashes, append([]byte(nil), NullHash...))
			continue
		}
		hashes = append(hashes, crypto.Sha256d(tx.Bytes()))
	}
	return hashes
}

// GetMerkleTree builds and returns the flattened merkle tree.
func (b *Block) GetMerkleTree() [][]byte {
	tree := b.GetTransactionHashes()
	if len(tree) == 0 {
		return tree
	}

	j := 0
	for size := len(b.Transactions); size > 1; size = (size + 1) / 2 {
		for i := 0; i < size; i += 2 {
			i2 := i + 1
			if i2 >= size {
				i2 = size - 1
			}
			concat := append(append([]byte{}, tree[j+i]...), tree[j+i2]...)
			tree = append(tree, crypto.Sha256d(concat))
		}
		j += size
	}
	return tree
}

// GetMerkleRoot calculates the merkle root from transactions.
func (b *Block) GetMerkleRoot() []byte {
	tree := b.GetMerkleTree()
	if len(tree) == 0 {
		return nil
	}
	return tree[len(tree)-1]
}

// ValidMerkleRoot verifies tx merkle root against header merkleRoot.
func (b *Block) ValidMerkleRoot() bool {
	if b == nil || b.Header == nil {
		return false
	}
	calculated := b.GetMerkleRoot()
	if calculated == nil {
		return false
	}
	return bytes.Equal(fit32Bytes(b.Header.MerkleRoot), fit32Bytes(calculated))
}

// Inspect returns a compact debug string.
func (b *Block) Inspect() string {
	return fmt.Sprintf("<Block %s>", b.ID())
}
