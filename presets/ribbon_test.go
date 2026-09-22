package presets

import (
	"math"
	"slices"
	"testing"
)

const (
	cdZero = iota
	cdSlowSin
	cdMedSin
	cdFastSin
	cdSlowDist
	cdMedDist
	cdFastDist
	cdSplitted
	bgSin1
	bgSin2
	bgSin3
)

// The independent reference fixes the authored samples before generic compilation.
func referenceRibbonCurves(distortionRate float64) [][]int {
	curves := make([][]int, bgSin3+1)
	for curveType := cdZero; curveType <= bgSin3; curveType++ {
		var step, progress float64
		switch curveType {
		case cdZero:
			step = 2.25
		case cdSlowSin:
			step, progress = 0.20, 140
		case cdMedSin:
			step, progress = 0.25, 175
		case cdFastSin:
			step, progress = 0.30, 210
		case cdSlowDist:
			step, progress = 0.12, 175
		case cdMedDist:
			step, progress = 0.16, 210
		case cdFastDist:
			step, progress = 0.20, 245
		case cdSplitted:
			step = 0.18
		case bgSin1:
			step = 0.50
		case bgSin2:
			step = 0.80
		case bgSin3:
			step = 0.50
		}
		step *= distortionRate

		maxAngle := 360.0
		if curveType == cdSplitted {
			maxAngle = 720
		}
		values := make([]float64, 0, int(maxAngle/step)+1)
		for angle := 0.0; angle < maxAngle-step; angle += step {
			radians := angle * math.Pi / 180
			var value float64
			switch curveType {
			case cdZero:
			case cdSlowSin:
				value = 100 * math.Sin(radians)
			case cdMedSin:
				value = 110 * math.Sin(radians)
			case cdFastSin:
				value = 120 * math.Sin(radians)
			case cdSlowDist:
				value = 100*math.Sin(radians) + 25*math.Sin(radians*10)
			case cdMedDist:
				value = 110*math.Sin(radians) + 27.5*math.Sin(radians*9)
			case cdFastDist:
				value = 120*math.Sin(radians) + 30*math.Sin(radians*8)
			case cdSplitted:
				direction := 1.0
				if len(values)%2 == 1 {
					direction = -1
				}
				amplitude := 12.0
				if angle < 160 {
					amplitude *= angle / 160
				} else if angle > 560 {
					amplitude *= (720 - angle) / 160
				}
				value = 90*math.Sin(radians) + direction*amplitude*math.Sin(radians*3)
			case bgSin1, bgSin2:
				value = -60 * math.Sin(radians)
			case bgSin3:
				value = -60*math.Sin(radians) - 15*math.Sin(radians*4)
			}
			values = append(values, value)
		}

		curve := make([]int, len(values))
		decal := 0.0
		previous := 0
		for i, value := range values {
			item := -int(math.Floor(value - decal))
			curve[i] = item - previous
			previous = item
			decal += progress / float64(len(values))
		}
		curves[curveType] = curve
	}
	return curves
}

func TestRibbonCurvePresetPreservesEveryIntegerSample(t *testing.T) {
	for _, rate := range []float64{.5, .75, 1, 1.25, 2} {
		got, err := RibbonCurves(rate)
		if err != nil {
			t.Fatal(err)
		}
		want := referenceRibbonCurves(rate)
		for i := range want {
			if !slices.Equal(got[i], want[i]) {
				for j := range want[i] {
					if j >= len(got[i]) || got[i][j] != want[i][j] {
						t.Fatalf("rate %g curve %d sample %d: generated mismatch", rate, i, j)
					}
				}
				t.Fatalf("rate %g curve %d: length mismatch", rate, i)
			}
		}
	}
}
