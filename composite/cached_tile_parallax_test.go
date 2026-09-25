package composite

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestCachedTileParallaxMatchesIntegerMenuCamera(t *testing.T) {
	tile := ebiten.NewImage(32, 32)
	defer tile.Deallocate()
	parallax, err := NewCachedTileParallax(CachedTileParallaxConfig{
		Tile: tile, Width: 768, Height: 400, PeriodX: 32, PeriodY: 32,
		OverscanX: 32, OverscanY: 32, DivisorX: 2, DivisorY: 2,
		WrapX: 32, WrapY: 32, ClampNegative: true, PixelQuantize: true,
		Blend: ebiten.BlendCopy, Unmanaged: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer parallax.Close()
	if bounds := parallax.surface.Bounds(); bounds.Dx() != 800 || bounds.Dy() != 432 {
		t.Fatalf("surface = %v, want 800x432", bounds)
	}
	for _, camera := range []int{-10, 0, 1, 31, 32, 63, 64, 1000} {
		gotX, gotY := parallax.Offset(float64(camera), float64(camera))
		clamped := max(0, camera)
		want := float64(-((clamped / 2) % 32))
		if gotX != want || gotY != want {
			t.Fatalf("camera %d offset = (%v,%v), want %v", camera, gotX, gotY, want)
		}
	}
	if allocations := testing.AllocsPerRun(100, func() { _, _ = parallax.Offset(99, 51) }); allocations != 0 {
		t.Fatalf("parallax sampling allocates %v times", allocations)
	}
}

func TestCachedTileParallaxRejectsExcessiveSurface(t *testing.T) {
	tile := ebiten.NewImage(1, 1)
	defer tile.Deallocate()
	if _, err := NewCachedTileParallax(CachedTileParallaxConfig{Tile: tile, Width: 768, Height: 400, DivisorX: 2, DivisorY: 2}); err == nil {
		t.Fatal("accepted more than 16,384 repeated tiles")
	}
}
