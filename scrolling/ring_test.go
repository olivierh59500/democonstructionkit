package scrolling

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestFractionalRingInitializationAndRecycling(t *testing.T) {
	img := ebiten.NewImage(666, 328)
	defer img.Deallocate()
	r, err := NewRing(RingConfig{Text: "ABCDEFGHIJKLMNOPQRSTUVWXYZ", Font: BitmapGrid{Image: img, Width: 83.25, Height: 41, Columns: 8, First: 32}, Viewport: 568, Speed: 8})
	if err != nil {
		t.Fatal(err)
	}
	if r.letters[0].X != 666 || r.letters[1].X != 750 {
		t.Fatal("fractional initial positions rounded incorrectly")
	}
	for i := 0; i < 94; i++ {
		r.Step()
	}
	if r.letters[0].Rune != 'J' || math.Abs(r.letters[0].X-663.25) > 1e-9 {
		t.Fatal("incorrect recycled slot", r.letters[0])
	}
	src, ok := r.config.Font.Region('!')
	if !ok || src.X != 83.25 || src.Width != 83.25 {
		t.Fatal("fractional crop lost", src)
	}
}
func TestRingControlCallbackAndPause(t *testing.T) {
	img := ebiten.NewImage(96, 8)
	defer img.Deallocate()
	called := 0
	r, err := NewRing(RingConfig{Text: "AAAA^Cnext;^P1ABCD", Font: BitmapGrid{Image: img, Width: 8, Height: 8, Columns: 12, First: 'A'}, Viewport: 16, Speed: 8, Controls: true, Commands: map[string]func(){"next": func() { called++ }}})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		r.Step()
	}
	if called != 1 || r.speed != 0 {
		t.Fatal("command or pause was not applied", called, r.speed)
	}
	for i := 0; i < 60; i++ {
		r.Step()
	}
	if r.speed != 8 {
		t.Fatal("pause did not restore speed")
	}
}
