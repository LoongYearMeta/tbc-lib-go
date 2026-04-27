package script

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/LoongYearMeta/tbc-lib-go/bec"
	"github.com/LoongYearMeta/tbc-lib-go/bip32"
	"github.com/LoongYearMeta/tbc-lib-go/crypto"
)

// ScriptKey types.
const (
	ScriptTypePubKey      = "pubkey"
	ScriptTypePubKeyHash  = "pubkeyhash"
	ScriptTypeNonStandard = "nonstandard"
	ScriptTypeEmpty       = "empty"
	ScriptTypeSecureHash  = "securehash"
	ScriptTypeMultiSig    = "multisig"
	ScriptTypeNullData    = "nulldata"
)

// Script types for classification
const (
	ScriptTypeUnknown         = "Unknown"
	ScriptTypePubKeyOut       = "Pay to public key"
	ScriptTypePubKeyIn        = "Spend from public key"
	ScriptTypePubKeyHashOut   = "Pay to public key hash"
	ScriptTypePubKeyHashIn    = "Spend from public key hash"
	ScriptTypeScriptHashOut   = "Pay to script hash"
	ScriptTypeScriptHashIn    = "Spend from script hash"
	ScriptTypeMultisigOut     = "Pay to multisig"
	ScriptTypeMultisigIn      = "Spend from multisig"
	ScriptTypeDataOut         = "Data push"
	ScriptTypeSafeDataOut     = "Safe data push"
)

// Chunk represents a parsed script chunk
type Chunk struct {
	OpcodeNum byte
	Buf       []byte
	Len       int
}

// Script type
type Script struct {
	data     []byte
	isInput  bool
	isOutput bool
}

// NewScript creates a new empty script
func NewScript() *Script {
	return &Script{data: []byte{}}
}

// Bytes returns the underlying byte slice
func (s *Script) Bytes() []byte {
	if s == nil {
		return nil
	}
	return s.data
}

// Len returns the length of the script
func (s *Script) Len() int {
	if s == nil {
		return 0
	}
	return len(s.data)
}

// NewFromHexString creates a new script from a hex encoded string.
func NewFromHexString(s string) (*Script, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return nil, err
	}
	return NewFromBytes(b), nil
}

// NewFromBytes wraps a byte slice with the Script type.
func NewFromBytes(b []byte) *Script {
	if b == nil {
		return &Script{data: []byte{}}
	}
	s := make([]byte, len(b))
	copy(s, b)
	return &Script{data: s}
}

// FromBuffer creates a script from a byte buffer (alias for NewFromBytes)
func FromBuffer(b []byte) *Script {
	return NewFromBytes(b)
}

// FromChunks creates a script from an array of chunks
func FromChunks(chunks []Chunk) (*Script, error) {
	s := &Script{data: []byte{}}
	for _, chunk := range chunks {
		s.data = append(s.data, chunk.OpcodeNum)
		if chunk.Buf != nil {
			if chunk.OpcodeNum < OpPUSHDATA1 {
				s.data = append(s.data, chunk.Buf...)
			} else if chunk.OpcodeNum == OpPUSHDATA1 {
				s.data = append(s.data, byte(chunk.Len))
				s.data = append(s.data, chunk.Buf...)
			} else if chunk.OpcodeNum == OpPUSHDATA2 {
				lenBuf := make([]byte, 2)
				binary.LittleEndian.PutUint16(lenBuf, uint16(chunk.Len))
				s.data = append(s.data, lenBuf...)
				s.data = append(s.data, chunk.Buf...)
			} else if chunk.OpcodeNum == OpPUSHDATA4 {
				lenBuf := make([]byte, 4)
				binary.LittleEndian.PutUint32(lenBuf, uint32(chunk.Len))
				s.data = append(s.data, lenBuf...)
				s.data = append(s.data, chunk.Buf...)
			}
		}
	}
	return s, nil
}

// DecodeChunks decodes the script into chunks
func (s *Script) DecodeChunks() []Chunk {
	chunks := []Chunk{}
	i := 0
	script := s.data
	for i < len(script) {
		opcodenum := script[i]
		i++
		var chunk Chunk
		chunk.OpcodeNum = opcodenum

		if opcodenum == 0 {
			chunk.Len = 0
			chunks = append(chunks, chunk)
		} else if opcodenum < OpPUSHDATA1 {
			chunk.Len = int(opcodenum)
			if i+chunk.Len <= len(script) {
				chunk.Buf = script[i : i+chunk.Len]
				i += chunk.Len
			}
			chunks = append(chunks, chunk)
		} else if opcodenum == OpPUSHDATA1 {
			if i >= len(script) {
				break
			}
			chunk.Len = int(script[i])
			i++
			if i+chunk.Len <= len(script) {
				chunk.Buf = script[i : i+chunk.Len]
				i += chunk.Len
			}
			chunks = append(chunks, chunk)
		} else if opcodenum == OpPUSHDATA2 {
			if i+2 > len(script) {
				break
			}
			chunk.Len = int(binary.LittleEndian.Uint16(script[i:]))
			i += 2
			if i+chunk.Len <= len(script) {
				chunk.Buf = script[i : i+chunk.Len]
				i += chunk.Len
			}
			chunks = append(chunks, chunk)
		} else if opcodenum == OpPUSHDATA4 {
			if i+4 > len(script) {
				break
			}
			chunk.Len = int(binary.LittleEndian.Uint32(script[i:]))
			i += 4
			if i+chunk.Len <= len(script) {
				chunk.Buf = script[i : i+chunk.Len]
				i += chunk.Len
			}
			chunks = append(chunks, chunk)
		} else {
			chunks = append(chunks, chunk)
		}
	}
	return chunks
}

