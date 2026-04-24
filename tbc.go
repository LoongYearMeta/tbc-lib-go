// Package tbc is the facade for the tbc-lib-go library. All historical tbc.X
// symbols are re-exported here from their new home in subpackages. Prefer
// importing the subpackage directly for new code, but this facade preserves
// backwards compatibility for every exported symbol that existed prior to
// the JS-style layout refactor.
//
// Design reference: docs/superpowers/specs/2026-04-24-js-style-layout-design.md
package tbc

import (
	"github.com/LoongYearMeta/tbc-lib-go/encoding"
	"github.com/LoongYearMeta/tbc-lib-go/networks"
)

// ======== networks ========

type Network = networks.Network

var (
	Livenet = networks.Livenet
	Testnet = networks.Testnet
	Regtest = networks.Regtest
	STN     = networks.STN

	AddNetwork    = networks.AddNetwork
	GetNetwork    = networks.GetNetwork
	RemoveNetwork = networks.RemoveNetwork
)

// DefaultNetwork is intentionally defined HERE rather than re-exported from
// networks/. A "var = pkg.Var" re-export would copy the value, so
// `tbc.DefaultNetwork = tbc.Testnet` would only mutate the root-package copy
// and silently drift from networks/. Keeping it in the root package preserves
// the original single-source-of-truth semantics.
var DefaultNetwork = Livenet

// ======== encoding ========

type VarInt = encoding.VarInt

var (
	NewVarIntFromBytes = encoding.NewVarIntFromBytes
	ReverseBytes       = encoding.ReverseBytes
	LittleEndianBytes  = encoding.LittleEndianBytes
)
