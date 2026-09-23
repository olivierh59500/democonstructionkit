package sprites

import (
	"image/color"
	"math"
	"testing"
)

func TestStreakFieldMatchesLegacyMotionWithDriftAndWrap(t *testing.T) {
	newRandom := func() func() float64 {
		seed := uint32(42)
		return func() float64 {
			seed = seed*1664525 + 1013904223
			return float64(seed) / 4294967296
		}
	}
	base := StreakConfig{
		Width: 320, Height: 200, Count: 80, Speed: 4, Focal: 100,
		CenterX: 200, CenterY: 120, Color: color.RGBA{170, 170, 170, 255},
	}
	base.Random = newRandom()
	legacy, err := NewStreaks(base)
	if err != nil {
		t.Fatal(err)
	}
	base.Random = newRandom()
	config, err := StreakField(base)
	if err != nil {
		t.Fatal(err)
	}
	field, err := NewField(config.Field)
	if err != nil {
		t.Fatal(err)
	}
	field.Sample(config.View)
	clear(field.history)
	var sampleByIndex [80]FieldSample
	var present [80]bool
	for tick := 1; tick <= 1200; tick++ {
		legacy.Step()
		field.Step(config.Delta, config.Velocity)
		clear(present[:])
		for _, sample := range field.Sample(config.View) {
			sampleByIndex[sample.Index] = sample
			present[sample.Index] = true
		}
		for i, old := range legacy.stars {
			visible := false
			var appearance FieldAppearance
			if present[i] {
				visible = config.Style.Sample(sampleByIndex[i], &appearance)
			}
			if visible != old.visible {
				t.Fatalf("tick %d star %d visibility = %t, want %t", tick, i, visible, old.visible)
			}
			if !visible {
				continue
			}
			sample := sampleByIndex[i]
			if math.Abs(sample.X-old.px) > 1e-10 || math.Abs(sample.Y-old.py) > 1e-10 ||
				math.Abs(sample.PreviousX-old.oldX) > 1e-10 || math.Abs(sample.PreviousY-old.oldY) > 1e-10 ||
				appearance.TrailWidth != old.width {
				t.Fatalf("tick %d star %d projected motion differs: sample=%+v, width=%g, old=%+v, point=%+v", tick, i, sample, appearance.TrailWidth, old, field.points[i])
			}
		}
	}
}
