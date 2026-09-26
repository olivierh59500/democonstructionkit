package motion

import (
	"math"
	"testing"
)

func sourceDNARibbonSlices(curve int) []TwistingRibbonSlice {
	poses := make([]TwistingRibbonSlice, 0, 20)
	position := curve
	for x := 0; x < 320; x += 16 {
		amplitude, scale := 15.0, 1.5
		decal := float64(position) * 8 * math.Pi / 1280
		if position >= 1280 {
			amplitude, scale = 30, 1
			decal = float64(position-1280) * 8 * math.Pi / 2560
		}
		position += 6
		if position > 3840 {
			position -= 3840
		}
		a1 := math.Mod(math.Pi+.4+decal, 2*math.Pi)
		a2 := math.Mod(a1+1.12, 2*math.Pi)
		y1 := math.Floor(amplitude*math.Sin(a1) + .5)
		y2 := math.Floor(amplitude*math.Sin(a2) + .5)
		pose := TwistingRibbonSlice{SourceX: x}
		if a1 > 3.6 || a1 < 1.5 {
			from, to := y1, y2
			if a1 > 3.6 && a1 < 4.6 {
				from = -amplitude
			}
			if a1 > .5 && a1 < 1.5 {
				to = amplitude
			}
			h := (to - from) / amplitude / scale
			if h > .075 {
				pose.BackVisible = true
				pose.BackY = amplitude + 30 + to
				pose.BackScaleY = -h
			}
		}
		if a1 > .5 && a1 < 4.6 {
			from, to := y2, y1
			if a1 > .5 && a1 < 1.5 {
				to = amplitude
			}
			if a1 > 3.6 && a1 < 4.6 {
				from = -amplitude
			}
			h := (to - from) / amplitude / scale
			if h > .075 {
				pose.FrontVisible = true
				pose.FrontY = amplitude + 30 + from
				pose.FrontScaleY = h
			}
		}
		poses = append(poses, pose)
	}
	return poses
}

func TestTwistingRibbonMatchesEverySourceSliceAcrossTwoWraps(t *testing.T) {
	config := TwistingRibbonConfig{
		Width: 320, StripWidth: 16, IndexStep: 6,
		PhaseStep: 8, PhaseWrap: 3840, FarStart: 1280,
		NearAmplitude: 15, FarAmplitude: 30, NearScale: 1.5, FarScale: 1,
		AngleMultiplier: 8, NearDivisor: 1280, FarDivisor: 2560,
		BaseAngle: math.Pi + .4, FaceGap: 1.12, BaseY: 30,
		BackVisibleBelow: 1.5, BackVisibleAbove: 3.6,
		FrontVisibleAbove: .5, FrontVisibleBelow: 4.6,
		UpperClipStart: .5, UpperClipEnd: 1.5,
		LowerClipStart: 3.6, LowerClipEnd: 4.6,
		MinimumHeight: .075,
	}
	ribbon, err := NewTwistingRibbon(config)
	if err != nil {
		t.Fatal(err)
	}
	curve := 0
	for frame := 0; frame < 1000; frame++ {
		if ribbon.Phase() != curve {
			t.Fatalf("frame %d phase %d, want %d", frame, ribbon.Phase(), curve)
		}
		want := sourceDNARibbonSlices(curve)
		for index, pose := range ribbon.Poses() {
			if pose != want[index] {
				t.Fatalf("frame %d slice %d = %+v, want %+v", frame, index, pose, want[index])
			}
		}
		curve += 8
		if curve > 3840 {
			curve -= 3840
		}
		ribbon.Advance()
	}
	if allocations := testing.AllocsPerRun(100, func() { ribbon.Advance() }); allocations != 0 {
		t.Fatalf("ribbon advance allocated %v times", allocations)
	}
}

func TestTwistingRibbonRejectsInvalidGeometry(t *testing.T) {
	for _, config := range []TwistingRibbonConfig{
		{},
		{Width: 320, StripWidth: 0},
		{Width: 320, StripWidth: 16, PhaseWrap: 3840, NearAmplitude: 15, FarAmplitude: 30,
			NearScale: 1, FarScale: 1, NearDivisor: 1280, FarDivisor: math.NaN()},
	} {
		if _, err := NewTwistingRibbon(config); err == nil {
			t.Fatalf("accepted %+v", config)
		}
	}
}
