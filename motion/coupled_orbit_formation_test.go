package motion

import (
	"math"
	"testing"
)

func TestCoupledOrbitFormationKeepsInterleavedPhaseOrder(t *testing.T) {
	config := CoupledOrbitFormationConfig{
		Orbit: CoupledOrbit{CenterX: 192, CenterY: 135, Radius: 60,
			DepthRadius: 30, PhaseStep: .00025},
		Ranges: []OrbitRange{{Start: 0, Count: 19, Step: 1}, {Start: 40, Count: 19, Step: 1}},
	}
	formation, err := NewCoupledOrbitFormation(config)
	if err != nil {
		t.Fatal(err)
	}
	if formation.Len() != 38 {
		t.Fatalf("formation has %d slots", formation.Len())
	}
	var xPhase, yPhase, zPhase float64
	for tick := 0; tick < 500; tick++ {
		orbit := formation.Orbit()
		if tick < 200 {
			orbit.XIncrement, orbit.YIncrement, orbit.ZIncrement, orbit.QIncrement = 1, -2, -1, -10
			orbit.XOffset, orbit.YOffset, orbit.ZOffset, orbit.QOffset, orbit.QScale = 5, 4, 1, 246, 251
		} else {
			orbit.XIncrement, orbit.YIncrement, orbit.ZIncrement, orbit.QIncrement = 3, -1, 2, -8
			orbit.XOffset, orbit.YOffset, orbit.ZOffset, orbit.QOffset, orbit.QScale = 7, 2, 3, 240, 250
		}
		got := formation.Step()
		for slot := range got {
			index := float64(slot)
			if slot >= 19 {
				index = float64(slot + 21)
			}
			xPhase += orbit.XIncrement * orbit.PhaseStep
			yPhase += orbit.YIncrement * orbit.PhaseStep
			zPhase += orbit.ZIncrement * orbit.PhaseStep
			depth := orbit.DepthRadius * math.Sin(zPhase+index*orbit.ZOffset*.02)
			xq := orbit.QIncrement / (255 - math.Min(254, orbit.QOffset))
			yq := orbit.QIncrement / (255 - math.Min(254, orbit.QScale))
			want := Point{
				X: orbit.CenterX + depth*2 + orbit.Radius*math.Sin(xPhase+index*orbit.XOffset*(.02*xq)),
				Y: orbit.CenterY + depth/2 + orbit.Radius*math.Cos(yPhase+index*orbit.YOffset*(.02*yq)),
			}
			if got[slot] != want {
				t.Fatalf("tick %d slot %d = %+v, want %+v", tick, slot, got[slot], want)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { formation.Step() }); allocations != 0 {
		t.Fatalf("coupled formation allocated %v times per step", allocations)
	}
	formation.Reset()
	if formation.Orbit().XIncrement != config.Orbit.XIncrement || formation.Poses()[0] != (Point{}) {
		t.Fatal("reset retained user controls or old poses")
	}
}

func TestCoupledOrbitFormationRejectsInvalidRanges(t *testing.T) {
	for _, config := range []CoupledOrbitFormationConfig{
		{},
		{Ranges: []OrbitRange{{Count: 0, Step: 1}}},
		{Ranges: []OrbitRange{{Count: 2, Step: 0}}},
		{Orbit: CoupledOrbit{Radius: math.NaN()}, Ranges: []OrbitRange{{Count: 1, Step: 1}}},
	} {
		if _, err := NewCoupledOrbitFormation(config); err == nil {
			t.Fatalf("accepted invalid formation %+v", config)
		}
	}
}
