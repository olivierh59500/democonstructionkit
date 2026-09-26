package motion

import (
	"math"
	"testing"
)

func TestSampledRowsKeepCuddlyDNALogoCoordinates(t *testing.T) {
	outer, err := NewOuterSineRows(OuterSineRowsConfig{
		Bases: []float64{60, 268}, PhaseOffsets: []float64{0, 10},
		BaseY: 100, HeightBase: .5, HeightAmplitude: .75, HeightTimeDivisor: 15,
		BounceAmplitude: 10, BounceTimeDivisor: 15,
		HorizontalAmplitude: 4, HorizontalDivisor: 10, SourceHeight: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	center, err := NewZoomSineRows(ZoomSineRowsConfig{
		BaseX: 212, BaseY: 120, ReverseSourceStart: 55, SourceFactor: .5, SourceHeight: 2,
		ZoomBase: .75, ZoomAmplitude: .25, ZoomFrequency: 4,
		ZoomTimeDivisor: 67, ZoomIndexDivisor: 131, XShift: 48,
		YAmplitude: 20, YFrequency: 5, YTimeDivisor: 61, YIndexDivisor: 127,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, tick := range []float64{0, 1, 33, 180, 500, 804, 1080, 1952, 4000, 9000} {
		for copyIndex, base := range []float64{60, 268} {
			decal := float64(copyIndex * 10)
			for row := 0; row < 56; row++ {
				height := .5 + (1+math.Sin(tick/15))*.75
				bounce := float64(int(10 * math.Cos(decal+tick/15)))
				offset := float64(int(4 * math.Sin(decal+(tick+float64(row)*height)/10)))
				want := SampledRowPose{SourceY: float64(int(float64(row) * height)), SourceHeight: 1,
					X: base + offset, Y: 100 + bounce + float64(row), ScaleX: 1, ScaleY: 1}
				if got := outer.Sample(tick, copyIndex, row); got != want {
					t.Fatalf("outer tick %v copy %d row %d = %+v, want %+v", tick, copyIndex, row, got, want)
				}
			}
		}
		for row := 0; row < 56; row++ {
			zoom := math.Sin(4*(tick/67+float64(row)/131))/4 + .75
			y := 20 * math.Sin(5*(tick/61+float64(row)*zoom/127))
			want := SampledRowPose{SourceY: float64(int(float64(55-row) * .5)), SourceHeight: 2,
				X: 212 - zoom*48, Y: 120 + y, ScaleX: zoom, ScaleY: zoom}
			if got := center.Sample(tick, 0, row); got != want {
				t.Fatalf("center tick %v row %d = %+v, want %+v", tick, row, got, want)
			}
		}
	}
	if allocations := testing.AllocsPerRun(100, func() {
		outer.Sample(1200, 1, 42)
		center.Sample(1200, 0, 42)
	}); allocations != 0 {
		t.Fatalf("row sample allocated %v times", allocations)
	}
}

func TestSampledRowProgramsRejectInvalidDivisors(t *testing.T) {
	if _, err := NewOuterSineRows(OuterSineRowsConfig{
		Bases: []float64{60}, PhaseOffsets: []float64{0}, SourceHeight: 1,
	}); err == nil {
		t.Fatal("accepted a stalled outer row clock")
	}
	if _, err := NewZoomSineRows(ZoomSineRowsConfig{SourceHeight: 2,
		ZoomTimeDivisor: 67, ZoomIndexDivisor: 131, YTimeDivisor: 61}); err == nil {
		t.Fatal("accepted a zero Y index divisor")
	}
}
