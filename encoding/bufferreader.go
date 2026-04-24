package encoding

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrShortBuffer signals a read past the end of the buffer.
var ErrShortBuffer = errors.New("encoding: buffer exhausted")

// BufferReader reads bytes little-endian style, mirroring JS BufferReader.
type BufferReader struct {
	b   []byte
	pos int
}

// NewBufferReader wraps the given bytes for sequential reads.
func NewBufferReader(b []byte) *BufferReader {
	cp := make([]byte, len(b))
	copy(cp, b)
	return &BufferReader{b: cp}
}

// Pos returns the current read offset.
func (r *BufferReader) Pos() int { return r.pos }

// Remaining returns how many unread bytes remain.
func (r *BufferReader) Remaining() int { return len(r.b) - r.pos }

// Read consumes n bytes. Returns ErrShortBuffer if insufficient.
func (r *BufferReader) Read(n int) ([]byte, error) {
	if r.pos+n > len(r.b) {
		return nil, io.ErrUnexpectedEOF
	}
	out := r.b[r.pos : r.pos+n]
	r.pos += n
	return out, nil
}

// ReadUInt8 reads a single byte.
func (r *BufferReader) ReadUInt8() (uint8, error) {
	b, err := r.Read(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

// ReadUInt16LE reads 2 bytes as little-endian unsigned int.
func (r *BufferReader) ReadUInt16LE() (uint16, error) {
	b, err := r.Read(2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(b), nil
}

// ReadUInt32LE reads 4 bytes as little-endian unsigned int.
func (r *BufferReader) ReadUInt32LE() (uint32, error) {
	b, err := r.Read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

// ReadUInt64LE reads 8 bytes as little-endian unsigned int.
func (r *BufferReader) ReadUInt64LE() (uint64, error) {
	b, err := r.Read(8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

// ReadVarInt reads a varint-encoded length. The second return is how many
// bytes were consumed.
func (r *BufferReader) ReadVarInt() (uint64, int, error) {
	if r.pos >= len(r.b) {
		return 0, 0, io.ErrUnexpectedEOF
	}
	v, n := NewVarIntFromBytes(r.b[r.pos:])
	if n == 0 {
		return 0, 0, ErrShortBuffer
	}
	r.pos += n
	return uint64(v), n, nil
}
