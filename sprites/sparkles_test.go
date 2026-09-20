package sprites

import (
	"github.com/hajimehoshi/ebiten/v2"
	"testing"
)

func TestSparklesAdvanceWithoutDependingOnDrawing(t *testing.T) {
	img := ebiten.NewImage(13, 7)
	defer img.Deallocate()
	s, err := NewSparkles(SparkleConfig{Images: []SparkleImage{{Image: img, Spin: 10}, {Image: img, Angles: []float64{0, 45}}}, Positions: []SparklePoint{{1, 2}, {3, 4}}, StartScale: 1, EndScale: 0, ScaleStep: -.25, PauseTicks: 3})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		s.Advance()
	}
	if s.visible || s.pause != 2 || s.rotation != 40 {
		t.Fatal("incorrect lifetime or first pause tick")
	}
	s.Advance()
	s.Advance()
	if !s.visible || s.image != 1 || s.position != 1 || s.scale != 1 {
		t.Fatal("sprite sequence did not resume")
	}
	s.Advance()
	if s.rotation != 45 {
		t.Fatal("alternating orientation missing")
	}
}
