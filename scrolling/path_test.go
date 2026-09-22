package scrolling

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

func TestPathMapperPreservesGlyphScaleAndRotation(t *testing.T) {
	p, err := motion.NewPolyline([]motion.Point{{X: 100, Y: 100}, {X: 100, Y: 300}}, false)
	if err != nil {
		t.Fatal(err)
	}
	mode, err := AlongPath(PathConfig{Path: p, Orient: true, NormalOffset: 5, Clip: true})
	if err != nil {
		t.Fatal(err)
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(2, 3)
	op.GeoM.Translate(30, 80)
	if !mode.Map(Sample{X: 30, Y: 80}, &op) {
		t.Fatal("visible glyph hidden")
	}
	x, y := op.GeoM.Apply(0, 0)
	if math.Abs(x-95) > 1e-9 || math.Abs(y-130) > 1e-9 {
		t.Fatal(x, y)
	}
	x, y = op.GeoM.Apply(1, 1)
	if math.Abs(x-92) > 1e-9 || math.Abs(y-132) > 1e-9 {
		t.Fatal("scale/orientation changed", x, y)
	}
	if mode.Map(Sample{X: -1}, &op) || mode.Map(Sample{X: 201}, &op) {
		t.Fatal("open path did not clip")
	}
}
func TestCustomPathReceivesDistanceAndTime(t *testing.T) {
	mode, err := AlongPath(PathConfig{Vertical: true, Offset: 2, Sample: func(d, t float64) (motion.Point, motion.Point) {
		return motion.Point{X: d, Y: t}, motion.Point{X: 1, Y: 0}
	}})
	if err != nil {
		t.Fatal(err)
	}
	op := ebiten.DrawImageOptions{}
	op.GeoM.Translate(99, 20)
	mode.Map(Sample{X: 99, Y: 20, Time: 7}, &op)
	x, y := op.GeoM.Apply(0, 0)
	if x != 22 || y != 7 {
		t.Fatal(x, y)
	}
	if _, err := AlongPath(PathConfig{}); err == nil {
		t.Fatal("missing path accepted")
	}
}
