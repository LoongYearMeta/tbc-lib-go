package unlocker

import (
	"bytes"
	"context"
	"errors"

	"github.com/libsv/go-bk/bec"
	"github.com/libsv/go-bk/crypto"
	tbc "github.com/LoongYearMeta/tbc-lib-go"
	"github.com/LoongYearMeta/tbc-lib-go/bscript"
	"github.com/LoongYearMeta/tbc-lib-go/transaction/sighash"
)

// Getter implements the `tbc.UnlockerGetter` interface. It unlocks a Tx locally,
// using a bec PrivateKey.
type Getter struct {
	PrivateKey *bec.PrivateKey
}

// Unlocker builds a new `*unlocker.Local` with the same private key
// as the calling `*local.Getter`.
//
// For an example implementation, see `examples/unlocker_getter/`.
func (g *Getter) Unlocker(ctx context.Context, lockingScript *bscript.Script) (tbc.Unlocker, error) {
	return &Simple{PrivateKey: g.PrivateKey}, nil
}

// Simple implements the a simple `tbc.Unlocker` interface. It is used to build an unlocking script
// using a bec Private Key.
type Simple struct {
	PrivateKey *bec.PrivateKey
}

// UnlockingScript create the unlocking script for a given input using the PrivateKey passed in through the
// the `unlock.Local` struct.
//
// UnlockingScript generates and uses an ECDSA signature for the provided hash digest using the private key
// as well as the public key corresponding to the private key used. The produced
// signature is deterministic (same message and same key yield the same signature) and
// canonical in accordance with RFC6979 and BIP0062.
//
// For example usage, see `examples/create_tx/create_tx.go`
func (l *Simple) UnlockingScript(ctx context.Context, tx *tbc.Tx, params tbc.UnlockerParams) (*bscript.Script, error) {
	if params.SigHashFlags == 0 {
		params.SigHashFlags = sighash.AllForkID
	}

	prevScript := tx.Inputs[params.InputIdx].PreviousTxScript
	switch prevScript.ScriptType() {
	case bscript.ScriptTypePubKeyHash:
		// 与 tbc-lib-js tx.sign() 对齐：pkh 不匹配时静默跳过（返回空脚本），
		// 避免产生无效签名导致后续 OP_EQUALVERIFY 失败。
		if scriptPKH, err := prevScript.PublicKeyHash(); err == nil {
			keyPKH := crypto.Hash160(l.PrivateKey.PubKey().SerialiseCompressed())
			if !bytes.Equal(scriptPKH, keyPKH) {
				return bscript.NewFromBytes(nil), nil
			}
		}

		sh, err := tx.CalcInputSignatureHash(params.InputIdx, params.SigHashFlags)
		if err != nil {
			return nil, err
		}

		sig, err := l.PrivateKey.Sign(sh)
		if err != nil {
			return nil, err
		}

		pubKey := l.PrivateKey.PubKey().SerialiseCompressed()
		signature := sig.Serialise()

		uscript, err := bscript.NewP2PKHUnlockingScript(pubKey, signature, params.SigHashFlags)
		if err != nil {
			return nil, err
		}

		return uscript, nil
	}

	return nil, errors.New("currently only p2pkh supported")
}
