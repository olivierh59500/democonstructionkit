package presets

import (
	"testing"

	"github.com/olivierh59500/democonstructionkit/sprites"
)

func TestVectorballsSpherePresetUsesBoundedDensity(t *testing.T) {
	config, err := VectorballsProjectedObject("sphere", sprites.Edges, 6, 640, 120)
	if err != nil {
		t.Fatal(err)
	}
	if config.Sphere == nil || config.Sphere.Count != 144 || config.Sphere.Radius != 320 || config.Sphere.Image != 120 {
		t.Fatalf("sphere preset = %+v", config.Sphere)
	}
	if config.AscendingDepth {
		t.Fatal("camera must draw larger, farther Z before smaller, nearer Z")
	}
	if _, err := sprites.NewProjectedObject(config); err != nil {
		t.Fatal(err)
	}
	if _, err := VectorballsProjectedObject("sphere", sprites.Surface, 33, 640, 120); err == nil {
		t.Fatal("unbounded sphere density accepted")
	}
	for _, test := range []struct {
		segments, image int
		cull            bool
	}{
		{6, 10, false}, {16, 10, true}, {32, 8, true},
	} {
		config, err := VectorballsProjectedObject("sphere", sprites.Edges, test.segments, 640, -1)
		if err != nil {
			t.Fatal(err)
		}
		if config.Sphere.Image != test.image || config.CullPositiveModelZ != test.cull ||
			config.ScaleImages != test.cull || !config.Batch {
			t.Fatalf("segments %d: default sphere material %+v, cull %v, scaled %v, batch %v",
				test.segments, config.Sphere, config.CullPositiveModelZ, config.ScaleImages, config.Batch)
		}
	}
	if _, err := VectorballsProjectedObject("sphere", sprites.Edges, 6, 640, -2); err == nil {
		t.Fatal("invalid negative sprite index accepted")
	}
}
