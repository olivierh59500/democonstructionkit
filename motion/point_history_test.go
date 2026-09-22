package motion

import "testing"

func TestPointHistoryStartupAndWraparound(t *testing.T) {
	h, err := NewPointHistory(4, -1)
	if err != nil {
		t.Fatal(err)
	}
	if h.Capacity() != 4 {
		t.Fatal("wrong history capacity")
	}
	for delay := 0; delay < h.Capacity(); delay++ {
		if v, ok := h.At(delay); !ok || v != -1 {
			t.Fatalf("startup delay %d was not filled: %d %v", delay, v, ok)
		}
	}
	for current := 0; current < 30; current++ {
		h.Push(current)
		for delay := 0; delay < h.Capacity(); delay++ {
			want := max(-1, current-delay)
			if v, ok := h.At(delay); !ok || v != want {
				t.Fatalf("push %d delay %d = %d %v, want %d", current, delay, v, ok, want)
			}
		}
	}
	for _, delay := range []int{-1, 4, 1000} {
		if _, ok := h.At(delay); ok {
			t.Fatalf("accepted out-of-range delay %d", delay)
		}
	}
	h.Reset(42)
	for delay := 0; delay < h.Capacity(); delay++ {
		if v, ok := h.At(delay); !ok || v != 42 {
			t.Fatal("reset did not refill all delays")
		}
	}
}

func TestPointHistorySingleEntryAndIndependentInstances(t *testing.T) {
	a, _ := NewPointHistory(1, [2]float64{4, 8})
	b, _ := NewPointHistory(1, [2]float64{4, 8})
	for i := range 8 {
		a.Push([2]float64{float64(i), -float64(i)})
		if v, ok := a.At(0); !ok || v != [2]float64{float64(i), -float64(i)} {
			t.Fatal("single-entry history did not advance")
		}
	}
	if v, ok := b.At(0); !ok || v != [2]float64{4, 8} {
		t.Fatal("histories share storage")
	}
	if allocations := testing.AllocsPerRun(1000, func() { a.Push([2]float64{1, 2}); a.At(0) }); allocations != 0 {
		t.Fatalf("history allocates per tick: %g", allocations)
	}
}

func TestPointHistoryInvalidAndZeroValue(t *testing.T) {
	for _, capacity := range []int{0, -1} {
		if _, err := NewPointHistory(capacity, 0); err == nil {
			t.Fatalf("accepted capacity %d", capacity)
		}
	}
	var h PointHistory[int]
	h.Reset(1)
	h.Push(2)
	if _, ok := h.At(0); ok || h.Capacity() != 0 {
		t.Fatal("zero history reports a value")
	}
}
