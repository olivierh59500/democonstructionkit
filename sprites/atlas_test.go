package sprites

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestAtlasCachesRowMajorTilesAndRegions(t *testing.T) {
	image := ebiten.NewImage(60, 44)
	defer image.Deallocate()
	atlas, err := NewAtlas(AtlasConfig{Image: image, TileW: 30, TileH: 22})
	if err != nil {
		t.Fatal(err)
	}
	if atlas.Count() != 4 || atlas.Columns != 2 || atlas.Tile(-1) != atlas.Tile(0) || atlas.Tile(4) != atlas.Tile(0) {
		t.Fatal("atlas count, columns or index policy changed")
	}
	region := atlas.Region(3)
	if region.X != 30 || region.Y != 22 || region.Width != 30 || region.Height != 22 {
		t.Fatalf("last region = %+v", region)
	}
	if rect := atlas.Rect(3); rect.Min.X != 30 || rect.Min.Y != 22 || rect.Max.X != 60 || rect.Max.Y != 44 {
		t.Fatalf("integer atlas crop = %v", rect)
	}
	if allocations := testing.AllocsPerRun(100, func() { _ = atlas.Tile(3); _ = atlas.Region(3) }); allocations != 0 {
		t.Fatalf("atlas sampling allocates %v times", allocations)
	}
}

func TestAtlasOptionalWholeImageFallback(t *testing.T) {
	image := ebiten.NewImage(20, 10)
	defer image.Deallocate()
	if _, err := NewAtlas(AtlasConfig{Image: image, TileW: 30, TileH: 22}); err == nil {
		t.Fatal("invalid grid silently fell back")
	}
	atlas, err := NewAtlas(AtlasConfig{Image: image, TileW: 30, TileH: 22, FallbackWhole: true})
	if err != nil {
		t.Fatal(err)
	}
	if atlas.Count() != 1 || atlas.Columns != 1 || atlas.Tile(9) != image {
		t.Fatal("whole-image fallback differs from authored behavior")
	}
}
