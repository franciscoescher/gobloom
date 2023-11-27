package gobloom

import (
	"fmt"
	"testing"
)

func TestBloomFilter_AddAndTest(t *testing.T) {
	t.Parallel()
	// Test adding an item and then checking its presence
	n := uint64(10000)
	p := 0.01
	bf, err := New(Params{N: n, FalsePositiveRate: p})
	if err != nil {
		t.Errorf("Failed to create Bloom filter: %s", err)
	}

	item := "test-item"
	bf.Add([]byte(item))

	b, err := bf.Test([]byte(item))
	if err != nil {
		t.Errorf("Failed to test item '%s': %s", item, err)
	}
	if !b {
		t.Errorf("Item '%s' should be present in the Bloom filter, but it's not.", item)
	}
}

func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	t.Parallel()
	// Test that the Bloom filter maintains an acceptable false positive rate.
	// This is an empirical test and won't be 100% accurate.
	n := uint64(1000000)
	p := 0.00100
	bf, err := New(Params{N: n, FalsePositiveRate: p})
	if err != nil {
		t.Errorf("Failed to create Bloom filter: %s", err)
	}

	// Add n items.
	for i := uint64(0); i < n; i++ {
		item := fmt.Sprintf("test-item-%d", i)
		bf.Add([]byte(item))
	}

	// Check for a different set of n items to estimate the false positive rate.
	falsePositives := 0
	for i := uint64(0); i < n; i++ {
		item := fmt.Sprintf("different-item-%d", i)
		b, err := bf.Test([]byte(item))
		if err != nil {
			t.Errorf("Failed to test item '%s': %s", item, err)
		}
		if b {
			falsePositives++
		}
	}

	estimatedFalsePositiveRate := float64(falsePositives) / float64(n)
	if estimatedFalsePositiveRate > (p * 1.15) {
		t.Errorf("Estimated false positive rate is higher than expected: got %.3f%% (%d/%d), want <= %.3f%%",
			estimatedFalsePositiveRate*100, falsePositives, n, p*100)
	}
}

func TestBloomFilter_AddTestNonExistentItem(t *testing.T) {
	t.Parallel()
	// Test that a non-existent item returns false
	n := uint64(1000)
	p := 0.01
	bf, err := New(Params{N: n, FalsePositiveRate: p})
	if err != nil {
		t.Errorf("Failed to create Bloom filter: %s", err)
	}

	b, err := bf.Test([]byte("non-existent-item"))
	if err != nil {
		t.Errorf("Failed to test item: %s", err)
	}
	if b {
		t.Errorf("Non-existent item should not be present in the Bloom filter.")
	}
}
