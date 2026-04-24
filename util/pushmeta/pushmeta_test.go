package pushmeta_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"

	"github.com/LoongYearMeta/tbc-lib-go/transaction"
	"github.com/LoongYearMeta/tbc-lib-go/util/pushmeta"
)

func TestCurrentInputOutpointBytes(t *testing.T) {
	// Display-order txid (as in explorer / PreviousTxIDStr); wire order is reverse.
	displayHex := "0874930ab8db047123024944c23f2981937b4a92db91c1658ef1e2adc539df1b"
	displayID, err := hex.DecodeString(displayHex)
	if err != nil || len(displayID) != 32 {
		t.Fatal(err)
	}
	tx := transaction.NewTx()
	in := &transaction.Input{}
	if err := in.PreviousTxIDAdd(displayID); err != nil {
		t.Fatal(err)
	}
	in.PreviousTxOutIndex = 0
	in.SequenceNumber = 0xfffffffd // common for CLTV-style; any LE value
	tx.Inputs = []*transaction.Input{in}

	got := pushmeta.CurrentInputOutpointBytes(tx, 0)
	if len(got) != 40 {
		t.Fatalf("len=%d", len(got))
	}
	want := make([]byte, 40)
	copy(want[:32], reverseCopy(displayID))
	binary.LittleEndian.PutUint32(want[32:36], in.PreviousTxOutIndex)
	binary.LittleEndian.PutUint32(want[36:40], in.SequenceNumber)
	if !bytes.Equal(got, want) {
		t.Fatalf("got %x want %x", got, want)
	}
}

func reverseCopy(b []byte) []byte {
	out := make([]byte, len(b))
	copy(out, b)
	for i := 0; i < len(out)/2; i++ {
		out[i], out[len(out)-1-i] = out[len(out)-1-i], out[i]
	}
	return out
}
