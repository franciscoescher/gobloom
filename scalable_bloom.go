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

// ParamsScalable represents the parameters for creating a new scalable Bloom filter.
type ParamsScalable struct {
	// InitialSize is the estimated number of elements you expect to store in the bloom filter initially.
	// This is not a hard limit but rather a guideline for preparing the initial Bloom filter layer.
	// A size that's too small could lead to rapid addition of new slices, increasing memory usage,
	// and a size that's too large could waste memory upfront.
	InitialSize uint64
	// FalsePositiveRate id the desired false positive probability for the first Bloom filter slice.
	// It determines how likely it is that a 'Test' operation falsely indicates the presence of an element.
	// A smaller false positive rate will increase the number of bits used in the initial filter (decreasing the chance of false positives),
	// but also increase consumption of memory. Typical values are between 0.01 and 0.001 (1% - 0.1%).
	FalsePositiveRate float64
	// FalsePositiveGrowth is the growth rate of the false positive probability with the addition of each subsequent filter slice.
	// If set to a value greater than 1, each new filter layer tolerates a higher false positive rate,
	// which can be useful to postpone the addition of new layers and control memory usage.
	// Commonly, this parameter should be set to a value close to 1 (e.g. 1.5 - 2),
	// as a higher growth rate could lead to a rapidly deteriorating false positive rate.
	FalsePositiveGrowth float64
	// Hasher is the hash provider to use. Defaults to MurMur3Hasher.
	Hasher Hasher
	// LockType is the lock type to use. Defaults to ExclusiveLock.
	// The use of ReadWriteLock can improve performance when there are many concurrent reads.
	// If you have much more writes, avoid using ReadWriteLock, cause it may lead to reader starvation.
	LockType LockType
}

// NewScalable creates a new scalable Bloom filter.
// Best Practices:
//
//   - Conduct a pre-analysis based on expected data growth to estimate initialSize and what false positive growth
//     rate could be appropriate. Consider both current and future memory availability and access patterns of
//     the data.
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
func NewScalable(p ParamsScalable) (*ScalableBloomFilter, error) {
	applyDefaultsScalable(&p)
	if p.InitialSize <= 0 {
		return nil, errors.New("invalid initial size, must be greater than 0")
	}
	if p.FalsePositiveRate <= 0 || p.FalsePositiveRate >= 1 {
		return nil, fmt.Errorf("invalid false positive rate, must be between 0 and 1, got %f", p.FalsePositiveRate)
	}
	if p.FalsePositiveGrowth <= 0 {
		return nil, fmt.Errorf("invalid false positive growth rate, must be greater than 0, got %f", p.FalsePositiveGrowth)
	}

	bf, err := New(Params{
		N:                 p.InitialSize,
		FalsePositiveRate: p.FalsePositiveRate,
		Hasher:            p.Hasher,
		LockType:          p.LockType,
	})
	if err != nil {
		return nil, err
	}

	// Return a new scalable Bloom filter struct with the initialized slice and parameters.
	return &ScalableBloomFilter{
		filters:  []*BloomFilter{bf},    // Start with one filter slice
		fpRate:   p.FalsePositiveRate,   // Set the initial false positive rate
		fpGrowth: p.FalsePositiveGrowth, // Set the growth rate for false positives as the filter scales
		n:        0,                     // Initialize with zero elements added
	}, nil
}

// applyDefaultsScalable applies the default values to the parameters if they are not set.
func applyDefaultsScalable(p *ParamsScalable) {
	if p.Hasher == nil {
		p.Hasher = NewMurMur3Hasher()
	}
	if p.LockType == Default {
		p.LockType = ExclusiveLock
	}
}

// Add inserts the given item into the scalable Bloom filter.
// If the current filter slice exceeds its capacity based on the growth rate, a new slice is added.
func (sbf *ScalableBloomFilter) Add(data []byte) error {
	// Add the item to all existing filter slices.
	for _, filter := range sbf.filters {
		for i := uint64(0); i < filter.k; i++ {
			err := filter.Add(data)
			if err != nil {
				return err
			}
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
		nbf, _ := New(Params{N: sbf.n, FalsePositiveRate: newFpRate})
		sbf.filters = append(sbf.filters, nbf)
	}
	return nil
}

func (sbf *ScalableBloomFilter) Test(data []byte) (bool, error) {
	// Check the item against all filter slices from the oldest to the newest.
	for _, filter := range sbf.filters {
		// Assume the item is in the filter until proven otherwise.
		isPresent := true

		// Use the same hash functions that were used to add the item.
		for i := uint64(0); i < filter.k; i++ {
			// If any of the bits corresponding to the item's hash values are not set, it's definitely not present in this filter.
			b, err := filter.Test(data)
			if err != nil {
				return false, err
			}
			if !b {
				isPresent = false
				// We break out of the hash function loop as soon as we find a bit that is not set.
				break
			}
		}

		// If all the bits for this filter are set, then the item is potentially present (with some false positive rate).
		if isPresent {
			return true, nil
		}
		// Otherwise, continue checking the next filter to see if the item may be present there.
	}

	// If none of the filters had all bits set, the item is definitely not in the set.
	return false, nil
}
