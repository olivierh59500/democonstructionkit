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
	if _, err := sprites.NewProjectedObject(config); err != nil {
		t.Fatal(err)
	}
	if _, err := VectorballsProjectedObject("sphere", sprites.Surface, 33, 640, 120); err == nil {
		t.Fatal("unbounded sphere density accepted")
	}
}
