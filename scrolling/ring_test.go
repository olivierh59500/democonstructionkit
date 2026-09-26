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

func TestRingSeamlessSeedAvoidsInitialShortMessageGap(t *testing.T) {
	img := ebiten.NewImage(16, 8)
	defer img.Deallocate()
	config := RingConfig{Text: "AB", Font: BitmapGrid{Image: img, Width: 8, Height: 8, Columns: 2, First: 'A'},
		Viewport: 32, Speed: 8}
	legacy, err := NewRing(config)
	if err != nil {
		t.Fatal(err)
	}
	config.SeamlessSeed = true
	seamless, err := NewRing(config)
	if err != nil {
		t.Fatal(err)
	}
	for i, letter := range seamless.letters {
		want := rune('A' + i%2)
		if letter.Rune != want {
			t.Fatalf("initial slot %d = %q, want %q", i, letter.Rune, want)
		}
	}
	if legacy.letters[2].Rune != -1 || seamless.Cursor() != 0 || legacy.Cursor() != len(legacy.letters) {
		t.Fatal("legacy spacing or seamless cursor changed")
	}
	for range 6 {
		legacy.Step()
		seamless.Step()
	}
	if legacy.letters[0].Rune != -1 || seamless.letters[0].Rune != 'A' || seamless.Cursor() != 1 {
		t.Fatal("first recycled glyph did not bridge the short-text loop")
	}
}
