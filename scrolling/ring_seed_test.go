package scrolling

import "testing"

func TestSeamlessRingSeedFillsShortViewportWithoutChangingLongText(t *testing.T) {
	const slots, short, long = 10, 4, 30
	for i := 0; i < slots; i++ {
		if got := ringSeedIndex(i, short, true); got != i%short {
			t.Fatalf("short seed slot %d = %d, want %d", i, got, i%short)
		}
		if got := ringSeedIndex(i, long, true); got != i {
			t.Fatalf("long seed slot %d = %d, want %d", i, got, i)
		}
		if got := ringSeedIndex(i, short, false); got != i {
			t.Fatalf("legacy seed slot %d changed to %d", i, got)
		}
	}
	if got := ringSeedCursor(slots, short, true); got != 2 {
		t.Fatalf("short cursor = %d, want 2", got)
	}
	if got := ringSeedCursor(slots, long, true); got != slots {
		t.Fatalf("long cursor = %d, want %d", got, slots)
	}
	if got := ringSeedCursor(slots, short, false); got != slots {
		t.Fatalf("legacy cursor = %d, want %d", got, slots)
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_ = ringSeedIndex(9, 4, true)
		_ = ringSeedCursor(10, 4, true)
	}); allocations != 0 {
		t.Fatalf("ring seed allocated %v times", allocations)
	}
}
