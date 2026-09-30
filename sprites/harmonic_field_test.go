package sprites

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestHarmonicFieldKeepsAbsoluteSampledPoses(t *testing.T) {
	f, err := NewHarmonicField(HarmonicFieldConfig{
		Count: 3, ClockScale: [2]float64{85}, PixelSnap: true,
		Motion: motion.HarmonicFormationConfig{Origin: motion.Point{X: 100, Y: 80},
			X: []motion.IndexedHarmonic{{Amplitude: 25, Rate: .0242, IndexRate: .1254, PhasePeriod: 2 * math.Pi}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	for _, seconds := range []float64{0, 120, .25, 120} {
		if err := f.Update(kit.Frame{Time: seconds}); err != nil {
			t.Fatal(err)
		}
		for i, p := range f.Samples() {
			want := math.Round(100 + 25*math.Sin(math.Mod(seconds*85*.0242+float64(i)*.1254, 2*math.Pi)))
			if p.X != want || p.Y != 80 || p.Index != i || p.Scale != 1 {
				t.Fatalf("sample %+v at %v, want X %v", p, seconds, want)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { f.Sample([2]float64{200}, 1) }); allocations != 0 {
		t.Fatalf("steady sampling allocated %v times", allocations)
	}
	f.Close()
	if err := f.Update(kit.Frame{}); err == nil {
		t.Fatal("updated a closed formation")
	}
}

func TestHarmonicFieldRejectsInvalidPopulation(t *testing.T) {
	for _, c := range []HarmonicFieldConfig{
		{Count: -1}, {Count: 65537}, {ClockScale: [2]float64{math.NaN()}},
		{Count: 2, Motion: motion.HarmonicFormationConfig{IndexOffsets: []float64{1}}},
	} {
		if f, err := NewHarmonicField(c); err == nil {
			f.Close()
			t.Fatalf("accepted %+v", c)
		}
	}
}
