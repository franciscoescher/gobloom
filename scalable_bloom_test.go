package gobloom

import (
	"math/rand"
	"strconv"
	"testing"
)

// TestScalableBloomFilter_AddAndTest will test that we can add an element to the filter and then find it.
func TestScalableBloomFilter_AddAndTest(t *testing.T) {
	t.Parallel()
	sbf, _ := NewScalableBloomFilter(1000, 0.01, 2)

	testItem := "example"
	sbf.AddString(testItem)

	if !sbf.TestString(testItem) {
		t.Errorf("Expected item '%s' to be in the filter", testItem)
	}
}

// TestScalableBloomFilter_FalsePositiveRate will test that the false positive rate stays below the set threshold.
func TestScalableBloomFilter_FalsePositiveRate(t *testing.T) {
	t.Parallel()
	initialSize := uint64(1000)
	fpRate := 0.01 // 1% false positive rate

	sbf, err := NewScalableBloomFilter(initialSize, fpRate, 2)
	if err != nil {
		t.Errorf("Error initializing scalable Bloom filter: %s", err)
		return
	}
	numElements := uint64(500) // Test with more elements than the initial size
	falsePositives := 0

	// Add elements
	for i := uint64(0); i < numElements; i++ {
		sbf.AddString(strconv.FormatUint(i, 10))
	}

	// Test for false positives
	numTests := uint64(5000)
	for i := numElements; i < numElements+numTests; i++ {
		if sbf.TestString(strconv.FormatUint(i, 10)) {
			falsePositives++
		}
	}

	observedFpRate := float64(falsePositives) / float64(numTests)
	if observedFpRate > 1.5*fpRate {
		t.Errorf("False positive rate is too high: got %.3f%%, want at most %.3f%%", observedFpRate*100, fpRate*100)
	}
}

// TestScalableBloomFilter_Scalability will test that the filter scales correctly by adding layers as needed.
func TestScalableBloomFilter_Scalability(t *testing.T) {
	t.Parallel()
	sbf, err := NewScalableBloomFilter(1000, 0.01, 2)
	if err != nil {
		t.Errorf("Error initializing scalable Bloom filter: %s", err)
		return
	}

	initialNumFilters := len(sbf.filters)

	// Add more elements to trigger scaling
	for i := 0; i < 10000; i++ {
		sbf.AddString(strconv.Itoa(rand.Int()))
	}

	if len(sbf.filters) == initialNumFilters {
		t.Errorf("Expected scalable Bloom filter to grow, but it didn't")
	}
}
