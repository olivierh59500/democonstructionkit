package composite

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestMagnifierCompilesAndCloses(t *testing.T) {
	m, err := NewMagnifier()
	if err != nil {
		t.Fatal(err)
	}
	src, dst := ebiten.NewImage(32, 32), ebiten.NewImage(32, 32)
	defer src.Deallocate()
	defer dst.Deallocate()
	c := DefaultMagnifierOptions()
	c.CenterX, c.CenterY = 16, 16
	m.Draw(dst, src, c)
	c.Filter = ebiten.FilterLinear
	m.Draw(dst, src, c)
	c.Radius = math.NaN()
	m.Draw(dst, src, c)
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
	m.Draw(dst, src, DefaultMagnifierOptions())
	if err = m.Close(); err != nil {
		t.Fatal(err)
	}
}
