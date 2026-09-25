package sprites

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
	"github.com/olivierh59500/democonstructionkit/scrolling"
)

// AtlasConfig describes a row-major sprite sheet. Zero Columns and Count use
// all complete cells in the source image. FallbackWhole keeps one borrowed
// full-image frame if a deliberately irregular sheet cannot be sliced.
type AtlasConfig struct {
	Image          *ebiten.Image
	TileW, TileH   int
	Columns, Count int
	FallbackWhole  bool
}

// Atlas caches borrowed sub-images once. The exported dimensions allow a
// background grid or editor to reuse the same atlas without re-slicing it.
type Atlas struct {
	Image   *ebiten.Image
	Tiles   []*ebiten.Image
	TileW   int
	TileH   int
	Columns int
}

func NewAtlas(config AtlasConfig) (*Atlas, error) {
	if config.Image == nil || config.TileW < 1 || config.TileH < 1 || config.Columns < 0 || config.Count < 0 {
		return nil, fmt.Errorf("sprites: invalid atlas image or cell geometry")
	}
	bounds := config.Image.Bounds()
	columns := config.Columns
	if columns == 0 {
		columns = max(1, bounds.Dx()/config.TileW)
	}
	rows := max(1, bounds.Dy()/config.TileH)
	count := config.Count
	if count == 0 {
		count = columns * rows
	}
	tiles, err := scrolling.GridImages(config.Image, image.Pt(config.TileW, config.TileH), columns, count)
	if err != nil {
		if !config.FallbackWhole {
			return nil, err
		}
		tiles, columns = []*ebiten.Image{config.Image}, 1
	}
	return &Atlas{Image: config.Image, Tiles: tiles, TileW: config.TileW, TileH: config.TileH, Columns: columns}, nil
}

func (atlas *Atlas) Count() int { return len(atlas.Tiles) }

func (atlas *Atlas) normalized(index int) int {
	if index < 0 {
		return 0
	}
	if index >= len(atlas.Tiles) {
		return index % len(atlas.Tiles)
	}
	return index
}

// Tile clamps negative indices and wraps positive overflow, retaining the
// authored atlas sequence. Its source image is borrowed, not copied.
func (atlas *Atlas) Tile(index int) *ebiten.Image {
	if len(atlas.Tiles) == 0 {
		return atlas.Image
	}
	return atlas.Tiles[atlas.normalized(index)]
}

// Region returns the absolute source crop for a renderer that must keep
// fractional or triangle-based sampling instead of drawing a sub-image.
func (atlas *Atlas) Region(index int) composite.Region {
	if len(atlas.Tiles) == 0 {
		b := atlas.Image.Bounds()
		return composite.Region{X: float64(b.Min.X), Y: float64(b.Min.Y), Width: float64(b.Dx()), Height: float64(b.Dy())}
	}
	index = atlas.normalized(index)
	b := atlas.Image.Bounds()
	return composite.Region{
		X:     float64(b.Min.X + index%atlas.Columns*atlas.TileW),
		Y:     float64(b.Min.Y + index/atlas.Columns*atlas.TileH),
		Width: float64(atlas.TileW), Height: float64(atlas.TileH),
	}
}
