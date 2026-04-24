package tbc

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/libsv/go-bk/crypto"

	"github.com/LoongYearMeta/tbc-lib-go/encoding"
)

// MerkleBlock wraps a partial merkle tree payload.
type MerkleBlock struct {
	Header          *BlockHeader
	NumTransactions uint32
	Hashes          []string
	Flags           []byte
}

type merkleTraverseState struct {
	flagBitsUsed uint32
	hashesUsed   uint32
	txs          []string
}

// NewMerkleBlock creates an empty merkle block.
func NewMerkleBlock() *MerkleBlock {
	return &MerkleBlock{
		Header: NewBlockHeader(),
		Hashes: make([]string, 0),
		Flags:  make([]byte, 0),
	}
}

// NewMerkleBlockFromBytes decodes a merkle block from bytes.
func NewMerkleBlockFromBytes(b []byte) (*MerkleBlock, error) {
	mb := NewMerkleBlock()
	if _, err := mb.ReadFromBytes(b); err != nil {
		return nil, err
	}
	return mb, nil
}

// NewMerkleBlockFromString decodes a hex string merkle block.
func NewMerkleBlockFromString(str string) (*MerkleBlock, error) {
	b, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return NewMerkleBlockFromBytes(b)
}

// ReadFrom decodes a merkle block from reader.
func (m *MerkleBlock) ReadFrom(r io.Reader) (int64, error) {
	if m == nil {
		return 0, ErrMerkleBlockNil
	}
	*m = *NewMerkleBlock()

	header := NewBlockHeader()
	n, err := header.ReadFrom(r)
	if err != nil {
		return n, err
	}
	total := n

	u32 := make([]byte, 4)
	nRead, err := io.ReadFull(r, u32)
	total += int64(nRead)
	if err != nil {
		return total, err
	}
	numTransactions := binaryLEToUint32(u32)

	var hashCount encoding.VarInt
	nVar, err := hashCount.ReadFrom(r)
	total += nVar
	if err != nil {
		return total, err
	}

	hashes := make([]string, 0, int(hashCount))
	for i := uint64(0); i < uint64(hashCount); i++ {
		bb := make([]byte, 32)
		nHash, readErr := io.ReadFull(r, bb)
		total += int64(nHash)
		if readErr != nil {
			return total, readErr
		}
		hashes = append(hashes, hex.EncodeToString(bb))
	}

	var flagCount encoding.VarInt
	nFlags, err := flagCount.ReadFrom(r)
	total += nFlags
	if err != nil {
		return total, err
	}

	flags := make([]byte, int(flagCount))
	nRead, err = io.ReadFull(r, flags)
	total += int64(nRead)
	if err != nil {
		return total, err
	}

	m.Header = header
	m.NumTransactions = numTransactions
	m.Hashes = hashes
	m.Flags = flags
	return total, nil
}

// ReadFromBytes decodes merkle block from bytes and returns consumed size.
func (m *MerkleBlock) ReadFromBytes(b []byte) (int, error) {
	if m == nil {
		return 0, ErrMerkleBlockNil
	}
	reader := bytes.NewReader(b)
	n, err := m.ReadFrom(reader)
	return int(n), err
}

// Bytes encodes merkle block.
func (m *MerkleBlock) Bytes() []byte {
	if m == nil || m.Header == nil {
		return nil
	}

	buf := bytes.NewBuffer(make([]byte, 0))
	buf.Write(m.Header.Bytes())
	u32 := make([]byte, 4)
	putBinaryLEUint32(u32, m.NumTransactions)
	buf.Write(u32)
	buf.Write(encoding.VarInt(uint64(len(m.Hashes))).Bytes())
	for _, h := range m.Hashes {
		hashBytes, err := hex.DecodeString(h)
		if err != nil {
			continue
		}
		buf.Write(fit32Bytes(hashBytes))
	}
	buf.Write(encoding.VarInt(uint64(len(m.Flags))).Bytes())
	buf.Write(m.Flags)

	return buf.Bytes()
}

// String returns hex representation.
func (m *MerkleBlock) String() string {
	return hex.EncodeToString(m.Bytes())
}

