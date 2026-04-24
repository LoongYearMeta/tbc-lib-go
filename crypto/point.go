package crypto

// Point aliases the bec public key point representation.
// Kept minimal: callers that need arithmetic should pull bec directly.
type Point = PublicKey
