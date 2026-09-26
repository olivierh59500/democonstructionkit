package sprites

import (
	"image/color"
	"testing"

	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

func TestRotatingDiscCloudRetainsSortedPosesAndTint(t *testing.T) {
	cloud, err := NewRotatingDiscCloud(RotatingDiscCloudConfig{
		Motion: geometry.RotatingDiscCloudConfig{
			Points:  []geometry.Vec3{{X: 10, Z: -5}, {X: -10, Z: 5}},
			CenterX: 50, CenterY: 40, DepthBase: 100, Focal: 50,
			AngleStep: .03, RadiusScale: 1,
		},
		Tint: color.RGBA{R: 238, G: 136, A: 255}, Antialias: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cloud.Close()
	if err := cloud.Update(kit.Frame{}); err != nil {
		t.Fatal(err)
	}
	poses := cloud.Motion().Poses()
	if len(poses) != 2 || poses[0].Z > poses[1].Z || !cloud.renderer.Antialias {
		t.Fatalf("disc order or material changed: %+v", poses)
	}
	for i, disc := range cloud.discs {
		if disc.X != poses[i].X || disc.Y != poses[i].Y || disc.Radius != poses[i].Radius ||
			disc.ColorScale.R() != cloud.tint.R() || disc.ColorScale.G() != cloud.tint.G() {
			t.Fatalf("disc %d material differs: %+v", i, disc)
		}
	}
}
