package bt

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"math/big"
	"time"

	"github.com/libsv/go-bk/crypto"
)

const (
	// GenesisBits is the compact target for difficulty-1.
	GenesisBits uint32 = 0x1d00ffff

	// BlockHeaderSize is the serialized block header size in bytes.
	BlockHeaderSize = 80

	// BlockHeaderStartOffset is where header starts in a raw block payload.
	BlockHeaderStartOffset = 8

	// MaxBlockTimeOffsetSeconds is the max future timestamp drift.
	MaxBlockTimeOffsetSeconds = 2 * 60 * 60
)

var (
	// LargestHash is 2^256 and used as an upper bound marker.
	LargestHash = new(big.Int).Lsh(big.NewInt(1), 256)
)

// BlockHeader 表示区块头，包含版本、前一区块哈希、默克尔根、时间戳、难度目标和 nonce。
//
// 对应 tbc-lib-js 的 block.header，详见 docs/block.md。
type BlockHeader struct {
	Version    int32  // 区块版本
	PrevHash   []byte // 前一区块哈希
	MerkleRoot []byte // 交易默克尔根
	Time       uint32 // Unix 时间戳
	Bits       uint32 // 难度目标（紧凑表示）
	Nonce      uint32 // 工作量证明随机数
}

// NewBlockHeader returns an empty header instance.
func NewBlockHeader() *BlockHeader {
	return &BlockHeader{
		PrevHash:   make([]byte, 32),
		MerkleRoot: make([]byte, 32),
	}
}

// NewBlockHeaderFromBytes decodes one block header from bytes.
func NewBlockHeaderFromBytes(b []byte) (*BlockHeader, error) {
	if len(b) < BlockHeaderSize {
		return nil, ErrBlockHeaderTooShort
	}

	h := NewBlockHeader()
	if _, err := h.ReadFromBytes(b); err != nil {
		return nil, err
	}
	return h, nil
}

// NewBlockHeaderFromString decodes a hex-encoded header.
func NewBlockHeaderFromString(str string) (*BlockHeader, error) {
	b, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return NewBlockHeaderFromBytes(b)
}

// NewBlockHeaderFromRawBlock decodes a header from raw block bytes (skip 8-byte prefix).
func NewBlockHeaderFromRawBlock(raw []byte) (*BlockHeader, error) {
	if len(raw) < BlockHeaderStartOffset+BlockHeaderSize {
		return nil, ErrBlockHeaderTooShort
	}
	return NewBlockHeaderFromBytes(raw[BlockHeaderStartOffset:])
}

// ReadFrom reads a block header from io.Reader.
func (h *BlockHeader) ReadFrom(r io.Reader) (int64, error) {
	if h == nil {
		return 0, ErrBlockHeaderNil
	}

	buf := make([]byte, BlockHeaderSize)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return int64(n), err
	}
	_, err = h.ReadFromBytes(buf)
	return int64(n), err
}

// ReadFromBytes reads a header from bytes and returns consumed size.
func (h *BlockHeader) ReadFromBytes(b []byte) (int, error) {
	if h == nil {
		return 0, ErrBlockHeaderNil
	}
	if len(b) < BlockHeaderSize {
		return 0, ErrBlockHeaderTooShort
	}

	offset := 0
	h.Version = int32(binary.LittleEndian.Uint32(b[offset : offset+4]))
	offset += 4

	h.PrevHash = make([]byte, 32)
	copy(h.PrevHash, b[offset:offset+32])
	offset += 32

	h.MerkleRoot = make([]byte, 32)
	copy(h.MerkleRoot, b[offset:offset+32])
	offset += 32

	h.Time = binary.LittleEndian.Uint32(b[offset : offset+4])
	offset += 4
	h.Bits = binary.LittleEndian.Uint32(b[offset : offset+4])
	offset += 4
	h.Nonce = binary.LittleEndian.Uint32(b[offset : offset+4])
	offset += 4

	return offset, nil
}

