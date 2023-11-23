package gobloom

import "testing"

func TestMurmur(t *testing.T) {
	n := uint64(10)
	h := NewMurMur3Hasher()
	hashers := h.GetHashes(n)
	if len(hashers) != int(n) {
		t.Errorf("Expected %d hashers, got %d", n, len(hashers))
	}
	for i := 0; i < len(hashers); i++ {
		if hashers[i] == nil {
			t.Errorf("Expected hasher at index %d to not be nil", i)
		}
	}
}
