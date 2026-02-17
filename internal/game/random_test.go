package game

import "testing"

func TestSeededRNGDeterministic(t *testing.T) {
	rngA := seededRNG(12345)
	rngB := seededRNG(12345)

	for i := 0; i < 20; i++ {
		gotA := rngA.IntN(100000)
		gotB := rngB.IntN(100000)
		if gotA != gotB {
			t.Fatalf("expected deterministic sequence, mismatch at %d: %d != %d", i, gotA, gotB)
		}
	}
}

func TestSeedWordChangesWithSalt(t *testing.T) {
	a := seedWord(99, "a")
	b := seedWord(99, "b")
	if a == b {
		t.Fatalf("expected different seed words for different salts")
	}
}
