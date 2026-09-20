package scrolling

import (
	"math"
	"sort"
	"testing"
)

func TestPlaneRecurrenceAgainstOriginalEquations(t *testing.T) {
	forms := []PlaneForm{{Height: 55, VerticalPhase: 1.5}, {DepthAmplitude: 200, DepthStep: -.3, DepthSpeed: 4, Height: 55, VerticalStep: .3, VerticalSpeed: 2, VerticalPhase: 1.5}, {DepthAmplitude: 200, DepthStep: .4, DepthSpeed: -4, DepthPhase: 5, Height: -70, VerticalStep: .4, VerticalSpeed: -4, VerticalPhase: 1.5}}
	slots := make([]PlaneSlot, 43)
	for i := range slots {
		slots[i] = PlaneSlot{Rune: rune('A' + i%26), Advance: 32, Form: -1}
	}
	slots[7].Form = 1
	slots[18].Form = 2
	slots[34].Form = 0
	p, err := NewPlanes(PlanesConfig{Forms: forms, Slots: slots, Visible: 30, PhaseStep: .02, Projection: PlaneProjection{Focal: 250, Depth: 150, OriginX: -450, CenterX: 160, CenterY: 100, XBias: -16, YBias: -14, VerticalOffset: -4}})
	if err != nil {
		t.Fatal(err)
	}
	want := make([]PlanePoint, 30)
	for frame := 0; frame < 700; frame++ {
		form := p.form
		phase := p.phase + .02
		for i := range want {
			index := (p.first + i) % len(slots)
			if slots[index].Form >= 0 {
				form = slots[index].Form
			}
			f := forms[form]
			z := f.DepthAmplitude*math.Sin(f.DepthPhase+float64(index)*f.DepthStep+phase*f.DepthSpeed) + 150
			y := f.Height*math.Cos(1.5+float64(index)*f.VerticalStep+phase*f.VerticalSpeed) - 4
			scale := 250 / (250 + z)
			x := -450 + float64(i)*32 - p.offset
			want[i] = PlanePoint{X: (x-16)*scale + 160, Y: (y-14)*scale + 100, Scale: scale, Rune: slots[index].Rune, Index: index}
		}
		sort.SliceStable(want, func(i, j int) bool { return want[i].Scale < want[j].Scale })
		if err := p.Step(4); err != nil {
			t.Fatal(err)
		}
		for i, got := range p.Points() {
			w := want[i]
			if got.Rune != w.Rune || math.Abs(got.X-w.X) > 1e-9 || math.Abs(got.Y-w.Y) > 1e-9 || math.Abs(got.Scale-w.Scale) > 1e-12 {
				t.Fatalf("frame %d point %d: %+v != %+v", frame, i, got, w)
			}
		}
	}
	if n := testing.AllocsPerRun(1000, func() { p.Step(4) }); n != 0 {
		t.Fatal("plane step allocates", n)
	}
}

func TestPlanesUseProportionalAdvancesAndLargeSteps(t *testing.T) {
	p, err := NewPlanes(PlanesConfig{Forms: []PlaneForm{{}}, Slots: []PlaneSlot{{'A', 7, -1}, {'B', 19, -1}}, Visible: 3, Projection: PlaneProjection{Focal: 100}})
	if err != nil {
		t.Fatal(err)
	}
	p.Step(26*1000 + 8)
	if p.first != 1 || p.offset != 1 {
		t.Fatal("large proportional step failed")
	}
	p.Step(0)
	if p.points[1].X-p.points[0].X != 19 || p.points[2].X-p.points[1].X != 7 {
		t.Fatal("font advances ignored")
	}
}
