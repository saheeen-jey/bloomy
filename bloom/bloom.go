// Package bloom implements a simple, dependency-free Bloom filter.
//
// A Bloom filter is a space-efficient probabilistic data structure used to
// test whether an element is a member of a set. False positives are
// possible ("this item is probably in the set"), but false negatives are
// not ("this item is definitely not in the set").
package bloom

import (
	"errors"
	"hash/fnv"
	"math"
)

// ErrIncompatibleFilters is returned when attempting to combine two
// filters that don't share the same bit-array size and hash count.
var ErrIncompatibleFilters = errors.New("bloom: filters must have the same m and k to be combined")

// Filter is a Bloom filter backed by a bit array of m bits and k hash
// functions. The zero value is not usable; construct one with New or
// NewWithEstimates.
type Filter struct {
	bits  []uint64 // bit array, packed 64 bits per word
	m     uint     // number of bits in the array
	k     uint     // number of hash functions
	count uint64   // approximate number of items added (informational only)
}

// New creates a Filter with an explicit bit-array size m and hash count k.
// Most callers should prefer NewWithEstimates, which derives m and k from
// the expected number of items and a target false-positive rate.
func New(m, k uint) *Filter {
	if m == 0 {
		m = 1
	}
	if k == 0 {
		k = 1
	}
	numWords := (m + 63) / 64
	return &Filter{
		bits: make([]uint64, numWords),
		m:    m,
		k:    k,
	}
}

// NewWithEstimates creates a Filter sized for n expected items with a
// target false-positive rate fp (e.g. 0.01 for 1%). The resulting m and k
// are chosen using the standard Bloom filter sizing formulas.
func NewWithEstimates(n uint, fp float64) *Filter {
	m, k := EstimateParameters(n, fp)
	return New(m, k)
}

// EstimateParameters returns the bit-array size m and hash count k that
// minimize memory use for n expected items at a target false-positive
// rate fp.
func EstimateParameters(n uint, fp float64) (m uint, k uint) {
	if n == 0 {
		n = 1
	}
	if fp <= 0 || fp >= 1 {
		fp = 0.01
	}
	mf := math.Ceil(-1 * float64(n) * math.Log(fp) / (math.Ln2 * math.Ln2))
	kf := math.Round((mf / float64(n)) * math.Ln2)
	if kf < 1 {
		kf = 1
	}
	return uint(mf), uint(kf)
}

// Add inserts data into the filter.
func (f *Filter) Add(data []byte) {
	h1, h2 := baseHashes(data)
	for i := uint(0); i < f.k; i++ {
		pos := f.location(h1, h2, i)
		f.setBit(pos)
	}
	f.count++
}

// Test reports whether data is possibly in the set. A false result means
// data is definitely not in the set. A true result may be a false
// positive.
func (f *Filter) Test(data []byte) bool {
	h1, h2 := baseHashes(data)
	for i := uint(0); i < f.k; i++ {
		pos := f.location(h1, h2, i)
		if !f.getBit(pos) {
			return false
		}
	}
	return true
}

// Union merges other into f in place, so that f.Test reports true for
// anything either filter would have reported true for. Both filters must
// share the same m and k, otherwise ErrIncompatibleFilters is returned.
func (f *Filter) Union(other *Filter) error {
	if f.m != other.m || f.k != other.k {
		return ErrIncompatibleFilters
	}
	for i := range f.bits {
		f.bits[i] |= other.bits[i]
	}
	f.count += other.count
	return nil
}

// EstimatedFalsePositiveRate returns the approximate probability that
// Test will return true for an item that was never Added, given the
// number of items added so far.
func (f *Filter) EstimatedFalsePositiveRate() float64 {
	n := float64(f.count)
	m := float64(f.m)
	k := float64(f.k)
	if m == 0 {
		return 1
	}
	return math.Pow(1-math.Exp(-k*n/m), k)
}

// M returns the number of bits in the underlying bit array.
func (f *Filter) M() uint { return f.m }

// K returns the number of hash functions used per item.
func (f *Filter) K() uint { return f.k }

// Count returns the number of times Add has been called. Adding the same
// item twice counts twice, so this is an upper bound on the number of
// distinct items, not an exact count.
func (f *Filter) Count() uint64 { return f.count }

func (f *Filter) setBit(pos uint) {
	f.bits[pos/64] |= 1 << (pos % 64)
}

func (f *Filter) getBit(pos uint) bool {
	return f.bits[pos/64]&(1<<(pos%64)) != 0
}

// location computes the bit position for the i-th hash function using
// double hashing (Kirsch-Mitzenmacher), which lets us derive k hash
// values from just two independent hashes.
func (f *Filter) location(h1, h2 uint64, i uint) uint {
	return uint((h1 + uint64(i)*h2) % uint64(f.m))
}

// baseHashes computes two independent 64-bit hashes of data using two
// different FNV variants, avoiding any external dependency.
func baseHashes(data []byte) (uint64, uint64) {
	h1 := fnv.New64a()
	h1.Write(data)
	sum1 := h1.Sum64()

	h2 := fnv.New64()
	h2.Write(data)
	sum2 := h2.Sum64()

	return sum1, sum2
}
