package composite

import (
	"image"
	"math"
	"testing"
)

func TestBandsRetainIndependentSignedPhaseAndCopyConfig(t *testing.T) {
	c := BandsConfig{Bands: []MovingBand{{Source: image.Rect(0, 0, 1024, 10), VelocityX: -8, WrapX: 256, MotionScaleX: 2}, {Source: image.Rect(0, 10, 1024, 20), VelocityX: -.5, WrapX: 256}}, CopyOffsets: [][2]float64{{0, 0}, {640, 0}}}
	b, err := NewBands(c)
	if err != nil {
		t.Fatal(err)
	}
	c.Bands[0].VelocityX = 99
	c.CopyOffsets[1][0] = 0
	for i := 0; i < 33; i++ {
		b.Step()
	}
	if b.config.Bands[0].PhaseX != -8 || b.config.Bands[1].PhaseX != -16.5 || b.config.CopyOffsets[1][0] != 640 {
		t.Fatalf("unexpected shared motion: %+v", b.config)
	}
	if allocs := testing.AllocsPerRun(100, func() { b.Step() }); allocs != 0 {
		t.Fatalf("Step allocated %g times", allocs)
	}
}
func TestBandsRejectUnboundedOrInvalidMotion(t *testing.T) {
	for _, c := range []BandsConfig{{}, {Bands: []MovingBand{{Source: image.Rect(0, 0, 2, 2), VelocityX: math.NaN()}}}, {Bands: []MovingBand{{Source: image.Rect(0, 0, 2, 2), WrapX: -1}}}, {Bands: make([]MovingBand, 10000), CopyOffsets: make([][2]float64, 10000)}} {
		if _, err := NewBands(c); err == nil {
			t.Fatal("invalid bands accepted")
		}
	}
}

func TestBandControlsPreserveIndependentPhases(t *testing.T) {
	bands, err := NewBands(BandsConfig{Bands: []MovingBand{{Source: image.Rect(0, 0, 16, 4), VelocityX: -1, WrapX: 16}, {Source: image.Rect(0, 4, 16, 8), VelocityX: -2, WrapX: 16}}})
	if err != nil {
		t.Fatal(err)
	}
	bands.Step()
	if err := bands.SetVelocity(0, -4, 0); err != nil {
		t.Fatal(err)
	}
	first, _ := bands.Band(0)
	if first.PhaseX != -1 {
		t.Fatal("changing velocity reset phase")
	}
	if err := bands.SetPhase(1, -15, 0); err != nil {
		t.Fatal(err)
	}
	bands.Step()
	first, _ = bands.Band(0)
	second, _ := bands.Band(1)
	if first.PhaseX != -5 || second.PhaseX != -1 {
		t.Fatalf("independent control lost: %g,%g", first.PhaseX, second.PhaseX)
	}
	if bands.SetVelocity(2, 0, 0) == nil || bands.SetPhase(0, math.NaN(), 0) == nil {
		t.Fatal("invalid control accepted")
	}
}
