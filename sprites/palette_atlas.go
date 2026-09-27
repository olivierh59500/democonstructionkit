package sprites

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// PaletteAtlasConfig packs irregular source frames into one row per palette.
// Index returns a zero-based palette entry, or a negative/out-of-range value
// to retain the source pixel. Transparent source pixels are never remapped.
// OriginalFrames are copied after the palette rows without recoloring.
type PaletteAtlasConfig struct {
	Source         image.Image
	Frames         []image.Rectangle
	Palettes       [][]color.RGBA
	Index          func(color.Color) int
	RowHeight      int // Zero uses the tallest frame.
	RepeatLast     bool
	OriginalFrames []image.Rectangle
}

// PaletteAtlas holds CPU-side pixels and ordered crops. Build it during asset
// preparation, then call Upload from an Ebitengine update to create one GPU
// image. Repeated crops share the same image region and consume no atlas space.
type PaletteAtlas struct {
	Pixels *image.RGBA
	Rects  []image.Rectangle
}

// BuildPaletteAtlas performs all recoloring on the CPU. It does not read back
// GPU pixels and preserves source color where Index has no matching entry.
func BuildPaletteAtlas(config PaletteAtlasConfig) (*PaletteAtlas, error) {
	if config.Source == nil || config.Index == nil || len(config.Frames) == 0 ||
		len(config.Frames) > 4096 || len(config.OriginalFrames) > 4096 ||
		len(config.Palettes) == 0 || len(config.Palettes) > 256 || config.RowHeight < 0 {
		return nil, fmt.Errorf("sprites: invalid palette atlas source or frame count")
	}
	maxHeight, paletteWidth, originalWidth := 0, 0, 0
	for _, frame := range config.Frames {
		if frame.Empty() || !frame.In(config.Source.Bounds()) {
			return nil, fmt.Errorf("sprites: palette atlas frame outside source")
		}
		if frame.Dx() > 8192-paletteWidth {
			return nil, fmt.Errorf("sprites: palette atlas exceeds dimensions")
		}
		paletteWidth += frame.Dx()
		maxHeight = max(maxHeight, frame.Dy())
	}
	for _, frame := range config.OriginalFrames {
		if frame.Empty() || !frame.In(config.Source.Bounds()) {
			return nil, fmt.Errorf("sprites: original atlas frame outside source")
		}
		if frame.Dx() > 8192-originalWidth {
			return nil, fmt.Errorf("sprites: palette atlas exceeds dimensions")
		}
		originalWidth += frame.Dx()
		maxHeight = max(maxHeight, frame.Dy())
	}
	for _, palette := range config.Palettes {
		if len(palette) == 0 || len(palette) > 256 {
			return nil, fmt.Errorf("sprites: invalid palette atlas colors")
		}
	}
	rowHeight := config.RowHeight
	if rowHeight == 0 {
		rowHeight = maxHeight
	}
	rows := len(config.Palettes)
	if len(config.OriginalFrames) > 0 {
		rows++
	}
	width := max(paletteWidth, originalWidth)
	if rowHeight < maxHeight || width < 1 || width > 8192 ||
		rowHeight > 8192 || rows > 8192/rowHeight {
		return nil, fmt.Errorf("sprites: palette atlas exceeds dimensions")
	}
	if width > 16_777_216/(rows*rowHeight) {
		return nil, fmt.Errorf("sprites: palette atlas pixel budget exceeded")
	}
	atlas := &PaletteAtlas{
		Pixels: image.NewRGBA(image.Rect(0, 0, width, rows*rowHeight)),
		Rects:  make([]image.Rectangle, 0, len(config.Palettes)*(len(config.Frames)+boolInt(config.RepeatLast))+len(config.OriginalFrames)),
	}
	for paletteIndex, colors := range config.Palettes {
		x, y := 0, paletteIndex*rowHeight
		for _, frame := range config.Frames {
			region := image.Rect(x, y, x+frame.Dx(), y+frame.Dy())
			copyPaletteFrame(atlas.Pixels, region.Min, config.Source, frame, colors, config.Index)
			atlas.Rects = append(atlas.Rects, region)
			x += frame.Dx()
		}
		if config.RepeatLast {
			atlas.Rects = append(atlas.Rects, atlas.Rects[len(atlas.Rects)-1])
		}
	}
	x, y := 0, len(config.Palettes)*rowHeight
	for _, frame := range config.OriginalFrames {
		region := image.Rect(x, y, x+frame.Dx(), y+frame.Dy())
		copyPaletteFrame(atlas.Pixels, region.Min, config.Source, frame, nil, nil)
		atlas.Rects = append(atlas.Rects, region)
		x += frame.Dx()
	}
	return atlas, nil
}

func copyPaletteFrame(dst *image.RGBA, at image.Point, source image.Image, frame image.Rectangle, colors []color.RGBA, index func(color.Color) int) {
	for y := 0; y < frame.Dy(); y++ {
		for x := 0; x < frame.Dx(); x++ {
			pixel := source.At(frame.Min.X+x, frame.Min.Y+y)
			if index != nil {
				_, _, _, alpha := pixel.RGBA()
				if alpha == 0 {
					continue
				}
				if slot := index(pixel); slot >= 0 && slot < len(colors) {
					pixel = colors[slot]
				}
			}
			dst.Set(at.X+x, at.Y+y, pixel)
		}
	}
}

// Upload returns the owning GPU image and borrowed frame views. The caller
// deallocates the image when the sprites are no longer in use.
func (atlas *PaletteAtlas) Upload() (*ebiten.Image, []*ebiten.Image) {
	if atlas == nil || atlas.Pixels == nil {
		return nil, nil
	}
	image := ebiten.NewImageFromImage(atlas.Pixels)
	frames := make([]*ebiten.Image, len(atlas.Rects))
	for index, rect := range atlas.Rects {
		frames[index] = image.SubImage(rect).(*ebiten.Image)
	}
	return image, frames
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
