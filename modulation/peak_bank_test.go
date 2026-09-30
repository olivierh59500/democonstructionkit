package modulation

import (
	"math"
	"testing"
)

func TestPeakBankChargesBeforeSameStepDecay(t *testing.T) {
	p, err := NewPeakBank(PeakBankConfig{Count: 3, Ceiling: 100, Decay: 3})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Step([]PeakCharge{{Index: 1, Level: 15}, {Index: 1, Level: 21}, {Index: 2, Level: 30}}, 1); err != nil {
		t.Fatal(err)
	}
	if p.Level(0) != 0 || p.Level(1) != 18 || p.Level(2) != 27 {
		t.Fatal("charges must combine by max before all peaks decay")
	}
	if err := p.Step([]PeakCharge{{Index: 1, Level: 8}}, 2); err != nil {
		t.Fatal(err)
	}
	if p.Level(1) != 12 || p.Level(2) != 21 {
		t.Fatal("a weaker charge must not lower the previous peak")
	}
	p.Reset()
	if p.Level(1) != 0 || p.Level(-1) != 0 || p.Level(3) != 0 || p.Len() != 3 {
		t.Fatal("reset or bounds are incorrect")
	}
}

func TestPeakBankBoundsAndLargeDelta(t *testing.T) {
	p, err := NewPeakBank(PeakBankConfig{Count: 2, Floor: -10, Ceiling: 20, Decay: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !p.Charge(0, 100) || !p.Charge(1, -20) || p.Level(0) != 20 || p.Level(1) != -10 {
		t.Fatal("charges were not clamped to the configured bounds")
	}
	if err := p.Step(nil, math.MaxFloat64); err != nil {
		t.Fatal(err)
	}
	if p.Level(0) != -10 || p.Level(1) != -10 {
		t.Fatal("an overflowing decay must safely reach the floor")
	}
}

func TestPeakBankRejectsInvalidStepBeforeMutation(t *testing.T) {
	p, err := NewPeakBank(PeakBankConfig{Count: 2, Ceiling: 100, Decay: 1})
	if err != nil {
		t.Fatal(err)
	}
	p.Charge(0, 5)
	for _, charges := range [][]PeakCharge{
		{{Index: 0, Level: 20}, {Index: 2, Level: 10}},
		{{Index: 0, Level: 20}, {Index: 1, Level: math.NaN()}},
		{{Index: -1, Level: 10}},
	} {
		if err := p.Step(charges, 1); err == nil || p.Level(0) != 5 {
			t.Fatal("invalid charge changed the bank")
		}
	}
	for _, delta := range []float64{-1, math.NaN(), math.Inf(1)} {
		if err := p.Step(nil, delta); err == nil || p.Level(0) != 5 {
			t.Fatal("invalid delta changed the bank")
		}
	}
	if p.Charge(1, math.Inf(1)) || p.Charge(2, 3) {
		t.Fatal("invalid immediate charge accepted")
	}
}

func TestPeakBankValidatesConfiguration(t *testing.T) {
	for _, config := range []PeakBankConfig{
		{}, {Count: 65537}, {Count: 1, Floor: 1}, {Count: 1, Decay: -1},
		{Count: 1, Ceiling: math.Inf(1)}, {Count: 1, Floor: math.NaN()},
		{Count: 1, Floor: -math.MaxFloat64, Ceiling: math.MaxFloat64},
	} {
		if _, err := NewPeakBank(config); err == nil {
			t.Fatalf("accepted invalid config: %+v", config)
		}
	}
	var p *PeakBank
	if p.Len() != 0 || p.Level(0) != 0 || p.Charge(0, 1) || p.Step(nil, 1) == nil {
		t.Fatal("nil bank was not handled")
	}
	p.Reset()
}

func TestPeakBankDoesNotAllocatePerStep(t *testing.T) {
	p, err := NewPeakBank(PeakBankConfig{Count: 80, Ceiling: 100, Decay: 3})
	if err != nil {
		t.Fatal(err)
	}
	charges := []PeakCharge{{Index: 11, Level: 95.301}, {Index: 25, Level: 60.66}}
	if allocations := testing.AllocsPerRun(100, func() { p.Step(charges, 1) }); allocations != 0 {
		t.Fatalf("allocated %v times per step", allocations)
	}
}
