package sprites

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestFormationCarouselCyclesModesAndKeepsSlideGeometry(t *testing.T) {
	image := ebiten.NewImage(24, 8)
	defer image.Deallocate()
	atlas, err := NewAtlas(AtlasConfig{Image: image, TileW: 8, TileH: 8})
	if err != nil {
		t.Fatal(err)
	}
	carousel, err := NewFormationCarousel(FormationCarouselConfig{
		Atlas: atlas, Count: 3, Hold: 8, Slide: 1, SlideIndex: .5,
		Modes: []motion.FormulaFormationConfig{
			{X: motion.ExprConst(10), Y: motion.ExprIndex()},
			{X: motion.ExprConst(20), Y: motion.ExprIndex()},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, check := range []struct {
		time float64
		x    float64
	}{
		{0, 160}, {1, 60}, {10, 170}, {20, 160},
	} {
		position, frame := carousel.PoseAt(check.time, 2, 100, 100)
		if position.X != check.x || position.Y != 52 || frame != 2 {
			t.Fatalf("time %g = %+v frame %d, want x=%g y=52 frame=2", check.time, position, frame, check.x)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { _, _ = carousel.PoseAt(10, 2, 100, 100) }); allocations != 0 {
		t.Fatalf("carousel pose allocates %v times", allocations)
	}
}
