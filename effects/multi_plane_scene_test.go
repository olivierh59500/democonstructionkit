package effects

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/composite"
	bitmap "github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/scrolling"
	"github.com/olivierh59500/democonstructionkit/sprites"
)

func TestMultiPlaneSceneKeepsNativeAndDirectSurfaceOwnership(t *testing.T) {
	mountains, logo, atlas := ebiten.NewImage(8, 8), ebiten.NewImage(8, 8), ebiten.NewImage(8, 8)
	defer mountains.Deallocate()
	defer logo.Deallocate()
	defer atlas.Deallocate()
	metrics, err := bitmap.NewGrid(bitmap.Grid{Bounds: atlas.Bounds(), Cell: image.Pt(8, 8), Columns: 1, Order: "A"})
	if err != nil {
		t.Fatal(err)
	}
	config := MultiPlaneSceneConfig{
		Mountains: mountains, Logo: logo,
		LogoSource:   image.Rect(0, 2, 8, 10), // The missing rows must remain transparent.
		CenterSource: image.Rect(0, 0, 8, 8),
		Bands:        composite.BandsConfig{Bands: []composite.MovingBand{{Source: mountains.Bounds()}}},
		Rows:         composite.ProfileImageConfig{Offsets: []float64{0, 1}, PhaseStep: 1, RowStep: 1, PhaseWrap: 1, Batch: true},
		Center:       sprites.AxisFlipConfig{Saw: &motion.SawToggleConfig{Start: 0, Velocity: 1, Boundary: 1, Restart: -1}},
		Scroll: scrolling.Config{Projected: &scrolling.ProjectedConfig{
			Planes: scrolling.PlanesConfig{
				Slots: []scrolling.PlaneSlot{{Rune: 'A', Advance: 8}}, Forms: []scrolling.PlaneForm{{Height: 8}},
				Visible: 1, Projection: scrolling.PlaneProjection{Focal: 100, Depth: 100},
			},
			Face: scrolling.Face{Atlas: atlas, Metrics: metrics}, PixelsPerUpdate: 1,
			Draw: scrolling.PlaneDraw{ScaleX: 1, ScaleY: 1},
		}},
		Viewport: image.Rect(0, 0, 64, 40), StageSize: image.Pt(32, 20),
		CenterX: 16, CenterY: 10,
	}
	for _, native := range []bool{false, true} {
		config.NativeStage = native
		scene, err := NewMultiPlaneScene(config)
		if err != nil {
			t.Fatal(err)
		}
		if (scene.stage != nil) != native || (scene.background != nil) != native {
			t.Fatalf("native %v allocated wrong surfaces", native)
		}
		if err := scene.Update(kit.Frame{Tick: 1}); err != nil {
			t.Fatal(err)
		}
		dst := ebiten.NewImage(64, 40)
		scene.Draw(dst)
		dst.Deallocate()
		if err := scene.Close(); err != nil {
			t.Fatal(err)
		}
		if err := scene.Close(); err != nil {
			t.Fatal(err)
		}
	}
}
