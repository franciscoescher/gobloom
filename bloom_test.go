package gobloom

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBloomFilter_AddAndTest(t *testing.T) {
	t.Parallel()
	n := uint64(10000)
	p := 0.01
	bf, err := New(Params{N: n, FalsePositiveRate: p})
	assert.NoError(t, err, "Failed to create Bloom filter")

	item := "test-item"
	bf.Add([]byte(item))

	b, err := bf.Test([]byte(item))
	assert.NoError(t, err, "Failed to test item '%s'", item)
	assert.True(t, b, "Item '%s' should be present in the Bloom filter, but it's not.", item)
}
func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	t.Parallel()
	n := uint64(1000000)
	p := 0.00100
	bf, err := New(Params{N: n, FalsePositiveRate: p})
	assert.NoError(t, err, "Failed to create Bloom filter")

	for i := uint64(0); i < n; i++ {
		item := fmt.Sprintf("test-item-%d", i)
		bf.Add([]byte(item))
	}

	falsePositives := 0
	for i := uint64(0); i < n; i++ {
		item := fmt.Sprintf("different-item-%d", i)
		b, err := bf.Test([]byte(item))
		assert.NoError(t, err, "Failed to test item '%s'", item)
		if b {
			falsePositives++
		}
	}

	estimatedFalsePositiveRate := float64(falsePositives) / float64(n)
	assert.LessOrEqual(t, estimatedFalsePositiveRate, p*1.15, "Estimated false positive rate is higher than expected")
}

func TestBloomFilter_AddTestNonExistentItem(t *testing.T) {
	t.Parallel()
	n := uint64(1000)
	p := 0.01
	bf, err := New(Params{N: n, FalsePositiveRate: p})
	assert.NoError(t, err, "Failed to create Bloom filter")

	b, err := bf.Test([]byte("non-existent-item"))
	assert.NoError(t, err, "Failed to test non-existent item")
	assert.False(t, b, "Non-existent item should not be present in the Bloom filter.")
}