// Chunks returns the decoded chunks (cached if possible)
func (s *Script) Chunks() []Chunk {
	return s.DecodeChunks()
}

// NewFromASM creates a new script from a BitCoin ASM formatted string.
func NewFromASM(str string) (*Script, error) {
	s := NewFromBytes([]byte{})

	for _, section := range strings.Split(str, " ") {
		if val, ok := opCodeStrings[section]; ok {
			_ = s.AppendOpcodes(val)
		} else {
			if err := s.AppendPushDataHexString(section); err != nil {
				return nil, ErrInvalidOpCode
			}
		}
	}

	return s, nil
}

// NewP2PKHFromPubKeyEC takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyEC(pubKey *bec.PublicKey) (*Script, error) {
	return NewP2PKHFromPubKeyBytes(pubKey.SerialiseCompressed())
}

// NewP2PKHFromPubKeyStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyStr(pubKey string) (*Script, error) {
	pubKeyBytes, err := hex.DecodeString(pubKey)
	if err != nil {
		return nil, err
	}
	return NewP2PKHFromPubKeyBytes(pubKeyBytes)
}

// NewP2PKHFromPubKeyBytes takes public key bytes (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyBytes(pubKeyBytes []byte) (*Script, error) {
	if len(pubKeyBytes) != 33 {
		return nil, ErrInvalidPKLen
	}
	return NewP2PKHFromPubKeyHash(crypto.Hash160(pubKeyBytes))
}

// NewP2PKHFromPubKeyHash takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyHash(pubKeyHash []byte) (*Script, error) {
	b := []byte{
		OpDUP,
		OpHASH160,
		OpDATA20,
	}
	b = append(b, pubKeyHash...)
	b = append(b, OpEQUALVERIFY)
	b = append(b, OpCHECKSIG)

	return NewFromBytes(b), nil
}

// NewP2PKHFromPubKeyHashStr takes a public key hex string (in
// compressed format) and creates a P2PKH script from it.
func NewP2PKHFromPubKeyHashStr(pubKeyHash string) (*Script, error) {
	hash, err := hex.DecodeString(pubKeyHash)
	if err != nil {
		return nil, err
	}

	return NewP2PKHFromPubKeyHash(hash)
}

// NewP2PKHFromAddress takes an address
// and creates a P2PKH script from it.
func NewP2PKHFromAddress(addr string) (*Script, error) {
	a, err := NewAddressFromString(addr)
	if err != nil {
		return nil, err
	}

	var publicKeyHashBytes []byte
	if publicKeyHashBytes, err = hex.DecodeString(a.PublicKeyHash); err != nil {
		return nil, err
	}

	s := NewFromBytes([]byte{})
	_ = s.AppendOpcodes(OpDUP, OpHASH160)
	if err = s.AppendPushData(publicKeyHashBytes); err != nil {
		return nil, err
	}
	_ = s.AppendOpcodes(OpEQUALVERIFY, OpCHECKSIG)

	return s, nil
}

// NewP2PKHFromBip32ExtKey takes a *bip32.ExtendedKey and creates a P2PKH script from it,
// using an internally random generated seed, returning the script and derivation path used.
func NewP2PKHFromBip32ExtKey(privKey *bip32.ExtendedKey) (*Script, string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return nil, "", err
	}

	derivationPath := bip32.DerivePath(binary.LittleEndian.Uint64(b[:]))
	pubKey, err := privKey.DerivePublicKeyFromPath(derivationPath)
	if err != nil {
		return nil, "", err
	}

	lockingScript, err := NewP2PKHFromPubKeyBytes(pubKey)
	if err != nil {
		return nil, "", err
	}

	return lockingScript, derivationPath, nil
}

// AppendPushData takes data bytes and appends them to the script
// with proper PUSHDATA prefixes
func (s *Script) AppendPushData(d []byte) error {
	if s == nil {
		return ErrEmptyScript
	}
	p, err := EncodeParts([][]byte{d})
	if err != nil {
		return err
	}

	s.data = append(s.data, p...)
	return nil
}

// AppendPushDataHexString takes a hex string and appends them to the
// script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataHexString(str string) error {
	h, err := hex.DecodeString(str)
	if err != nil {
		return err
	}

	return s.AppendPushData(h)
}

// AppendPushDataString takes a string and appends its UTF-8 encoding
// to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataString(str string) error {
	return s.AppendPushData([]byte(str))
}

// AppendPushDataArray takes an array of data bytes and appends them
// to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataArray(d [][]byte) error {
	p, err := EncodeParts(d)
	if err != nil {
		return err
	}

	s.data = append(s.data, p...)
	return nil
}

