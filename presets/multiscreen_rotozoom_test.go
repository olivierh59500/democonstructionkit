package presets

import (
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func TestMultiscreenRotozoomPresetsKeepIndependentMaterials(t *testing.T) {
	tests := []struct {
		name           string
		config         VivaRotozoomConfig
		phaseX, phaseY float64
		color          [4]float32
	}{
		{"Coco", MultiscreenCocoRotozoom(800, 600), 3200, 2400, [4]float32{.5, .5, .5, 1}},
		{"Viva", MultiscreenVivaRotozoom(800, 600), 6400, 4800, [4]float32{1, 1, 1, 1}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			program, err := NewVivaRotozoom(test.config)
			if err != nil {
				t.Fatal(err)
			}
			for tick := 1; tick <= 750; tick++ {
				if err := program.Update(kit.Frame{}); err != nil {
					t.Fatal(err)
				}
				pose := program.Repetition()
				x, z, r := .008*float64(tick), .003*float64(tick), .005*float64(tick)
				wantX := 400 + 200*math.Cos(x*4-math.Cos(x-.1))
				wantY := 300 - (600.0/2.7)*math.Sin(x*2.3-math.Cos(x-.1))
				wantZoom := .5 + math.Abs(math.Sin(z)*2.5)
				wantRotation := 90 * math.Cos(r*4-math.Cos(r-.01)) * .3 * math.Pi / 180
				if math.Abs(pose.CenterX-wantX) > 1e-10 || math.Abs(pose.CenterY-wantY) > 1e-10 ||
					math.Abs(pose.Zoom-wantZoom) > 1e-10 || math.Abs(pose.Rotation-wantRotation) > 1e-10 ||
					pose.PhaseX != test.phaseX || pose.PhaseY != test.phaseY || pose.Color != test.color {
					t.Fatalf("%s tick %d pose %+v differs", test.name, tick, pose)
				}
			}
		})
	}
}