// WriteTo writes the header into writer.
func (h *BlockHeader) WriteTo(w io.Writer) (int64, error) {
	if h == nil {
		return 0, ErrBlockHeaderNil
	}

	b := h.Bytes()
	n, err := w.Write(b)
	return int64(n), err
}

// Bytes encodes the header into wire format.
func (h *BlockHeader) Bytes() []byte {
	if h == nil {
		return nil
	}

	buf := make([]byte, BlockHeaderSize)
	offset := 0
	binary.LittleEndian.PutUint32(buf[offset:offset+4], uint32(h.Version))
	offset += 4
	copy(buf[offset:offset+32], fit32Bytes(h.PrevHash))
	offset += 32
	copy(buf[offset:offset+32], fit32Bytes(h.MerkleRoot))
	offset += 32
	binary.LittleEndian.PutUint32(buf[offset:offset+4], h.Time)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:offset+4], h.Bits)
	offset += 4
	binary.LittleEndian.PutUint32(buf[offset:offset+4], h.Nonce)

	return buf
}

// String encodes the header into hex.
func (h *BlockHeader) String() string {
	return hex.EncodeToString(h.Bytes())
}

// HashBytes returns little-endian double-sha256 hash of header bytes.
func (h *BlockHeader) HashBytes() []byte {
	if h == nil {
		return nil
	}
	return crypto.Sha256d(h.Bytes())
}

// ID returns the big-endian hash string (block hash).
func (h *BlockHeader) ID() string {
	hash := h.HashBytes()
	if hash == nil {
		return ""
	}
	return hex.EncodeToString(ReverseBytes(hash))
}

// TargetDifficulty decodes compact bits into full target.
func (h *BlockHeader) TargetDifficulty(bits ...uint32) *big.Int {
	useBits := h.Bits
	if len(bits) > 0 {
		useBits = bits[0]
	}

	exponent := useBits >> 24
	mantissa := useBits & 0x007fffff
	negative := (useBits & 0x00800000) != 0

	target := new(big.Int).SetUint64(uint64(mantissa))
	if exponent <= 3 {
		shift := 8 * (3 - exponent)
		target.Rsh(target, uint(shift))
	} else {
		shift := 8 * (exponent - 3)
		target.Lsh(target, uint(shift))
	}
	if negative {
		target.Neg(target)
	}
	return target
}

// Difficulty calculates floating-point difficulty relative to GenesisBits.
func (h *BlockHeader) Difficulty() float64 {
	target := h.TargetDifficulty()
	if target.Sign() <= 0 {
		return 0
	}
	base := new(big.Rat).SetInt(new(big.Int).Set(h.TargetDifficulty(GenesisBits)))
	cur := new(big.Rat).SetInt(target)
	ratio := new(big.Rat).Quo(base, cur)
	// 注：big.Rat.Float64 可能返回非精确值，对难度计算使用近似值即可
	f, _ := ratio.Float64()
	return f
}

// ValidTimestamp checks if timestamp is not too far in the future.
func (h *BlockHeader) ValidTimestamp(now ...time.Time) bool {
	if h == nil {
		return false
	}
	ref := time.Now()
	if len(now) > 0 {
		ref = now[0]
	}
	maxTime := ref.Unix() + MaxBlockTimeOffsetSeconds
	return int64(h.Time) <= maxTime
}

// ValidProofOfWork verifies if header hash is <= target.
func (h *BlockHeader) ValidProofOfWork() bool {
	if h == nil {
		return false
	}
	target := h.TargetDifficulty()
	if target.Sign() <= 0 || target.Cmp(LargestHash) > 0 {
		return false
	}

	pow := new(big.Int).SetBytes(ReverseBytes(h.HashBytes()))
	return pow.Cmp(target) <= 0
}

// Inspect returns a compact debug string.
func (h *BlockHeader) Inspect() string {
	return fmt.Sprintf("<BlockHeader %s>", h.ID())
}

func fit32Bytes(b []byte) []byte {
	out := make([]byte, 32)
	if len(b) >= 32 {
		copy(out, b[:32])
		return out
	}
	copy(out, b)
	return out
}

// ClampDifficulty returns a safe finite value for callers that need finite float.
func ClampDifficulty(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}
