package gobloom

import (
	"math/rand"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScalableBloomFilter_AddAndTest(t *testing.T) {
	t.Parallel()
	sbf, _ := NewScalable(ParamsScalable{InitialSize: 1000, FalsePositiveRate: 0.01, FalsePositiveGrowth: 2})

	testItem := "example"
	sbf.Add([]byte(testItem))

	b, err := sbf.Test([]byte(testItem))
	assert.NoError(t, err, "Failed to test item '%s'", testItem)
	assert.True(t, b, "Expected item '%s' to be in the filter", testItem)
}

func TestScalableBloomFilter_FalsePositiveRate(t *testing.T) {
	t.Parallel()
	initialSize := uint64(1000)
	fpRate := 0.01 // 1% false positive rate

	sbf, err := NewScalable(ParamsScalable{InitialSize: initialSize, FalsePositiveRate: fpRate, FalsePositiveGrowth: 2})
	assert.NoError(t, err, "Error initializing scalable Bloom filter")

	numElements := uint64(500) // Test with more elements than the initial size
	falsePositives := 0

	// Add elements
	for i := uint64(0); i < numElements; i++ {
		sbf.Add([]byte(strconv.FormatUint(i, 10)))
	}

	// Test for false positives
	numTests := uint64(5000)
	for i := numElements; i < numElements+numTests; i++ {
		testItem := []byte(strconv.FormatUint(i, 10))
		b, err := sbf.Test(testItem)
		assert.NoError(t, err, "Failed to test item '%s'", testItem)
		if b {
			falsePositives++
		}
	}

	observedFpRate := float64(falsePositives) / float64(numTests)
	assert.LessOrEqual(t, observedFpRate, 1.15*fpRate, "False positive rate is too high: got %.3f%%, want at most %.3f%%", observedFpRate*100, fpRate*100)
}

func TestScalableBloomFilter_Scalability(t *testing.T) {
	t.Parallel()
	sbf, err := NewScalable(ParamsScalable{InitialSize: 1000, FalsePositiveRate: 0.01, FalsePositiveGrowth: 2})
	assert.NoError(t, err, "Error initializing scalable Bloom filter")

	initialNumFilters := len(sbf.filters)

	// Add more elements to trigger scaling
	for i := 0; i < 10000; i++ {
		sbf.Add([]byte(strconv.Itoa(rand.Int())))
	}

	assert.NotEqual(t, len(sbf.filters), initialNumFilters, "Expected scalable Bloom filter to grow, but it didn't")
}
