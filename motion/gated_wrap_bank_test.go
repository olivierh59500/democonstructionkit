package motion

import (
	"math"
	"testing"
)

func TestGatedWrapBankPreservesThreeIndependentAtlasClocks(t *testing.T) {
	config := GatedWrapBankConfig{
		Start: []float64{0, 0, 0}, Velocity: []float64{.35}, Gates: []float64{120, 500, 760},
		Upper: &WrapLimit{Boundary: 82, Restart: 0},
	}
	gated, err := NewGatedWrapBank(config)
	if err != nil {
		t.Fatal(err)
	}
	legacy := [3]float64{}
	for tick := 0; tick < 3000; tick++ {
		time := float64(tick) * .5
		for i, begin := range []float64{120, 500, 760} {
			if time > begin {
				legacy[i] += .35
				if legacy[i] > 82 {
					legacy[i] = 0
				}
			}
		}
		if err := gated.StepAt(time); err != nil {
			t.Fatal(err)
		}
		for i, value := range legacy {
			if gated.At(i) != value || gated.Frame(i) != int(math.Floor(value)) {
				t.Fatalf("tick %d lane %d = %v, want %v", tick, i, gated.At(i), value)
			}
		}
	}
	gated.Reset()
	for i := 0; i < 3; i++ {
		if gated.At(i) != 0 {
			t.Fatal("reset retained an atlas phase")
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = gated.StepAt(1200) }); allocations != 0 {
		t.Fatalf("gated wrap step allocates %v times", allocations)
	}
}

func TestGatedWrapBankRejectsInvalidTimes(t *testing.T) {
	if _, err := NewGatedWrapBank(GatedWrapBankConfig{Start: []float64{0}, Velocity: []float64{1}, Gates: []float64{math.NaN()}, Upper: &WrapLimit{Boundary: 10}}); err == nil {
		t.Fatal("accepted nonfinite gate")
	}
}
