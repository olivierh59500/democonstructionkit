package motion

import (
	"math"
	"testing"
)

func TestFrameFieldPreservesFractionalFrameAndOrderedRespawn(t *testing.T) {
	var initial, resets [2]int
	initialCount, resetCount := 0, 0
	field, err := NewFrameField(FrameFieldConfig{Count: 2, EndPhase: 9,
		Spawn: func(index int, reset bool) FrameParticle {
			if reset {
				if resetCount < len(resets) {
					resets[resetCount] = index
				}
				resetCount++
				return FrameParticle{X: float64(index * 64), Y: 100, Phase: 0, Rate: .25}
			}
			initial[initialCount] = index
			initialCount++
			return FrameParticle{X: float64(index * 64), Y: 50, Phase: 8.75, Rate: .25}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if initialCount != 2 || initial[0] != 0 || initial[1] != 1 || resetCount != 0 {
		t.Fatalf("initial spawn order %v, resets %v", initial, resets)
	}
	if err := field.Step(); err != nil {
		t.Fatal(err)
	}
	if resetCount != 2 || resets[0] != 0 || resets[1] != 1 {
		t.Fatalf("respawn order %v", resets)
	}
	if field.Samples()[1] != (FrameParticle{X: 64, Y: 100, Phase: 0, Rate: .25}) {
		t.Fatalf("respawn changed frame or position: %+v", field.Samples()[1])
	}
	if err := field.SetSpeedMultiplier(.5); err != nil {
		t.Fatal(err)
	}
	if err := field.Step(); err != nil {
		t.Fatal(err)
	}
	if got := field.Samples()[0].Phase; got != .125 {
		t.Fatalf("half-speed frame phase %v", got)
	}
	if got := testing.AllocsPerRun(100, func() { _ = field.Step() }); got != 0 {
		t.Fatalf("frame field update allocated %.2f objects", got)
	}
}

func TestFrameFieldRejectsInvalidState(t *testing.T) {
	if _, err := NewFrameField(FrameFieldConfig{Count: 1, EndPhase: 9}); err == nil {
		t.Fatal("accepted a field without a spawn recipe")
	}
	if _, err := NewFrameField(FrameFieldConfig{Count: 1, EndPhase: 9,
		Spawn: func(int, bool) FrameParticle { return FrameParticle{Rate: 0} },
	}); err == nil {
		t.Fatal("accepted a particle that cannot advance")
	}
}

func TestFrameFieldMovesPlanarLayersAndPreservesSingleWrapOvershoot(t *testing.T) {
	wraps := 0
	field, err := NewFrameField(FrameFieldConfig{Count: 2, EndPhase: 0,
		Spawn: func(index int, _ bool) FrameParticle {
			return FrameParticle{X: []float64{639, 0}[index], Y: 5,
				VelocityX: []float64{11.2, 5.6}[index], Image: index}
		},
		WrapX: &FrameAxisWrap{Boundary: 640, Shift: 640,
			OnWrap: func(index int, p *FrameParticle) { wraps++; p.Y = float64(50 + index) }},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := field.SetSpeedMultiplier(1.4); err != nil {
		t.Fatal(err)
	}
	if err := field.Step(); err != nil {
		t.Fatal(err)
	}
	first, second := field.Samples()[0], field.Samples()[1]
	if wraps != 1 || math.Abs(first.X-(639+11.2*1.4-640)) > 1e-12 || first.Y != 50 || first.Image != 0 ||
		math.Abs(second.X-5.6*1.4) > 1e-12 || second.Y != 5 || second.Image != 1 {
		t.Fatalf("planar wrap changed positions: wraps=%d first=%+v second=%+v", wraps, first, second)
	}
	if got := testing.AllocsPerRun(100, func() { _ = field.Step() }); got != 0 {
		t.Fatalf("planar frame field update allocates %.2f objects", got)
	}
}

func BenchmarkFrameField500(b *testing.B) {
	field, err := NewFrameField(FrameFieldConfig{Count: 500, EndPhase: 9,
		Spawn: func(index int, _ bool) FrameParticle {
			return FrameParticle{X: float64(index), Phase: 0, Rate: .125}
		},
	})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := field.Step(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkFrameFieldPlanar105(b *testing.B) {
	field, err := NewFrameField(FrameFieldConfig{Count: 105,
		Spawn: func(index int, _ bool) FrameParticle {
			return FrameParticle{X: float64(index * 6), VelocityX: []float64{11.2, 5.6, 2.8}[index/35]}
		},
		WrapX: &FrameAxisWrap{Boundary: 640, Shift: 640,
			OnWrap: func(index int, p *FrameParticle) { p.Y = float64(index % 280) }},
	})
	if err != nil {
		b.Fatal(err)
	}
	if err := field.SetSpeedMultiplier(1.4); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		if err := field.Step(); err != nil {
			b.Fatal(err)
		}
	}
}