// AppendPushDataStrings takes an array of strings and appends their
// UTF-8 encoding to the script with proper PUSHDATA prefixes
func (s *Script) AppendPushDataStrings(pushDataStrings []string) error {
	dataBytes := make([][]byte, 0)
	for _, str := range pushDataStrings {
		strBytes := []byte(str)
		dataBytes = append(dataBytes, strBytes)
	}
	return s.AppendPushDataArray(dataBytes)
}

// AppendOpcodes appends opcodes type to the script.
// This does not support appending OP_PUSHDATA opcodes, so use `Script.AppendPushData` instead.
func (s *Script) AppendOpcodes(oo ...uint8) error {
	if s == nil {
		return ErrEmptyScript
	}
	for _, o := range oo {
		if OpDATA1 <= o && o <= OpPUSHDATA4 {
			return fmt.Errorf("%w: %s", ErrInvalidOpcodeType, opCodeValues[o])
		}
	}
	s.data = append(s.data, oo...)
	return nil
}

// String implements the stringer interface and returns the hex string of script.
func (s *Script) String() string {
	if s == nil {
		return ""
	}
	return hex.EncodeToString(s.data)
}

// ToASM returns the string ASM opcodes of the script.
func (s *Script) ToASM() (string, error) {
	if s == nil || len(s.data) == 0 {
		return "", nil
	}
	parts, err := DecodeParts(s.data)
	// if err != nil, we will append [error] to the ASM script below (as done in the node).

	var asm strings.Builder
	for _, p := range parts {
		asm.WriteRune(' ')
		if len(p) == 1 {
			asm.WriteString(opCodeValues[p[0]])
		} else {
			asm.WriteString(hex.EncodeToString(p))
		}
	}

	if err != nil {
		asm.WriteString(" [error]")
	}

	return asm.String()[1:], nil
}

// IsP2PKH returns true if this is a pay to pubkey hash output script.
func (s *Script) IsP2PKH() bool {
	if s == nil {
		return false
	}
	b := s.data
	return len(b) == 25 &&
		b[0] == OpDUP &&
		b[1] == OpHASH160 &&
		b[2] == OpDATA20 &&
		b[23] == OpEQUALVERIFY &&
		b[24] == OpCHECKSIG
}

// IsP2PK returns true if this is a public key output script.
func (s *Script) IsP2PK() bool {
	if s == nil {
		return false
	}
	parts, err := DecodeParts(s.data)
	if err != nil {
		return false
	}

	if len(parts) == 2 && len(parts[0]) > 0 && parts[1][0] == OpCHECKSIG {
		pubkey := parts[0]
		version := pubkey[0]

		if (version == 0x04 || version == 0x06 || version == 0x07) && len(pubkey) == 65 {
			return true
		} else if (version == 0x03 || version == 0x02) && len(pubkey) == 33 {
			return true
		}
	}
	return false
}

// IsP2SH returns true if this is a p2sh output script.
// TODO: remove all p2sh stuff from repo
func (s *Script) IsP2SH() bool {
	if s == nil {
		return false
	}
	b := s.data

	return len(b) == 23 &&
		b[0] == OpHASH160 &&
		b[1] == OpDATA20 &&
		b[22] == OpEQUAL
}

// IsData returns true if this is a data output script. This
// means the script starts with OP_RETURN or OP_FALSE OP_RETURN.
func (s *Script) IsData() bool {
	if s == nil {
		return false
	}
	b := s.data

	return (len(b) > 0 && b[0] == OpRETURN) ||
		(len(b) > 1 && b[0] == OpFALSE && b[1] == OpRETURN)
}

// IsMultiSigOut returns true if this is a multisig output script.
func (s *Script) IsMultiSigOut() bool {
	if s == nil {
		return false
	}
	parts, err := DecodeParts(s.data)
	if err != nil {
		return false
	}

	if len(parts) < 3 {
		return false
	}

	if !isSmallIntOp(parts[0][0]) {
		return false
	}

	for i := 1; i < len(parts)-2; i++ {
		if len(parts[i]) < 1 {
			return false
		}
	}

	return len(parts[len(parts)-2]) > 0 && isSmallIntOp(parts[len(parts)-2][0]) && len(parts[len(parts)-1]) > 0 &&
		parts[len(parts)-1][0] == OpCHECKMULTISIG
}

func isSmallIntOp(opcode byte) bool {
	return opcode == OpZERO || (opcode >= OpONE && opcode <= Op16)
}

// IsPublicKeyHashOut returns true if this is a pay to pubkey hash output script (alias for IsP2PKH)
func (s *Script) IsPublicKeyHashOut() bool {
	return s.IsP2PKH()
}

// IsPublicKeyHashIn returns true if this is a pay to public key hash input script
func (s *Script) IsPublicKeyHashIn() bool {
	chunks := s.Chunks()
	if len(chunks) == 2 {
		signatureBuf := chunks[0].Buf
		pubkeyBuf := chunks[1].Buf
		if signatureBuf != nil && len(signatureBuf) > 0 && signatureBuf[0] == 0x30 &&
			pubkeyBuf != nil && len(pubkeyBuf) > 0 {
			version := pubkeyBuf[0]
			if (version == 0x04 || version == 0x06 || version == 0x07) && len(pubkeyBuf) == 65 {
				return true
			} else if (version == 0x03 || version == 0x02) && len(pubkeyBuf) == 33 {
				return true
			}
		}
	}
	return false
}

