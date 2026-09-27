package bloom

import (
	"fmt"
	"testing"
)

func TestAddAndTest(t *testing.T) {
	f := NewWithEstimates(1000, 0.01)
	words := []string{"apple", "banana", "cherry"}

	for _, w := range words {
		f.Add([]byte(w))
	}

	for _, w := range words {
		if !f.Test([]byte(w)) {
			t.Errorf("expected %q to be reported present", w)
		}
	}
}

func TestNoFalseNegatives(t *testing.T) {
	const n = 500
	f := NewWithEstimates(n, 0.001)

	for i := 0; i < n; i++ {
		f.Add([]byte(fmt.Sprintf("item-%d", i)))
	}

	for i := 0; i < n; i++ {
		key := fmt.Sprintf("item-%d", i)
		if !f.Test([]byte(key)) {
			t.Fatalf("false negative for %q: Bloom filters must never report a false negative", key)
		}
	}
}

func TestFalsePositiveRateWithinBounds(t *testing.T) {
	const n = 1000
	const fp = 0.01

	f := NewWithEstimates(n, fp)
	for i := 0; i < n; i++ {
		f.Add([]byte(fmt.Sprintf("item-%d", i)))
	}

	const trials = 10000
	falsePositives := 0
	for i := 0; i < trials; i++ {
		key := fmt.Sprintf("nonexistent-%d", i)
		if f.Test([]byte(key)) {
			falsePositives++
		}
	}

	rate := float64(falsePositives) / float64(trials)
	// Bloom filters are probabilistic, so we allow a generous margin
	// above the target rate rather than asserting an exact match.
	const margin = 5
	if rate > fp*margin {
		t.Errorf("observed false positive rate %.4f exceeds %.4f (target %.4f x %d margin)", rate, fp*margin, fp, margin)
	}
}

func TestUnion(t *testing.T) {
	a := NewWithEstimates(100, 0.01)
	b := NewWithEstimates(100, 0.01)

	a.Add([]byte("foo"))
	b.Add([]byte("bar"))

	if err := a.Union(b); err != nil {
		t.Fatalf("Union returned unexpected error: %v", err)
	}

	if !a.Test([]byte("foo")) {
		t.Error("expected union to still contain items from the receiver")
	}
	if !a.Test([]byte("bar")) {
		t.Error("expected union to contain items from the merged filter")
	}
}

func TestUnionRejectsIncompatibleFilters(t *testing.T) {
	a := New(1000, 3)
	b := New(2000, 3)

	if err := a.Union(b); err != ErrIncompatibleFilters {
		t.Errorf("expected ErrIncompatibleFilters, got %v", err)
	}
}

func TestEstimateParametersScalesWithN(t *testing.T) {
	mSmall, _ := EstimateParameters(10, 0.01)
	mLarge, _ := EstimateParameters(10000, 0.01)

	if mLarge <= mSmall {
		t.Errorf("expected larger n to require a larger bit array: m(10)=%d, m(10000)=%d", mSmall, mLarge)
	}
}

func TestEmptyFilterNeverMatches(t *testing.T) {
	f := NewWithEstimates(100, 0.01)
	if f.Test([]byte("anything")) {
		t.Error("expected empty filter to report nothing as present")
	}
}

func TestClear(t *testing.T) {
	f := NewWithEstimates(100, 0.01)
	f.Add([]byte("present"))
	if !f.Test([]byte("present")) {
		t.Fatal("expected added item to be present before clear")
	}

	f.Clear()
	if f.Count() != 0 {
		t.Errorf("expected count to be reset, got %d", f.Count())
	}
	if f.Test([]byte("present")) {
		t.Error("expected cleared filter to report nothing as present")
	}
}

func TestCopy(t *testing.T) {
	original := NewWithEstimates(100, 0.01)
	original.Add([]byte("original"))
	copyFilter := original.Copy()

	if copyFilter.M() != original.M() || copyFilter.K() != original.K() || copyFilter.Count() != original.Count() {
		t.Fatal("expected copy to preserve filter metadata")
	}
	copyFilter.Add([]byte("copy-only"))
	if original.Test([]byte("copy-only")) {
		t.Error("expected mutations to the copy not to affect the original")
	}
	original.Add([]byte("original-only"))
	if copyFilter.Test([]byte("original-only")) {
		t.Error("expected mutations to the original not to affect the copy")
	}
}

func TestString(t *testing.T) {
	f := New(230, 7)
	f.Add([]byte("one"))
	f.Add([]byte("two"))

	if got, want := f.String(), "Filter{m=230, k=7, count=2}"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestBinaryRoundTrip(t *testing.T) {
	original := NewWithEstimates(100, 0.01)
	for _, item := range []string{"apple", "banana", "cherry"} {
		original.Add([]byte(item))
	}

	data, err := original.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary returned unexpected error: %v", err)
	}
	restored := New(1, 1)
	if err := restored.UnmarshalBinary(data); err != nil {
		t.Fatalf("UnmarshalBinary returned unexpected error: %v", err)
	}
	if restored.M() != original.M() || restored.K() != original.K() || restored.Count() != original.Count() {
		t.Fatal("expected restored filter metadata to match original")
	}
	for _, item := range []string{"apple", "banana", "cherry"} {
		if !restored.Test([]byte(item)) {
			t.Errorf("expected restored filter to contain %q", item)
		}
	}
}

func BenchmarkAdd(b *testing.B) {
	for _, size := range []int{100, 100000} {
		b.Run(fmt.Sprintf("%d-items", size), func(b *testing.B) {
			items := make([][]byte, size)
			for i := range items {
				items[i] = []byte(fmt.Sprintf("item-%d", i))
			}

			f := NewWithEstimates(uint(size), 0.01)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f.Add(items[i%size])
			}
		})
	}
}

func BenchmarkTest(b *testing.B) {
	for _, size := range []int{100, 100000} {
		b.Run(fmt.Sprintf("%d-items", size), func(b *testing.B) {
			items := make([][]byte, size)
			f := NewWithEstimates(uint(size), 0.01)
			for i := range items {
				items[i] = []byte(fmt.Sprintf("item-%d", i))
				f.Add(items[i])
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f.Test(items[i%size])
			}
		})
	}
}