// ValidMerkleTree validates the partial merkle tree.
func (m *MerkleBlock) ValidMerkleTree() bool {
	if m == nil || m.Header == nil {
		return false
	}
	if len(m.Hashes) > int(m.NumTransactions) {
		return false
	}
	if len(m.Flags)*8 < len(m.Hashes) {
		return false
	}

	height := m.calcTreeHeight()
	state := &merkleTraverseState{}
	root, ok := m.traverseMerkleTree(height, 0, state, false)
	if !ok || root == nil {
		return false
	}
	if int(state.hashesUsed) != len(m.Hashes) {
		return false
	}
	return bytes.Equal(root, fit32Bytes(m.Header.MerkleRoot))
}

// FilteredTxsHash returns all tx hashes matching the filter.
func (m *MerkleBlock) FilteredTxsHash() ([]string, error) {
	if m == nil || m.Header == nil {
		return nil, ErrMerkleBlockNil
	}
	if len(m.Hashes) > int(m.NumTransactions) {
		return nil, ErrInvalidMerkleTree
	}
	if len(m.Flags)*8 < len(m.Hashes) {
		return nil, ErrInvalidMerkleTree
	}
	if len(m.Hashes) == 1 {
		return []string{}, nil
	}

	height := m.calcTreeHeight()
	state := &merkleTraverseState{txs: make([]string, 0)}
	_, ok := m.traverseMerkleTree(height, 0, state, true)
	if !ok {
		return nil, ErrInvalidMerkleTree
	}
	if int(state.hashesUsed) != len(m.Hashes) {
		return nil, ErrInvalidMerkleTree
	}
	return state.txs, nil
}

// HasTransaction checks if tx id or tx object exists in filtered results.
func (m *MerkleBlock) HasTransaction(tx interface{}) bool {
	if tx == nil {
		return false
	}

	var hash string
	switch v := tx.(type) {
	case string:
		hash = v
	case *Tx:
		if v == nil {
			return false
		}
		// Match JS behavior: reverse displayed txid for internal lookup.
		raw, err := hex.DecodeString(v.TxID())
		if err != nil {
			return false
		}
		hash = hex.EncodeToString(encoding.ReverseBytes(raw))
	default:
		return false
	}

	txs, err := m.FilteredTxsHash()
	if err != nil {
		return false
	}
	for _, h := range txs {
		if h == hash {
			return true
		}
	}
	return false
}

func (m *MerkleBlock) calcTreeWidth(height uint32) uint32 {
	return (m.NumTransactions + (1 << height) - 1) >> height
}

func (m *MerkleBlock) calcTreeHeight() uint32 {
	height := uint32(0)
	for m.calcTreeWidth(height) > 1 {
		height++
	}
	return height
}

func (m *MerkleBlock) traverseMerkleTree(depth, pos uint32, state *merkleTraverseState, collectTxs bool) ([]byte, bool) {
	if state == nil {
		state = &merkleTraverseState{}
	}
	if state.flagBitsUsed >= uint32(len(m.Flags))*8 {
		return nil, false
	}

	flagByte := m.Flags[state.flagBitsUsed>>3]
	isParentOfMatch := (flagByte >> (state.flagBitsUsed & 7)) & 1
	state.flagBitsUsed++

	if depth == 0 || isParentOfMatch == 0 {
		if state.hashesUsed >= uint32(len(m.Hashes)) {
			return nil, false
		}
		hashStr := m.Hashes[state.hashesUsed]
		state.hashesUsed++

		hashBytes, err := hex.DecodeString(hashStr)
		if err != nil {
			return nil, false
		}

		if depth == 0 && isParentOfMatch == 1 && collectTxs {
			state.txs = append(state.txs, hashStr)
		}
		return fit32Bytes(hashBytes), true
	}

	left, ok := m.traverseMerkleTree(depth-1, pos*2, state, collectTxs)
	if !ok || left == nil {
		return nil, false
	}

	right := left
	if pos*2+1 < m.calcTreeWidth(depth-1) {
		var rightOK bool
		right, rightOK = m.traverseMerkleTree(depth-1, pos*2+1, state, collectTxs)
		if !rightOK || right == nil {
			return nil, false
		}
	}

	if collectTxs {
		return nil, true
	}
	return crypto.Sha256d(append(append([]byte{}, left...), right...)), true
}

func binaryLEToUint32(b []byte) uint32 {
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func putBinaryLEUint32(b []byte, v uint32) {
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}
