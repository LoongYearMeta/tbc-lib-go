package tbc

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/pkg/errors"

	"github.com/LoongYearMeta/tbc-lib-go/bscript"
	"github.com/LoongYearMeta/tbc-lib-go/encoding"
)

/*
General format (inside a block) of each output of a transaction - Txout
Field	                        Description	                                Size
-----------------------------------------------------------------------------------------------------
value                         non-negative integer giving the number of   8 bytes
                              Satoshis(BTC/10^8) to be transferred
Txout-script length           non-negative integer                        1 - 9 bytes VI = encoding.VarInt
Txout-script / scriptPubKey   Script                                      <out-script length>-many bytes
(lockingScript)

*/

// Output is a representation of a transaction output
type Output struct {
	Satoshis      uint64
	LockingScript *bscript.Script
}

// ReadFrom reads from the `io.Reader` into the `tbc.Output`.
func (o *Output) ReadFrom(r io.Reader) (int64, error) {
	*o = Output{}
	var bytesRead int64

	satoshis := make([]byte, 8)
	n, err := io.ReadFull(r, satoshis)
	bytesRead += int64(n)
	if err != nil {
		return bytesRead, errors.Wrapf(err, "satoshis(8): got %d bytes", n)
	}

	var l encoding.VarInt
	n64, err := l.ReadFrom(r)
	bytesRead += n64
	if err != nil {
		return bytesRead, err
	}

	script := make([]byte, l)
	n, err = io.ReadFull(r, script)
	bytesRead += int64(n)
	if err != nil {
		return bytesRead, errors.Wrapf(err, "lockingScript(%d): got %d bytes", l, n)
	}

	o.Satoshis = binary.LittleEndian.Uint64(satoshis)
	o.LockingScript = bscript.NewFromBytes(script)

	return bytesRead, nil
}

// LockingScriptHexString returns the locking script
// of an output encoded as a hex string.
func (o *Output) LockingScriptHexString() string {
	return hex.EncodeToString(o.LockingScript.Bytes())
}

func (o *Output) String() string {
	return fmt.Sprintf(`value:     %d
scriptLen: %d
script:    %s
`, o.Satoshis, o.LockingScript.Len(), o.LockingScript)
}

// Bytes encodes the Output into a byte array.
func (o *Output) Bytes() []byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, o.Satoshis)

	h := make([]byte, 0)
	h = append(h, b...)
	h = append(h, encoding.VarInt(uint64(o.LockingScript.Len())).Bytes()...)
	h = append(h, o.LockingScript.Bytes()...)

	return h
}

// BytesForSigHash returns the proper serialisation
// of an output to be hashed and signed (sighash).
func (o *Output) BytesForSigHash() []byte {
	buf := make([]byte, 0)

	satoshis := make([]byte, 8)
	binary.LittleEndian.PutUint64(satoshis, o.Satoshis)
	buf = append(buf, satoshis...)

	buf = append(buf, encoding.VarInt(uint64(o.LockingScript.Len())).Bytes()...)
	buf = append(buf, o.LockingScript.Bytes()...)

	return buf
}

// NodeJSON returns a wrapped *tbc.Output for marshalling/unmarshalling into a node output format.
//
// Marshalling usage example:
//  bb, err := json.Marshal(output.NodeJSON())
//
// Unmarshalling usage example:
//  output := &tbc.Output{}
//  if err := json.Unmarshal(bb, output.NodeJSON()); err != nil {}
func (o *Output) NodeJSON() interface{} {
	return &nodeOutputWrapper{Output: o}
}
