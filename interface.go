package gobloom

import "hash"

type Interface interface {
	Add([]byte)
	AddString(string)
	Test([]byte) bool
	TestString(string) bool
}

// Hasher is an interface for a hash function that returns a slice of hash.Hash64.
type Hasher interface {
	GetHashes(n uint64) []hash.Hash64
}