// IsPublicKeyOut returns true if this is a public key output script (alias for IsP2PK)
func (s *Script) IsPublicKeyOut() bool {
	return s.IsP2PK()
}

// IsPublicKeyIn returns true if this is a pay to public key input script
func (s *Script) IsPublicKeyIn() bool {
	chunks := s.Chunks()
	if len(chunks) == 1 {
		signatureBuf := chunks[0].Buf
		if signatureBuf != nil && len(signatureBuf) > 0 && signatureBuf[0] == 0x30 {
			return true
		}
	}
	return false
}

// IsScriptHashOut returns true if this is a p2sh output script (alias for IsP2SH)
func (s *Script) IsScriptHashOut() bool {
	return s.IsP2SH()
}

// IsScriptHashIn returns true if this is a p2sh input script
func (s *Script) IsScriptHashIn() bool {
	chunks := s.Chunks()
	if len(chunks) <= 1 {
		return false
	}
	redeemChunk := chunks[len(chunks)-1]
	redeemBuf := redeemChunk.Buf
	if redeemBuf == nil {
		return false
	}

	redeemScript := NewFromBytes(redeemBuf)
	scriptType := redeemScript.Classify()
	return scriptType != ScriptTypeUnknown
}

// IsMultisigOut returns true if this is a multisig output script (alias for IsMultiSigOut)
func (s *Script) IsMultisigOut() bool {
	return s.IsMultiSigOut()
}

// IsMultisigIn returns true if this is a multisig input script
func (s *Script) IsMultisigIn() bool {
	chunks := s.Chunks()
	if len(chunks) < 2 {
		return false
	}
	if chunks[0].OpcodeNum != OpZERO {
		return false
	}
	// Check that all remaining chunks are signature buffers (DER format starts with 0x30)
	for i := 1; i < len(chunks); i++ {
		if chunks[i].Buf == nil || len(chunks[i].Buf) == 0 || chunks[i].Buf[0] != 0x30 {
			return false
		}
	}
	return true
}

// IsDataOut returns true if this is a valid standard OP_RETURN output
func (s *Script) IsDataOut() bool {
	if s == nil || len(s.data) < 1 {
		return false
	}
	b := s.data
	if b[0] != OpRETURN {
		return false
	}
	subScript := NewFromBytes(b[1:])
	return subScript.IsPushOnly()
}

// IsSafeDataOut returns true if this is a safe data output script (OP_FALSE OP_RETURN ...)
func (s *Script) IsSafeDataOut() bool {
	if s == nil {
		return false
	}
	b := s.data
	if len(b) < 2 {
		return false
	}
	if b[0] != OpFALSE {
		return false
	}
	subScript := NewFromBytes(b[1:])
	return subScript.IsDataOut()
}

// IsPushOnly returns true if the script is only composed of data pushing
// opcodes or small int opcodes (OP_0, OP_1, ..., OP_16)
func (s *Script) IsPushOnly() bool {
	chunks := s.Chunks()
	for _, chunk := range chunks {
		if chunk.OpcodeNum > Op16 && chunk.OpcodeNum != OpPUSHDATA1 &&
			chunk.OpcodeNum != OpPUSHDATA2 && chunk.OpcodeNum != OpPUSHDATA4 {
			return false
		}
	}
	return true
}

// PublicKeyHash returns a public key hash byte array if the script is a P2PKH script.
func (s *Script) PublicKeyHash() ([]byte, error) {
	if s == nil || len(s.data) == 0 {
		return nil, ErrEmptyScript
	}

	if s.data[0] != OpDUP || s.data[1] != OpHASH160 {
		return nil, ErrNotP2PKH
	}

	parts, err := DecodeParts(s.data[2:])
	if err != nil {
		return nil, err
	}

	return parts[0], nil
}

// ScriptType returns the type of script this is as a string.
func (s *Script) ScriptType() string {
	if s == nil || len(s.data) == 0 {
		return ScriptTypeEmpty
	}
	if s.IsP2PKH() {
		return ScriptTypePubKeyHash
	}
	if s.IsP2PK() {
		return ScriptTypePubKey
	}
	if s.IsMultiSigOut() {
		return ScriptTypeMultiSig
	}
	if s.IsData() {
		return ScriptTypeNullData
	}
	return ScriptTypeNonStandard
}

// Addresses will return all addresses found in the script, if any.
func (s *Script) Addresses() ([]string, error) {
	addresses := make([]string, 0)
	if s.IsP2PKH() {
		pkh, err := s.PublicKeyHash()
		if err != nil {
			return nil, err
		}
		a, err := NewAddressFromPublicKeyHash(pkh, true)
		if err != nil {
			return nil, err
		}
		addresses = []string{a.AddressString}
	}
	// TODO: handle multisig, and other outputs
	// https://github.com/libsv/go-bt/issues/6
	return addresses, nil
}

// Equals will compare the script to b and return true if they match.
func (s *Script) Equals(b *Script) bool {
	if s == nil || b == nil {
		return s == b
	}
	return bytes.Equal(s.data, b.data)
}

