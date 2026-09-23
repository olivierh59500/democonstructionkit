package scrolling

import (
	"math"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestBitmapSlotsRetainDrawBeforeRecycleAndPreviousGlyphTangent(t *testing.T) {
	atlas := ebiten.NewImage(440, 264)
	defer atlas.Deallocate()
	grid, _ := (BitmapSpec{Width: 44, Height: 44, First: 32, Columns: 10}).Grid(atlas, ebiten.FilterLinear)
	text := strings.Repeat("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 10)
	cfg := BitmapSlotsConfig{Font: grid, Text: text, Count: 24, Start: 1056, Advance: 44, Speed: 3, RecycleBelow: -88, Period: 1056, Y: 250, PreviousTangent: true, Waves: []RingWave{{Amplitude: 50, LetterStep: -.5, TickStep: .05}}}
	shared, err := NewBitmapSlots(cfg)
	if err != nil {
		t.Fatal(err)
	}
	var xs [24]float64
	var letters [24]int
	for i := range xs {
		xs[i] = float64((24 + i) * 44)
		letters[i] = i
	}
	oldX, oldY, u, next := 0.0, 0.0, 0.0, 24
	for tick := 0; tick < 6000; tick++ {
		shared.Step()
		for i := range xs {
			y := 250 + math.Sin(u-float64(i)/2)*50
			xs[i] -= 3
			angle := math.Atan2(y-oldY, xs[i]-oldX)
			// Parameterized multiplication can differ by one rounding bit from
			// a compiler-fused expression using literal wave constants.
			got := shared.draws[i]
			if got.glyph != letters[i] || got.pose.X != xs[i] || math.Abs(got.pose.Y-y) > 1e-11 || math.Abs(got.pose.Angle-angle) > 1e-11 {
				t.Fatalf("tick %d slot %d diverged: %+v", tick, i, got)
			}
			oldX, oldY = xs[i], y
			if xs[i] < -88 {
				xs[i] += 1056
				letters[i] = next
				next = (next + 1) % len(text)
			}
		}
		u += .05
	}
	if allocations := testing.AllocsPerRun(1000, func() { shared.Step() }); allocations != 0 {
		t.Fatalf("slot update allocated %g times", allocations)
	}
	positions := shared.slots[0].x
	if err := shared.SetSpeed(0); err != nil {
		t.Fatal(err)
	}
	shared.Step()
	if shared.slots[0].x != positions {
		t.Fatal("pausing changed transport position")
	}
	cfg.Text = "A"
	cfg.Count = 100
	if _, err := NewBitmapSlots(cfg); err != nil {
		t.Fatal("short text should fill recycled slots", err)
	}
}
func TestBitmapSlotsRejectInvalidTransport(t *testing.T) {
	for _, cfg := range []BitmapSlotsConfig{{}, {Count: 1, Text: "A", Advance: 1, Period: 1, Speed: 2}, {Count: 1, Text: "A", Advance: 1, Speed: math.NaN()}} {
		if _, err := NewBitmapSlots(cfg); err == nil {
			t.Fatal("invalid slots accepted")
		}
	}
}
