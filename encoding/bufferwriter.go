package encoding

import (
	"bytes"
	"encoding/binary"
)

// BufferWriter accumulates bytes little-endian style, mirroring JS BufferWriter.
type BufferWriter struct {
	buf bytes.Buffer
}

// NewBufferWriter returns an empty BufferWriter.
func NewBufferWriter() *BufferWriter { return &BufferWriter{} }

// Bytes returns a copy of the accumulated bytes.
func (w *BufferWriter) Bytes() []byte {
	b := make([]byte, w.buf.Len())
	copy(b, w.buf.Bytes())
	return b
}

// Len returns the number of bytes written.
func (w *BufferWriter) Len() int { return w.buf.Len() }

// WriteBytes appends raw bytes.
func (w *BufferWriter) WriteBytes(b []byte) { w.buf.Write(b) }

// WriteUInt8 writes a 1-byte unsigned int.
func (w *BufferWriter) WriteUInt8(v uint8) { w.buf.WriteByte(v) }

// WriteUInt16LE writes a 2-byte little-endian unsigned int.
func (w *BufferWriter) WriteUInt16LE(v uint16) {
	var b [2]byte
	binary.LittleEndian.PutUint16(b[:], v)
	w.buf.Write(b[:])
}

// WriteUInt32LE writes a 4-byte little-endian unsigned int.
func (w *BufferWriter) WriteUInt32LE(v uint32) {
	var b [4]byte
	binary.LittleEndian.PutUint32(b[:], v)
	w.buf.Write(b[:])
}

// WriteUInt64LE writes an 8-byte little-endian unsigned int.
func (w *BufferWriter) WriteUInt64LE(v uint64) {
	var b [8]byte
	binary.LittleEndian.PutUint64(b[:], v)
	w.buf.Write(b[:])
}

// WriteVarInt writes a varint-encoded length.
func (w *BufferWriter) WriteVarInt(v uint64) {
	w.buf.Write(VarInt(v).Bytes())
}
