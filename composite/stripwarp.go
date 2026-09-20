package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/render"
)

// StripWarpConfig describes a row-sampling pass followed by a column-displacement
// pass. Share a configuration between text, logos or sprites; each StripWarp owns
// its own intermediate image. Callbacks run in increasing strip-index order.
type StripWarpConfig struct {
	RowHeight, ColumnWidth int
	// SampleX returns a source X offset relative to the input image's bounds.
	// Nil samples from X=0. Source crops are clipped before they are drawn.
	SampleX func(row int, frame kit.Frame) int
	// OffsetY returns a column's vertical displacement in pixels. Nil means zero.
	OffsetY func(column int, frame kit.Frame) float64
	// VerticalBias is added before OffsetY, preserving original arithmetic order.
	VerticalBias float64
}

// StripWarp applies the same deformation to any image without assuming font size.
// The intermediate size is the visible width and complete content height. Supply
// a wider source image with horizontal padding when SampleX requires overscan.
// DrawAt must run on the graphics goroutine. Source and destination stay borrowed.
type StripWarp struct {
	config  StripWarpConfig
	size    image.Point
	surface *ebiten.Image
	columns []*ebiten.Image
}

func NewStripWarp(size image.Point, config StripWarpConfig) (*StripWarp, error) {
	if size.X <= 0 || size.Y <= 0 || config.RowHeight <= 0 || config.ColumnWidth <= 0 ||
		math.IsNaN(config.VerticalBias) || math.IsInf(config.VerticalBias, 0) {
		return nil, fmt.Errorf("composite: invalid strip warp dimensions or bias")
	}
	w := &StripWarp{config: config, size: size, surface: render.NewSurface(size.X, size.Y)}
	for x := 0; x < size.X; x += config.ColumnWidth {
		r := image.Rect(x, 0, min(x+config.ColumnWidth, size.X), size.Y)
		w.columns = append(w.columns, w.surface.SubImage(r).(*ebiten.Image))
	}
	return w, nil
}

// Size returns the unwarped visible dimensions, excluding displacement margins.
func (w *StripWarp) Size() image.Point { return w.size }

// DrawAt draws into dst at a base origin. Height is never rounded down to a full
// strip: the last partial row and column retain their actual source dimensions.
// Drawing does not advance the frame or change the supplied source image.
func (w *StripWarp) DrawAt(dst, source *ebiten.Image, frame kit.Frame, x, y float64) {
	if w.surface == nil || source == nil || dst == nil {
		return
	}
	w.surface.Clear()
	bounds := source.Bounds()
	for row, top := 0, 0; top < w.size.Y; row, top = row+1, top+w.config.RowHeight {
		offset := 0
		if w.config.SampleX != nil {
			offset = w.config.SampleX(row, frame)
		}
		r := image.Rect(offset, top, offset+w.size.X, min(top+w.config.RowHeight, w.size.Y)).Add(bounds.Min)
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(0, float64(top))
		Instance{Image: source, Source: &r, Options: op}.Draw(w.surface)
	}
	for column, img := range w.columns {
		offset := 0.0
		if w.config.OffsetY != nil {
			offset = w.config.OffsetY(column, frame)
		}
		op := ebiten.DrawImageOptions{}
		op.GeoM.Translate(x+float64(column*w.config.ColumnWidth), y+w.config.VerticalBias+offset)
		dst.DrawImage(img, &op)
	}
}

// Close releases only the internal surface and is safe to call repeatedly.
func (w *StripWarp) Close() error {
	if w.surface != nil {
		w.surface.Deallocate()
		w.surface = nil
		w.columns = nil
	}
	return nil
}
