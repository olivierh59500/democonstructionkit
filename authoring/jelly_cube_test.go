package authoring

import (
	"bytes"
	"math"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
)

func TestIndependentCubeProjectsRoundTrip(t *testing.T) {
	zero, half, phase := 0.0, 40.0, 7.0
	p := Project{Version: Version, Units: DefaultUnits(), Canvas: Canvas{Width: 640, Height: 360, TPS: 60}, Layers: []Layer{
		{ID: "first", Kind: "jelly_cube", JellyCube: &JellyCube{Center: &Point{X: 150, Y: 180}, HalfEdge: &half, Transition: &zero}},
		{ID: "second", Kind: "jelly_cube", JellyCube: &JellyCube{Center: &Point{X: 450, Y: 180}, HalfEdge: &half, Phase: &phase, Steps: []JellyCubeStep{{Mode: "bounce", Duration: 2}, {Mode: "swing", Duration: 3}}}},
	}}
	var saved bytes.Buffer
	if err := Encode(&saved, p); err != nil {
		t.Fatal(err)
	}
	loaded, err := Decode(&saved)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := jellyConfig(*loaded.Layers[0].JellyCube)
	if err != nil || cfg.Transition != 0 {
		t.Fatal("explicit zero was lost", cfg, err)
	}
	compiled, err := Compile(*loaded, Assets{}, Options{})
	if err != nil {
		t.Fatal(err)
	}
	defer compiled.Close()
	for _, time := range []float64{0, 1, 5, 2, 10} {
		if err := compiled.Update(kit.Frame{Time: time}); err != nil {
			t.Fatal(err)
		}
	}
}

func TestSavedCubeRejectsUnboundedOrUnknownConfiguration(t *testing.T) {
	huge, negative := 1e20, -1.0
	for _, c := range []JellyCube{{Preset: "unknown"}, {Phase: &huge}, {HalfEdge: &negative}, {Steps: []JellyCubeStep{}}, {Steps: []JellyCubeStep{{Mode: "unknown", Duration: 2}}}, {Speed: func() *float64 { v := math.NaN(); return &v }()}} {
		if _, err := jellyConfig(c); err == nil {
			t.Fatal("invalid cube accepted", c)
		}
	}
}
