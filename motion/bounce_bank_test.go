package motion

import "testing"

func TestBounceBankPreservesRasterOvershootAndReset(t *testing.T) {
	starts := []float64{94, 124, 154}
	velocities := []float64{2}
	bank, err := NewBounceBank(BounceBankConfig{Start: starts, Velocity: velocities, Min: 94, Max: 160})
	if err != nil {
		t.Fatal(err)
	}
	starts[0], velocities[0] = 0, 99
	legacyPositions := []float64{94, 124, 154}
	legacyVelocities := []float64{2, 2, 2}
	for tick := 0; tick < 140; tick++ {
		bank.Step()
		for index := range legacyPositions {
			legacyPositions[index] += legacyVelocities[index]
			if legacyPositions[index] < 94 || legacyPositions[index] > 160 {
				legacyVelocities[index] = -legacyVelocities[index]
			}
			if got := bank.At(index); got != legacyPositions[index] {
				t.Fatalf("tick %d item %d = %v, want %v", tick, index, got, legacyPositions[index])
			}
		}
	}
	bank.Reset()
	if bank.At(0) != 94 || bank.At(1) != 124 || bank.At(2) != 154 {
		t.Fatalf("reset did not restore configured positions")
	}
	if allocs := testing.AllocsPerRun(100, bank.Step); allocs != 0 {
		t.Fatalf("bounce step allocates %v times", allocs)
	}
}

func TestBounceBankCanReverseAtBoundaryAndClamp(t *testing.T) {
	bank, err := NewBounceBank(BounceBankConfig{
		Start: []float64{2, 8}, Velocity: []float64{-3, 3}, Min: 0, Max: 10,
		Inclusive: true, Clamp: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	bank.Step()
	if bank.At(0) != 0 || bank.At(1) != 10 {
		t.Fatalf("clamped boundary positions = %v, %v", bank.At(0), bank.At(1))
	}
	bank.Step()
	if bank.At(0) != 3 || bank.At(1) != 7 {
		t.Fatalf("reversed directions = %v, %v", bank.At(0), bank.At(1))
	}
}

func TestBounceBankDirectionalEntranceMatchesMegaScroller(t *testing.T) {
	if _, err := NewBounceBank(BounceBankConfig{Start: []float64{45}, Velocity: []float64{-2}, Min: -70, Max: 20}); err == nil {
		t.Fatal("accepted an out-of-range start without entrance mode")
	}
	bank, err := NewBounceBank(BounceBankConfig{
		Start: []float64{45}, Velocity: []float64{-2}, Min: -70, Max: 20,
		Inclusive: true, Directional: true, AllowOutsideStart: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	position, velocity := 45.0, -2.0
	for tick := 0; tick < 1000; tick++ {
		position += velocity
		if position >= 20 {
			velocity = -2
		}
		if position <= -70 {
			velocity = 2
		}
		bank.Step()
		if got := bank.At(0); got != position {
			t.Fatalf("tick %d = %v, want %v", tick, got, position)
		}
	}
	if allocations := testing.AllocsPerRun(100, bank.Step); allocations != 0 {
		t.Fatalf("directional bounce allocates %v times", allocations)
	}
}
