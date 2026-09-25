package motion

import "testing"

func TestWrapBankMatchesThreeLayerParallaxAndLiveSpeedControls(t *testing.T) {
	config := WrapBankConfig{
		Start: []float64{-640, -640, -640}, Velocity: []float64{-2, -4, -6},
		Lower: &WrapLimit{Boundary: -640, Restart: 0},
		Upper: &WrapLimit{Boundary: 0, Restart: -640},
	}
	bank, err := NewWrapBank(config)
	if err != nil {
		t.Fatal(err)
	}
	config.Start[0] = 0
	config.Velocity[0] = 99
	config.Lower.Restart = 999
	config.Upper.Restart = 999
	legacyPositions := []float64{-640, -640, -640}
	legacySpeeds := []float64{-2, -4, -6}
	for tick := 0; tick < 400; tick++ {
		if tick == 80 {
			legacySpeeds[0] = 2
			if err := bank.SetVelocity(0, 2); err != nil {
				t.Fatal(err)
			}
		}
		if tick == 160 {
			legacySpeeds[1]--
			if err := bank.AddVelocity(1, -1); err != nil {
				t.Fatal(err)
			}
		}
		if tick == 240 {
			for index := range legacySpeeds {
				legacySpeeds[index] = -legacySpeeds[index]
			}
			bank.ReverseAll()
		}
		bank.Step()
		for index := range legacyPositions {
			legacyPositions[index] += legacySpeeds[index]
			if legacyPositions[index] > 0 {
				legacyPositions[index] = -640
			}
			if legacyPositions[index] < -640 {
				legacyPositions[index] = 0
			}
			if bank.At(index) != legacyPositions[index] || bank.Velocity(index) != legacySpeeds[index] {
				t.Fatalf("tick %d layer %d = (%v, %v), want (%v, %v)", tick, index, bank.At(index), bank.Velocity(index), legacyPositions[index], legacySpeeds[index])
			}
		}
	}
	bank.Reset()
	if bank.At(0) != -640 || bank.Velocity(0) != -2 {
		t.Fatalf("reset did not restore original state")
	}
	if allocs := testing.AllocsPerRun(100, bank.Step); allocs != 0 {
		t.Fatalf("wrap step allocates %v times", allocs)
	}
}

func TestWrapBankPreservesInclusiveSingleBoundaryLoops(t *testing.T) {
	water, err := NewWrapBank(WrapBankConfig{
		Start: []float64{0}, Velocity: []float64{2},
		Upper: &WrapLimit{Boundary: 220, Restart: 0, Inclusive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 109; tick++ {
		water.Step()
	}
	if water.At(0) != 218 {
		t.Fatal(water.At(0))
	}
	water.Step()
	if water.At(0) != 0 {
		t.Fatal("water did not wrap on the upper boundary")
	}
	raster, err := NewWrapBank(WrapBankConfig{
		Start: []float64{120}, Velocity: []float64{-2},
		Lower: &WrapLimit{Boundary: -25, Restart: 120, Inclusive: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	for tick := 0; tick < 72; tick++ {
		raster.Step()
	}
	if raster.At(0) != -24 {
		t.Fatal(raster.At(0))
	}
	raster.Step()
	if raster.At(0) != 120 {
		t.Fatal("raster did not wrap after crossing the lower boundary")
	}
}

func TestWrapBankRelativeRulePreservesOvershoot(t *testing.T) {
	bank, err := NewWrapBank(WrapBankConfig{
		Start: []float64{-252.5, -255}, Velocity: []float64{-8, -2.5},
		Lower: &WrapLimit{Boundary: -256, Restart: 256, Inclusive: true, Relative: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	bank.Step()
	if bank.At(0) != -4.5 || bank.At(1) != -1.5 {
		t.Fatalf("relative wrap lost fractional overshoot: %v, %v", bank.At(0), bank.At(1))
	}
}
