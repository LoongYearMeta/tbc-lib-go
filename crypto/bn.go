package crypto

import "math/big"

// BN aliases math/big.Int to mirror JS lib/crypto/bn.js. Callers needing
// the full arithmetic API should import math/big directly.
type BN = big.Int