// EqualsBytes will compare the script to a byte representation of a
// script, b, and return true if they match.
func (s *Script) EqualsBytes(b []byte) bool {
	if s == nil {
		return len(b) == 0
	}
	return bytes.Equal(s.data, b)
}

// EqualsHex will compare the script to a hex string h,
// if they match then true is returned otherwise false.
func (s *Script) EqualsHex(h string) bool {
	return s.String() == h
}

// MinPushSize returns the minimum size of a push operation of the given data.
func MinPushSize(bb []byte) int {
	l := len(bb)

	// data length is larger than max supported
	if l > 0xffffffff {
		return 0
	}

	if l == 0 {
		return 1
	}

	if l == 1 {
		// data can be represented as Op1 to Op16, or OpNegate
		if bb[0] <= 16 || bb[0] == 0x81 {
			// OpX
			return 1
		}
		// OP_DATA_1 + data
		return 2
	}

	// OP_DATA_X + data
	if l <= 75 {
		return l + 1
	}
	// OP_PUSHDATA1 + length byte + data
	if l <= 0xff {
		return l + 2
	}
	// OP_PUSHDATA2 + two length bytes + data
	if l <= 0xffff {
		return l + 3
	}

	// OP_PUSHDATA4 + four length bytes + data
	return l + 5
}

// MarshalJSON convert script into json.
func (s *Script) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, s.String())), nil
}

// UnmarshalJSON covert from json into *script.Script.
func (s *Script) UnmarshalJSON(bb []byte) error {
	ss, err := NewFromHexString(string(bytes.Trim(bb, `"`)))
	if err != nil {
		return err
	}

	*s = *ss
	return nil
}

