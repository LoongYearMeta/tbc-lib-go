package transaction

import (
	"context"

	"github.com/LoongYearMeta/tbc-lib-go/script"
	"github.com/LoongYearMeta/tbc-lib-go/transaction/sighash"
)

// UnlockerParams params used for unlocking an input with a `tbc.Unlocker`.
type UnlockerParams struct {
	// InputIdx the input to be unlocked. [DEFAULT 0]
	InputIdx uint32
	// SigHashFlags the be applied [DEFAULT ALL|FORKID]
	SigHashFlags sighash.Flag
}

// Unlocker interface to allow custom implementations of different unlocking mechanisms.
// Implement the Unlocker function as shown in LocalUnlocker, for example.
type Unlocker interface {
	UnlockingScript(ctx context.Context, tx *Tx, up UnlockerParams) (uscript *script.Script, err error)
}

// UnlockerGetter interfaces getting an unlocker for a given output/locking script.
type UnlockerGetter interface {
	Unlocker(ctx context.Context, lockingScript *script.Script) (Unlocker, error)
}
