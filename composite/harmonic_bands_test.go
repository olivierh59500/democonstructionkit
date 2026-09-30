package composite

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestHarmonicBandsOwnsPosesWithRoundThenClamp(t *testing.T) {
	white := ebiten.NewImage(1, 1)
	defer white.Deallocate()
	colors := []color.NRGBA{{R: 255, A: 255}, {B: 255, A: 255}}
	bounds := &motion.FormationBounds{Min: motion.Point{}, Max: motion.Point{X: 1.6, Y: 10}}
	b, err := NewHarmonicBands(HarmonicBandsConfig{
		LeftX: 2, RightX: 40, Thickness: 3, Colors: colors, White: white, PixelSnap: true,
		Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: 1.6, Y: 4.4}, Spacing: motion.Point{Y: 2}}, Bounds: bounds,
	})
	if err != nil {
		t.Fatal(err)
	}
	colors[0], bounds.Max = color.NRGBA{}, motion.Point{}
	if b.config.Colors[0].R != 255 || b.Poses()[0] != (motion.Point{X: 1.6, Y: 4}) || b.Poses()[1].Y != 6 {
		t.Fatalf("round/clamp or config ownership failed: %v", b.Poses())
	}
	if allocations := testing.AllocsPerRun(100, func() { b.Sample([2]float64{1, 2}, 1) }); allocations != 0 {
		t.Fatalf("steady sampling allocated %v times", allocations)
	}
	if err := b.Update(kit.Frame{Time: math.NaN()}); err == nil {
		t.Fatal("accepted nonfinite time")
	}
	if err := b.Close(); err != nil {
		t.Fatal(err)
	}
	b.Close()
	if err := b.Sample([2]float64{}, 1); err == nil {
		t.Fatal("sampled a closed band bank")
	}
	if white.Bounds().Dx() != 1 {
		t.Fatal("closed a borrowed material")
	}
}

func TestHarmonicBandsRejectsInvalidConfiguration(t *testing.T) {
	valid := HarmonicBandsConfig{RightX: 10, Thickness: 1, Colors: []color.NRGBA{{A: 255}}}
	for _, mutate := range []func(*HarmonicBandsConfig){
		func(c *HarmonicBandsConfig) { c.Colors = nil },
		func(c *HarmonicBandsConfig) { c.Thickness = 0 },
		func(c *HarmonicBandsConfig) { c.LeftX = 10 },
		func(c *HarmonicBandsConfig) { c.ClockScale[0] = math.Inf(1) },
		func(c *HarmonicBandsConfig) { c.Motion.X = []motion.IndexedHarmonic{{PhasePeriod: -1}} },
	} {
		c := valid
		mutate(&c)
		if b, err := NewHarmonicBands(c); err == nil {
			b.Close()
			t.Fatalf("accepted %+v", c)
		}
	}
}
