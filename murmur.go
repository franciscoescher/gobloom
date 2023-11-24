package gobloom

import (
	"hash"

	"github.com/spaolacci/murmur3"
)

type MurMur3Hasher struct{}

var _ Hasher = (*MurMur3Hasher)(nil)

func NewMurMur3Hasher() *MurMur3Hasher {
	return &MurMur3Hasher{}
}

func (h *MurMur3Hasher) GetHashes(n uint64) []hash.Hash64 {
	hashers := make([]hash.Hash64, n)
	for i := 0; uint64(i) < n; i++ {
		hashers[i] = murmur3.New64WithSeed(uint32(i))
	}
	return hashers
}
