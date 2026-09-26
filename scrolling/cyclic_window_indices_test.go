package scrolling

import "testing"

type syntheticCycleOffsets struct {
	offsets []float64
	period  float64
}

func (source syntheticCycleOffsets) offsetAt(index int) float64 {
	n := len(source.offsets)
	return source.offsets[index%n] + float64(index/n)*source.period
}

func TestCyclicWindowIndicesKeepProportionalSeamsVisible(t *testing.T) {
	source := syntheticCycleOffsets{offsets: []float64{0, 7, 19, 26}, period: 33}
	const count, scale, minimum, maximum = 8, 1.5, -4, 24
	for frame := 0; frame < 5000; frame++ {
		origin := -float64(frame%67) * .75
		first, end := cyclicWindowIndices(source, count, origin, scale, minimum, maximum)
		if end-first > 4 {
			t.Fatalf("frame %d visited %d glyphs for a four-glyph window", frame, end-first)
		}
		for index := 0; index < count; index++ {
			x := origin + source.offsetAt(index)*scale
			inside := x > minimum && x < maximum
			selected := index >= first && index < end
			if inside != selected {
				t.Fatalf("frame %d glyph %d x=%g selected=%t, want=%t", frame, index, x, selected, inside)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		_, _ = cyclicWindowIndices(source, count, -22.5, scale, minimum, maximum)
	}); allocations != 0 {
		t.Fatalf("cyclic search allocated %v times", allocations)
	}
}

func TestCyclicWindowEnvelopeKeepsMixedFontBearings(t *testing.T) {
	source := syntheticCycleOffsets{offsets: []float64{0, 8, 23, 37}, period: 48}
	bearings := [...]float64{-3, 5, -1, 4}
	const count, scale, minimum, maximum = 8, 2, -12, 96
	for frame := 0; frame < 5000; frame++ {
		origin := -float64(frame%113) * .75
		first, end := cyclicWindowIndices(source, count, origin, scale,
			minimum-5*scale, maximum-(-3)*scale)
		for index := 0; index < count; index++ {
			x := origin + (source.offsetAt(index)+bearings[index%len(bearings)])*scale
			if x > minimum && x < maximum && (index < first || index >= end) {
				t.Fatalf("frame %d glyph %d at %g was excluded by bearing envelope", frame, index, x)
			}
		}
	}
}
