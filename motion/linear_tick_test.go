package motion

import "testing"

func TestLinearTickMatchesMenuUncoverWidth(t *testing.T) {
	line, err := NewLinearTick(LinearTickConfig{Start: 840, Step: -5, Min: 0, Max: 840})
	if err != nil {
		t.Fatal(err)
	}
	width := 840
	for tick := 0; tick < 500; tick++ {
		if got := line.At(tick); got != width {
			t.Fatalf("tick %d cover = %d, want %d", tick, got, width)
		}
		width = max(0, width-5)
	}
}
