// Package sighash is preserved as a forwarder for backwards compatibility.
// New code should import "github.com/LoongYearMeta/tbc-lib-go/transaction/sighash" directly.
package sighash

import sh "github.com/LoongYearMeta/tbc-lib-go/transaction/sighash"

// Flag is an alias for transaction/sighash.Flag.
type Flag = sh.Flag

// Re-exported SIGHASH flag constants.
const (
	Old          = sh.Old
	All          = sh.All
	None         = sh.None
	Single       = sh.Single
	AnyOneCanPay = sh.AnyOneCanPay

	AllForkID          = sh.AllForkID
	NoneForkID         = sh.NoneForkID
	SingleForkID       = sh.SingleForkID
	AnyOneCanPayForkID = sh.AnyOneCanPayForkID

	ForkID = sh.ForkID

	Mask = sh.Mask
)
