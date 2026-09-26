package motion

import (
	"math"
	"testing"
)

func TestScaledTextClockPreservesDOMSizeCuesAndWrap(t *testing.T) {
	fonts := make([]int, 120)
	for i := range fonts {
		switch {
		case i >= 90:
			fonts[i] = 3
		case i >= 55:
			fonts[i] = 2
		case i >= 20:
			fonts[i] = 1
		}
	}
	scales := []float64{1, 2, 4, 8}
	baseSpeeds := []float64{8, 4, 2, 1}
	clock, err := NewScaledTextClock(ScaledTextClockConfig{
		FontAt: fonts, Scales: scales, BaseSpeeds: baseSpeeds,
		ViewportWidth: 640, TileWidth: 40, StartOffset: 640,
		SpeedMultiplier: 1, Lookahead: 1, WrapInclusive: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	oldOffset, oldBank, multiplier, wraps, switches := 640.0, 0, 1.0, 0, 0
	for tick := 1; tick <= 20000; tick++ {
		switch tick {
		case 300:
			multiplier = .5
		case 900:
			multiplier = 2
		case 1500:
			multiplier = .25
		}
		if err := clock.SetSpeedMultiplier(multiplier); err != nil {
			t.Fatal(err)
		}
		oldOffset -= baseSpeeds[oldBank] * multiplier
		if oldOffset <= -float64(len(fonts)*40) {
			oldOffset += float64(len(fonts)*40 + 640)
			wraps++
		}
		var offsets [4]float64
		for i, scale := range scales {
			offsets[i] = oldOffset*scale + (1-scale)*640
		}
		width := 40 * scales[oldBank]
		left := int(math.Floor(-offsets[oldBank] / width))
		if left < 0 {
			left = 0
		}
		lookahead := min(left+int(math.Ceil(640/width))+1, len(fonts)-1)
		if fonts[lookahead] != oldBank {
			switches++
		}
		oldBank = fonts[lookahead]
		clock.Step()
		if clock.ActiveBank() != oldBank || clock.LookaheadIndex() != lookahead {
			t.Fatalf("tick %d: bank/lookahead %d/%d, want %d/%d", tick, clock.ActiveBank(), clock.LookaheadIndex(), oldBank, lookahead)
		}
		for i, want := range offsets {
			if got := clock.Offset(i); math.Abs(got-want) > 1e-9 {
				t.Fatalf("tick %d bank %d: offset %.12f, want %.12f", tick, i, got, want)
			}
		}
	}
	if wraps < 2 || switches < 3 {
		t.Fatalf("test did not exercise enough wraps (%d) or size switches (%d)", wraps, switches)
	}
	if allocations := testing.AllocsPerRun(100, clock.Step); allocations != 0 {
		t.Fatalf("clock step allocated %.2f objects", allocations)
	}
}

func TestScaledTextClockRejectsUnknownCues(t *testing.T) {
	_, err := NewScaledTextClock(ScaledTextClockConfig{FontAt: []int{1}, Scales: []float64{1},
		BaseSpeeds: []float64{1}, ViewportWidth: 100, TileWidth: 8, StartOffset: 100, SpeedMultiplier: 1})
	if err == nil {
		t.Fatal("accepted a cue to an absent bank")
	}
}

func BenchmarkScaledTextClockFourBanks(b *testing.B) {
	fonts := make([]int, 160)
	for i := range fonts {
		fonts[i] = i / 40
	}
	clock, err := NewScaledTextClock(ScaledTextClockConfig{FontAt: fonts,
		Scales: []float64{1, 2, 4, 8}, BaseSpeeds: []float64{8, 4, 2, 1},
		ViewportWidth: 640, TileWidth: 40, StartOffset: 640,
		SpeedMultiplier: 1, Lookahead: 1, WrapInclusive: true})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		clock.Step()
	}
}
