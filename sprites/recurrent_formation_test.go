package sprites

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestRecurrentFormationMatchesIndexedHarmonicsWithoutAllocations(t *testing.T) {
	image := ebiten.NewImage(8, 8)
	defer image.Deallocate()
	config := RecurrentFormationConfig{
		Image: image, Count: 10, Origin: motion.Point{X: 176, Y: 143},
		XAmplitude: 176, YAmplitude: 51.5, YSecondaryAmplitude: 51.5,
		XDivisor: 25, XSecondaryDivisor: 300, YDivisor: 37, YSecondaryDivisor: 17,
		XIndexStep: .2, XSecondaryStep: 1.0 / 60.0,
		YIndexStep: 5.0 / 37.0, YSecondaryStep: 5.0 / 17.0,
		YSecondaryCos: true,
		ScaleX:        2, ScaleY: 2, OutputScaleX: 2, OutputScaleY: 2,
	}
	formation, err := NewRecurrentFormation(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, tick := range []float64{0, 1, 60, 240, 600, 1200, 4800} {
		if err := formation.Update(tick); err != nil {
			t.Fatal(err)
		}
		for index, pose := range formation.Poses() {
			i := float64(index)
			wantX := 176 + 176*math.Sin(tick/25+i*.2)*math.Cos(tick/300+i/60)
			wantY := 143 + 51.5*math.Sin(tick/37+i*5/37) + 51.5*math.Cos(tick/17+i*5/17)
			if math.Abs(pose.X-wantX) > 1e-9 || math.Abs(pose.Y-wantY) > 1e-9 {
				t.Fatalf("tick %.0f logo %d = %+v, want %v,%v", tick, index, pose, wantX, wantY)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = formation.Update(123) }); allocations != 0 {
		t.Fatalf("formation update allocates %.1f objects", allocations)
	}
	config.Count = 3
	if err := formation.SetConfig(config); err != nil || len(formation.Poses()) != 3 {
		t.Fatal("formation count change did not reuse the effect")
	}
	config.YSecondaryCos = false
	if err := formation.SetConfig(config); err != nil {
		t.Fatal(err)
	}
	if err := formation.Update(0); err != nil || formation.Poses()[0].Y != 143 {
		t.Fatal("sine alternative retained the cosine envelope")
	}
}
