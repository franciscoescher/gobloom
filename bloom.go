package gobloom

import (
	"fmt"
	"hash"
	"math"
)

var _ Interface = (*BloomFilter)(nil)

// LockType represents the type of lock to use.
type LockType int

const (
	Default LockType = iota
	NoLock
	ExclusiveLock
	ReadWriteLock
)

// BloomFilter represents a single Bloom filter structure.
type BloomFilter struct {
	m      uint64        // The number of bits in the bit set
	bitSet []uint64      // The bit array represented as a slice of uint64
	k      uint64        // The number of hash functions to use
	hashes []hash.Hash64 // The hash functions to use
	mutex  Mutex         // Mutex to ensure thread safety
}

// Params represents the parameters for creating a new Bloom filter.
type Params struct {
	// N is the number of elements expected to be added to the Bloom filter.
	N uint64
	// FalsePositiveRate is the acceptable false positive rate.
	FalsePositiveRate float64
	// Hasher is the hash provider to use. Defaults to MurMur3Hasher.
	Hasher Hasher
	// LockType is the lock type to use. Defaults to ReadLock.
	// LockType is the lock type to use. Defaults to ExclusiveLock.
	// The use of ReadWriteLock can improve performance when there are many concurrent reads.
	// If you have much more writes, avoid using ReadWriteLock, cause it may lead to reader starvation.
	LockType LockType
}

// New creates a new Bloom filter with the given number of elements (n) and false positive rate (p).
func New(p Params) (*BloomFilter, error) {
	applyDefaults(&p)
	if p.N == 0 {
		return nil, fmt.Errorf("number of elements cannot be 0")
	}
	if p.FalsePositiveRate <= 0 || p.FalsePositiveRate >= 1 {
		return nil, fmt.Errorf("false positive rate must be between 0 and 1")
	}
	if p.Hasher == nil {
		return nil, fmt.Errorf("hasher cannot be nil")
	}
	m, k := getOptimalParams(p.N, p.FalsePositiveRate)
	bitSetSize := (m + 63) / 64 // Round up to the nearest 64 bits
	mu, err := NewMutex(p.LockType)
	if err != nil {
		return nil, err
	}
	return &BloomFilter{
		m:      m,
		k:      k,
		bitSet: make([]uint64, bitSetSize),
		hashes: p.Hasher.GetHashes(k),
		mutex:  mu,
	}, nil
}

// applyDefaults applies the default values to the parameters if they are not set.
func applyDefaults(p *Params) {
	if p.Hasher == nil {
		p.Hasher = NewMurMur3Hasher()
	}
	if p.LockType == Default {
		p.LockType = ExclusiveLock
	}
}

// getOptimalParams calculates the optimal parameters for the Bloom filter,
// the number of bits in the bit set (m) and the number of hash functions (k).
func getOptimalParams(n uint64, p float64) (uint64, uint64) {
	m := uint64(math.Ceil(-1 * float64(n) * math.Log(p) / math.Pow(math.Log(2), 2)))
	if m == 0 {
		m = 1
	}
	k := uint64(math.Ceil((float64(m) / float64(n)) * math.Log(2)))
	if k == 0 {
		k = 1
	}
	return m, k
}

// Add adds an item to the Bloom filter.
func (bf *BloomFilter) Add(data []byte) {
	if bf.mutex != nil {
		bf.mutex.Lock()
		defer bf.mutex.Unlock()
	}
	for _, hash := range bf.hashes {
		hash.Reset()
		hash.Write(data)
		hashValue := hash.Sum64() % bf.m
		index := hashValue / 64    // Find the index in the bitSet
		position := hashValue % 64 // Find the position in the uint64
		bf.bitSet[index] |= 1 << position
	}
}

// Test checks if an item is in the Bloom filter.
func (bf *BloomFilter) Test(data []byte) bool {
	if bf.mutex != nil {
		bf.mutex.RLock()
		defer bf.mutex.RUnlock()
	}
	for _, hash := range bf.hashes {
		hash.Reset()
		hash.Write(data)
		hashValue := hash.Sum64() % bf.m
		index := hashValue / 64    // Find the index in the bitSet
		position := hashValue % 64 // Find the position in the uint64
		if bf.bitSet[index]&(1<<position) == 0 {
			return false
		}
	}
	return true
}
