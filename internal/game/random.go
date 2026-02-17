package game

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
)

func seededRNG(seed int64) *rand.Rand {
	// Non-cryptographic PRNG is intentional for deterministic simulation behavior.
	// #nosec G404
	return rand.New(rand.NewPCG(seedWord(seed, "a"), seedWord(seed, "b")))
}

func seedWord(seed int64, salt string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d:%s", seed, salt)))
	return h.Sum64()
}
