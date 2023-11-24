package gobloom

import (
	"errors"
	"fmt"
	"math" // Used for calculations needed by the Bloom filter
	// Used for arbitrary-precision arithmetic for the bit set
)

var _ Interface = (*ScalableBloomFilter)(nil)

// ScalableBloomFilter combines multiple BloomFilter slices to adapt to a growing number of elements.
type ScalableBloomFilter struct {
	filters  []*BloomFilter // A slice of BloomFilter pointers, representing each layer of the scalable filter
	n        uint64         // The number of items that have been added
	fpRate   float64        // The false positive rate for the current filter slice
	fpGrowth float64        // Factor by which the false positive probability should increase for each additional filter slice
}

// NewScalableBloomFilter initializes a new scalable Bloom filter with an estimated initial size,
// target false positive rate, and growth rate for the false positive probability with each new filter slice.
// It is essential to carefully select these values based on the use-case requirements to maintain
// system performance and desired accuracy as the dataset grows.
//
// Parameters:
//
//   - initialSize (uint64): The estimated number of elements you expect to store in the bloom filter initially.
//     This is not a hard limit but rather a guideline for preparing the initial Bloom filter layer.
//     A size that's too small could lead to rapid addition of new slices, increasing memory usage,
//     and a size that's too large could waste memory upfront.
//
//   - fpRate (float64): The desired false positive probability for the first Bloom filter slice.
//     It determines how likely it is that a 'Test' operation falsely indicates the presence of an element.
//     A smaller fpRate will increase the number of bits used in the initial filter (decreasing the chance of false positives),
//     but also increase consumption of memory. Typical values are between 0.01 and 0.001 (1% - 0.1%).
//
//   - fpGrowth (float64): The growth rate of the false positive probability with the addition of each subsequent filter slice.
//     If set to a value greater than 1, each new filter layer tolerates a higher false positive rate,
//     which can be useful to postpone the addition of new layers and control memory usage.
//     Commonly, this parameter should be set to a value close to 1 (e.g. 1.5 - 2),
//     as a higher growth rate could lead to a rapidly deteriorating false positive rate.
//
// Best Practices:
//
//   - Conduct a pre-analysis based on expected data growth to estimate initialSize and what fpGrowth rate could be appropriate.
//     Consider both current and future memory availability and access patterns of the data.
//
//   - Create logging or metrics around the scalable Bloom filter behavior, monitoring the number of slices created
//     and memory consumption over time, which can provide insights into usage patterns and need for adjustment.
//
//   - Avoid using very small fpGrowth values, as these will lead to frequent addition of filter slices,
//     compounding memory usage and potentially causing performance bottlenecks as the 'Test' function
//     needs to check more slices.
//
//   - In distributed systems where consistency across nodes is necessary, ensure that all instances
//     are initialized with the same parameters and handle the updates to filter slices identically.
//
//   - As with any Bloom filter, using multiple, independent hash functions improves the spread of elements
//     across the bit set. In a Scalable Bloom Filter, these hash functions need to maintain their properties
//     as additional layers are added.
func NewScalableBloomFilterWithHasher(initialSize uint64, fpRate float64, fpGrowth float64, h Hasher) (*ScalableBloomFilter, error) {
	if initialSize <= 0 {
		return nil, errors.New("invalid initial size, must be greater than 0")
	}
	if fpRate <= 0 || fpRate >= 1 {
		return nil, fmt.Errorf("invalid false positive rate, must be between 0 and 1, got %f", fpRate)
	}
	if fpGrowth <= 0 {
		return nil, fmt.Errorf("invalid false positive growth rate, must be greater than 0, got %f", fpGrowth)
	}

	bf, err := NewBloomFilterWithHasher(initialSize, fpRate, h)
	if err != nil {
		return nil, err
	}

	// Return a new scalable Bloom filter struct with the initialized slice and parameters.
	return &ScalableBloomFilter{
		filters:  []*BloomFilter{bf}, // Start with one filter slice
		fpRate:   fpRate,             // Set the initial false positive rate
		fpGrowth: fpGrowth,           // Set the growth rate for false positives as the filter scales
		n:        0,                  // Initialize with zero elements added
	}, nil
}

// NewScalableBloomFilter initializes a new scalable Bloom filter with the standard MurMur3 hasher.
func NewScalableBloomFilter(initialSize uint64, fpRate float64, fpGrowth float64) (*ScalableBloomFilter, error) {
	return NewScalableBloomFilterWithHasher(initialSize, fpRate, fpGrowth, NewMurMur3Hasher())
}

// Add inserts the given item into the scalable Bloom filter.
// If the current filter slice exceeds its capacity based on the growth rate, a new slice is added.
func (sbf *ScalableBloomFilter) Add(data []byte) {
	// Add the item to all existing filter slices.
	for _, filter := range sbf.filters {
		for i := uint64(0); i < filter.k; i++ {
			filter.Add(data)
		}
	}

	// Increment the total number of items added across all filter slices.
	sbf.n++

	// Check the last filter's capacity, and if needed, add a new filter slice.
	// The current implementation may be only adding new filters and not previous ones.
	// We need to base the condition on the properties of the last filter and the current count of added items.
	currentFilter := sbf.filters[len(sbf.filters)-1] // Get the most recent filter

	// Use the correct threshold to decide when to add a new filter.
	// This threshold should be defined by how full the current filter is.
	currentCapacity := float64(currentFilter.m) * math.Log(sbf.fpGrowth) / math.Log(2)
	// If the number of items exceeds the current capacity of the filter:
	if float64(sbf.n) > currentCapacity {
		newFpRate := sbf.fpRate * math.Pow(sbf.fpGrowth, float64(len(sbf.filters)))
		// Create and append the new filter slice.
		nbf, _ := NewBloomFilter(sbf.n, newFpRate)
		sbf.filters = append(sbf.filters, nbf)
	}
}

func (sbf *ScalableBloomFilter) Test(data []byte) bool {
	// Check the item against all filter slices from the oldest to the newest.
	for _, filter := range sbf.filters {
		// Assume the item is in the filter until proven otherwise.
		isPresent := true

		// Use the same hash functions that were used to add the item.
		for i := uint64(0); i < filter.k; i++ {
			// If any of the bits corresponding to the item's hash values are not set, it's definitely not present in this filter.
			if !filter.Test(data) {
				isPresent = false
				// We break out of the hash function loop as soon as we find a bit that is not set.
				break
			}
		}

		// If all the bits for this filter are set, then the item is potentially present (with some false positive rate).
		if isPresent {
			return true
		}
		// Otherwise, continue checking the next filter to see if the item may be present there.
	}

	// If none of the filters had all bits set, the item is definitely not in the set.
	return false
}

// TestString is a convenience function for testing a string item.
func (sbf *ScalableBloomFilter) TestString(item string) bool {
	return sbf.Test([]byte(item))
}

// AddString is a convenience function for adding a string item.
func (sbf *ScalableBloomFilter) AddString(item string) {
	sbf.Add([]byte(item))
}