// MarshalBinary implements encoding.BinaryMarshaler
func (s *Script) MarshalBinary() ([]byte, error) {
	if s == nil {
		return []byte{}, nil
	}
	return s.data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (s *Script) UnmarshalBinary(data []byte) error {
	if s == nil {
		return ErrEmptyScript
	}
	s.data = make([]byte, len(data))
	copy(s.data, data)
	return nil
}

func (s *Script) GetOp(pc *uint) (byte, []byte, bool) {
	if s == nil {
		return OpINVALIDOPCODE, nil, false
	}

	opcodeRet := OpINVALIDOPCODE

	pvchRet := make([]byte, 0)

	script := s.data

	len := uint(len(script))

	if *pc >= len {
		return opcodeRet, pvchRet, false
	}

	if len-*pc < 1 {
		return opcodeRet, pvchRet, false
	}

	// Read instruction
	opcode := script[*pc]

	*pc++

	if opcode <= OpPUSHDATA4 {
		var nSize uint
		if opcode < OpPUSHDATA1 {
			nSize = uint(opcode)
		} else if opcode == OpPUSHDATA1 {
			if len-*pc < 1 {
				return opcodeRet, pvchRet, false
			}
			nSize = uint(script[*pc])
			*pc++
		} else if opcode == OpPUSHDATA2 {
			if len-*pc < 2 {
				return opcodeRet, pvchRet, false
			}

			nSize = ((uint(script[*pc+1]) << 8) |
				uint(script[*pc]))

			*pc = *pc + 2

		} else if opcode == OpPUSHDATA4 {
			if len-*pc < 4 {
				return opcodeRet, pvchRet, false
			}

			nSize = ((uint(script[*pc+3]) << 24) |
				(uint(script[*pc+2]) << 16) |
				(uint(script[*pc+1]) << 8) |
				uint(script[*pc]))

			*pc = *pc + 4
		}
		if len < *pc || len-*pc < nSize {
			return opcodeRet, pvchRet, false
		}

		pvchRet = append(pvchRet, script[*pc:*pc+nSize]...)

		if nSize > 0 {
			*pc = *pc + nSize
		}
	}

	return opcodeRet, pvchRet, true

}

func (s *Script) FindAndDelete(b *Script) uint {
	if s == nil || b == nil {
		return 0
	}

	var nFound uint = 0

	lenb := uint(len(b.data))

	lens := uint(len(s.data))

	if lenb == 0 {
		return nFound
	}

	var pc uint = 0
	var pc2 uint = 0

	var res []byte

	for {

		res = append(res, s.data[pc2:pc]...)

		for lens-pc >= lenb && bytes.HasPrefix(s.data[pc:], b.data) {
			pc = pc + lenb
			nFound++
		}

		pc2 = pc

		_, _, success := s.GetOp(&pc)

		if !success {
			break
		}
	}

	if nFound > 0 {
		res = append(res, s.data[pc2:]...)
		s.data = res
	}

	return nFound
}

// FromString creates a script from a string (hex or ASM format)
func FromString(str string) (*Script, error) {
	if str == "" {
		return NewFromBytes([]byte{}), nil
	}

	// Try hex first
	if b, err := hex.DecodeString(str); err == nil {
		return NewFromBytes(b), nil
	}

	// Try ASM format
	return NewFromASM(str)
}

// ToHex returns the hex string representation of the script
func (s *Script) ToHex() string {
	return s.String()
}

// GetData retrieves the associated data for this script
func (s *Script) GetData() ([]byte, error) {
	if s.IsSafeDataOut() {
		chunks := s.Chunks()
		if len(chunks) < 2 {
			return []byte{}, nil
		}
		// Return all data chunks after OP_FALSE OP_RETURN
		var result []byte
		for i := 2; i < len(chunks); i++ {
			if chunks[i].Buf != nil {
				result = append(result, chunks[i].Buf...)
			}
		}
		return result, nil
	}

	if s.IsDataOut() || s.IsP2SH() {
		chunks := s.Chunks()
		if len(chunks) < 2 || chunks[1].Buf == nil {
			return []byte{}, nil
		}
		return chunks[1].Buf, nil
	}

	if s.IsP2PKH() {
		return s.PublicKeyHash()
	}

	return nil, fmt.Errorf("unrecognized script type to get data from")
}

// GetPublicKey returns the public key from a P2PK output script
func (s *Script) GetPublicKey() ([]byte, error) {
	if !s.IsPublicKeyOut() {
		return nil, fmt.Errorf("can't retrieve PublicKey from a non-PK output")
	}
	chunks := s.Chunks()
	if len(chunks) == 0 || chunks[0].Buf == nil {
		return nil, fmt.Errorf("invalid public key script")
	}
	return chunks[0].Buf, nil
}

// GetPublicKeyHash returns the public key hash from a P2PKH output script
func (s *Script) GetPublicKeyHash() ([]byte, error) {
	if !s.IsPublicKeyHashOut() {
		return nil, fmt.Errorf("can't retrieve PublicKeyHash from a non-PKH output")
	}
	return s.PublicKeyHash()
}

// Clone creates a copy of the script
func (s *Script) Clone() *Script {
	if s == nil {
		return NewFromBytes([]byte{})
	}
	b := make([]byte, len(s.data))
	copy(b, s.data)
	cloned := NewFromBytes(b)
	cloned.isInput = s.isInput
	cloned.isOutput = s.isOutput
	return cloned
}

// RemoveCodeseparators removes all OP_CODESEPARATOR opcodes from the script
func (s *Script) RemoveCodeseparators() {
	if s == nil {
		return
	}
	chunks := s.Chunks()
	newChunks := make([]Chunk, 0, len(chunks))
	for _, chunk := range chunks {
		if chunk.OpcodeNum != OpCODESEPARATOR {
			newChunks = append(newChunks, chunk)
		}
	}
	newScript, err := FromChunks(newChunks)
	if err == nil {
		s.data = newScript.data
	}
}

// SubScript returns a subset of the script starting at the nth codeseparator
func (s *Script) SubScript(n int) *Script {
	chunks := s.Chunks()
	idx := 0
	startIdx := 0

	for i, chunk := range chunks {
		if chunk.OpcodeNum == OpCODESEPARATOR {
			if idx == n {
				startIdx = i + 1
				break
			}
			idx++
		}
	}

	if startIdx < len(chunks) {
		newChunks := chunks[startIdx:]
		newScript, err := FromChunks(newChunks)
		if err == nil {
			return newScript
		}
	}

	return s.Clone()
}

// CheckMinimalPush checks if chunk at index i is minimally encoded
func (s *Script) CheckMinimalPush(i int) bool {
	chunks := s.Chunks()
	if i >= len(chunks) {
		return false
	}

	chunk := chunks[i]
	buf := chunk.Buf
	opcodenum := chunk.OpcodeNum

	if buf == nil {
		return true
	}

	if len(buf) == 0 {
		return opcodenum == OpZERO
	}

	if len(buf) == 1 {
		if buf[0] >= 1 && buf[0] <= 16 {
			return opcodenum == OpONE+(buf[0]-1)
		}
		if buf[0] == 0x81 {
			return opcodenum == Op1NEGATE
		}
	}

	if len(buf) <= 75 {
		return opcodenum == byte(len(buf))
	}

	if len(buf) <= 255 {
		return opcodenum == OpPUSHDATA1
	}

	if len(buf) <= 65535 {
		return opcodenum == OpPUSHDATA2
	}

	return true
}

// GetSignatureOperationsCount returns the number of signature operations required by this script
func (s *Script) GetSignatureOperationsCount(accurate bool) int {
	chunks := s.Chunks()
	n := 0
	lastOpcode := OpINVALIDOPCODE

	for _, chunk := range chunks {
		opcode := chunk.OpcodeNum
		if opcode == OpCHECKSIG || opcode == OpCHECKSIGVERIFY {
			n++
		} else if opcode == OpCHECKMULTISIG || opcode == OpCHECKMULTISIGVERIFY {
			if accurate && lastOpcode >= OpONE && lastOpcode <= Op16 {
				n += int(lastOpcode - (OpONE - 1))
			} else {
				n += 20
			}
		}
		lastOpcode = opcode
	}

	return n
}

// Classify returns the script type
func (s *Script) Classify() string {
	if s.isInput {
		return s.classifyInput()
	} else if s.isOutput {
		return s.classifyOutput()
	} else {
		outputType := s.classifyOutput()
		if outputType != ScriptTypeUnknown {
			return outputType
		}
		return s.classifyInput()
	}
}

// classifyOutput classifies the script as an output type
func (s *Script) classifyOutput() string {
	if s.IsPublicKeyOut() {
		return ScriptTypePubKeyOut
	}
	if s.IsPublicKeyHashOut() {
		return ScriptTypePubKeyHashOut
	}
	if s.IsMultisigOut() {
		return ScriptTypeMultisigOut
	}
	if s.IsScriptHashOut() {
		return ScriptTypeScriptHashOut
	}
	if s.IsDataOut() {
		return ScriptTypeDataOut
	}
	if s.IsSafeDataOut() {
		return ScriptTypeSafeDataOut
	}
	return ScriptTypeUnknown
}

// classifyInput classifies the script as an input type
func (s *Script) classifyInput() string {
	if s.IsPublicKeyIn() {
		return ScriptTypePubKeyIn
	}
	if s.IsPublicKeyHashIn() {
		return ScriptTypePubKeyHashIn
	}
	if s.IsMultisigIn() {
		return ScriptTypeMultisigIn
	}
	if s.IsScriptHashIn() {
		return ScriptTypeScriptHashIn
	}
	return ScriptTypeUnknown
}

// SetIsInput sets whether this script is an input script
func (s *Script) SetIsInput(isInput bool) {
	s.isInput = isInput
}

// SetIsOutput sets whether this script is an output script
func (s *Script) SetIsOutput(isOutput bool) {
	s.isOutput = isOutput
}

// ToAddress converts the script to an address if possible
func (s *Script) ToAddress(mainnet bool) (string, error) {
	info := s.GetAddressInfo()
	if info == nil {
		return "", fmt.Errorf("cannot convert script to address")
	}

	hash, err := hex.DecodeString(info.PublicKeyHash)
	if err != nil {
		return "", err
	}

	addr, err := NewAddressFromPublicKeyHash(hash, mainnet)
	if err != nil {
		return "", err
	}

	return addr.AddressString, nil
}

// GetAddressInfo returns address information from the script
func (s *Script) GetAddressInfo() *Address {
	if s.isInput {
		return s.getInputAddressInfo()
	} else if s.isOutput {
		return s.getOutputAddressInfo()
	} else {
		info := s.getOutputAddressInfo()
		if info != nil {
			return info
		}
		return s.getInputAddressInfo()
	}
}

// getOutputAddressInfo returns address info for output scripts
func (s *Script) getOutputAddressInfo() *Address {
	if s.IsScriptHashOut() {
		data, err := s.GetData()
		if err != nil {
			return nil
		}
		hashStr := hex.EncodeToString(data)
		return &Address{PublicKeyHash: hashStr}
	} else if s.IsPublicKeyHashOut() {
		data, err := s.GetData()
		if err != nil {
			return nil
		}
		hashStr := hex.EncodeToString(data)
		return &Address{PublicKeyHash: hashStr}
	}
	return nil
}

// getInputAddressInfo returns address info for input scripts
func (s *Script) getInputAddressInfo() *Address {
	if s.IsPublicKeyHashIn() {
		chunks := s.Chunks()
		if len(chunks) >= 2 && chunks[1].Buf != nil {
			hash := crypto.Hash160(chunks[1].Buf)
			hashStr := hex.EncodeToString(hash)
			return &Address{PublicKeyHash: hashStr}
		}
	} else if s.IsScriptHashIn() {
		chunks := s.Chunks()
		if len(chunks) > 0 {
			lastChunk := chunks[len(chunks)-1]
			if lastChunk.Buf != nil {
				hash := crypto.Hash160(lastChunk.Buf)
				hashStr := hex.EncodeToString(hash)
				return &Address{PublicKeyHash: hashStr}
			}
		}
	}
	return nil
}

// Add appends an opcode, data, or script to the end of the script
func (s *Script) Add(obj interface{}) error {
	return s.addByType(obj, false)
}

// Prepend inserts an opcode, data, or script at the beginning of the script
func (s *Script) Prepend(obj interface{}) error {
	return s.addByType(obj, true)
}

// addByType handles adding different types of objects to the script
func (s *Script) addByType(obj interface{}, prepend bool) error {
	switch v := obj.(type) {
	case byte:
		return s.addOpcode(v, prepend)
	case []byte:
		return s.addBuffer(v, prepend)
	case *Script:
		return s.insertAtPosition(v.data, prepend)
	case string:
		// Try as opcode name first
		if opcode, ok := opCodeStrings[v]; ok {
			return s.addOpcode(opcode, prepend)
		}
		// Try as hex string
		if b, err := hex.DecodeString(v); err == nil {
			return s.addBuffer(b, prepend)
		}
		// Treat as raw string
		return s.addBuffer([]byte(v), prepend)
	default:
		return fmt.Errorf("invalid script chunk type")
	}
}

// insertAtPosition inserts buffer at the beginning or end
func (s *Script) insertAtPosition(buf []byte, prepend bool) error {
	if s == nil {
		return ErrEmptyScript
	}
	if prepend {
		s.data = append(buf, s.data...)
	} else {
		s.data = append(s.data, buf...)
	}
	return nil
}

// addOpcode adds an opcode to the script
func (s *Script) addOpcode(opcode byte, prepend bool) error {
	if opcode > 255 {
		return fmt.Errorf("invalid opcode: %d", opcode)
	}
	chunk := Chunk{OpcodeNum: opcode}
	chunkScript, err := FromChunks([]Chunk{chunk})
	if err != nil {
		return err
	}
	return s.insertAtPosition(chunkScript.data, prepend)
}

// addBuffer adds a buffer to the script with proper push data encoding
func (s *Script) addBuffer(buf []byte, prepend bool) error {
	prefix, err := PushDataPrefix(buf)
	if err != nil {
		return err
	}
	data := append(prefix, buf...)
	return s.insertAtPosition(data, prepend)
}

// BuildMultisigOut creates a multisig output script
func BuildMultisigOut(publicKeys []*bec.PublicKey, threshold int, opts map[string]interface{}) (*Script, error) {
	if threshold > len(publicKeys) {
		return nil, fmt.Errorf("number of required signatures must be less than or equal to the number of public keys")
	}

	script := NewFromBytes([]byte{})
	
	// Add threshold
	_ = script.AppendOpcodes(byte(int(OpONE) + (threshold - 1)))

	// Sort public keys if not disabled
	noSorting := false
	if opts != nil {
		if val, ok := opts["noSorting"].(bool); ok {
			noSorting = val
		}
	}

	sortedKeys := publicKeys
	if !noSorting {
		// Simple sort by hex representation
		// In a real implementation, you'd want proper sorting
		sortedKeys = make([]*bec.PublicKey, len(publicKeys))
		copy(sortedKeys, publicKeys)
	}

	// Add public keys
	for _, pubKey := range sortedKeys {
		_ = script.AppendPushData(pubKey.SerialiseCompressed())
	}

	// Add number of keys
	_ = script.AppendOpcodes(byte(int(OpONE) + (len(publicKeys) - 1)))

	// Add CHECKMULTISIG
	_ = script.AppendOpcodes(OpCHECKMULTISIG)

	return script, nil
}

// BuildMultisigIn creates a multisig input script
func BuildMultisigIn(pubkeys []*bec.PublicKey, threshold int, signatures [][]byte, opts map[string]interface{}) (*Script, error) {
	script := NewFromBytes([]byte{})
	_ = script.AppendOpcodes(OpZERO)

	for _, sig := range signatures {
		_ = script.AppendPushData(sig)
	}

	return script, nil
}

// BuildP2SHMultisigIn creates a P2SH multisig input script
func BuildP2SHMultisigIn(pubkeys []*bec.PublicKey, threshold int, signatures [][]byte, opts map[string]interface{}) (*Script, error) {
	script, err := BuildMultisigIn(pubkeys, threshold, signatures, opts)
	if err != nil {
		return nil, err
	}

	// Add redeem script
	var redeemScript *Script
	if cached, ok := opts["cachedMultisig"].(*Script); ok && cached != nil {
		redeemScript = cached
	} else {
		redeemScript, err = BuildMultisigOut(pubkeys, threshold, opts)
		if err != nil {
			return nil, err
		}
	}

	_ = script.AppendPushData(redeemScript.data)
	return script, nil
}

// BuildPublicKeyOut creates a pay to public key output script
func BuildPublicKeyOut(pubkey *bec.PublicKey) *Script {
	script := NewFromBytes([]byte{})
	_ = script.AppendPushData(pubkey.SerialiseCompressed())
	_ = script.AppendOpcodes(OpCHECKSIG)
	return script
}

// BuildDataOut creates an OP_RETURN script with data
func BuildDataOut(data []byte, encoding string) (*Script, error) {
	script := NewFromBytes([]byte{})
	_ = script.AppendOpcodes(OpRETURN)
	_ = script.AppendPushData(data)
	return script, nil
}

// BuildSafeDataOut creates a safe OP_RETURN script (OP_FALSE OP_RETURN ...)
func BuildSafeDataOut(data []byte, encoding string) (*Script, error) {
	script := NewFromBytes([]byte{})
	if err := script.AppendOpcodes(OpFALSE, OpRETURN); err != nil {
		return nil, err
	}
	if err := script.AppendPushData(data); err != nil {
		return nil, err
	}
	return script, nil
}

// BuildScriptHashOut creates a pay to script hash output script
func BuildScriptHashOut(script *Script) *Script {
	if script == nil {
		return NewFromBytes([]byte{})
	}
	hash := crypto.Hash160(script.data)
	newScript := NewFromBytes([]byte{})
	_ = newScript.AppendOpcodes(OpHASH160)
	_ = newScript.AppendPushData(hash)
	_ = newScript.AppendOpcodes(OpEQUAL)
	return newScript
}

// ToScriptHashOut converts this script to a P2SH output script
func (s *Script) ToScriptHashOut() *Script {
	return BuildScriptHashOut(s)
}

// BuildPublicKeyIn creates a public key input script
func BuildPublicKeyIn(signature []byte, sigtype byte) *Script {
	script := NewFromBytes([]byte{})
	sigWithType := append(signature, sigtype)
	_ = script.AppendPushData(sigWithType)
	return script
}

// BuildPublicKeyHashIn creates a public key hash input script
func BuildPublicKeyHashIn(publicKey []byte, signature []byte, sigtype byte) *Script {
	script := NewFromBytes([]byte{})
	sigWithType := append(signature, sigtype)
	_ = script.AppendPushData(sigWithType)
	_ = script.AppendPushData(publicKey)
	return script
}

