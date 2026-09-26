package scrolling

import (
	"testing"
)

func TestCyclicWindowSelectsProportionalGlyphsAndClipsBearings(t *testing.T) {
	scroll, err := New(Config{Glyphs: []Glyph{
		{Advance: 10}, {Advance: 20, X: -3}, {Advance: 15, X: 5},
	}})
	if err != nil {
		t.Fatal(err)
	}
	window, err := NewCyclicWindow(scroll, CyclicWindowConfig{
		Scale: 2, Minimum: -4, Maximum: 40, Copies: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, origin := range []float64{0, -10, -30, -60, -89.5} {
		state := window.At(origin)
		if !state.Cycle || state.Map == nil || state.End-state.First >= 2*scroll.GlyphCount() {
			t.Fatalf("origin %g did not produce a bounded cyclic state: %+v", origin, state)
		}
		for index := 0; index < 2*scroll.GlyphCount(); index++ {
			glyph := scroll.glyphs[index%scroll.GlyphCount()]
			x := origin + (scroll.offsetAt(index)+glyph.X)*2
			inside := x > -4 && x < 40
			selected := index >= state.First && index < state.End
			if inside && !selected {
				t.Fatalf("origin %g glyph %d at %g was excluded", origin, index, x)
			}
			if selected && state.Map(Sample{X: x}, nil) != inside {
				t.Fatalf("origin %g glyph %d at %g ignored its clip", origin, index, x)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = window.At(-30) }); allocations != 0 {
		t.Fatalf("cyclic window allocated %v times per frame", allocations)
	}
}
