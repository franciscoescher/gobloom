package gobloom

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMurmur(t *testing.T) {
	n := uint64(10)
	h := NewMurMur3Hasher()
	hashers := h.GetHashes(n)

	assert.Equal(t, int(n), len(hashers), "Expected %d hashers, got %d", n, len(hashers))
	for i, hasher := range hashers {
		assert.NotNil(t, hasher, "Expected hasher at index %d to not be nil", i)
	}
}
