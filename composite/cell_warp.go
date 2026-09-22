package composite

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// CellTransform transforms one rigid image fragment in local cell coordinates.
// The zero value is identity. DrawAt adds the cell's grid position and the draw
// origin after GeoM, so a rotation or scale does not move the rest of the grid.
type CellTransform struct {
	GeoM       ebiten.GeoM
	ColorScale ebiten.ColorScale
	Hidden     bool
}

// CellWarpConfig splits any image into independent rectangular fragments.
// Source is an absolute source crop; an empty rectangle selects the whole image.
// Gap adds destination spacing between cells and can be negative for overlap.
// Sample runs in row-major order, with zero-based indices relative to Source.
// Each source cell retains its identity even when neighboring rows cross.
type CellWarpConfig struct {
	Cell, Gap image.Point
	Source    image.Rectangle
	Filter    ebiten.Filter
	Blend     ebiten.Blend
	Sample    func(row, column int, frame kit.Frame) CellTransform
}

// CellWarp deforms images without stretching a mesh across cell boundaries.
// Crops are cached for the most recent source image. Drawing that source again
// reuses them, including partial edge cells, without allocating image wrappers.
// Subimage sampling preserves filter clipping at each fragment's original edge.
// Images stay caller-owned. Use on the graphics goroutine, with distinct source
// and destination images. DrawAt never advances animation or reads image pixels.
type CellWarp struct {
	config CellWarpConfig
	source *ebiten.Image
	cells  []cellFragment
}

type cellFragment struct {
	image       *ebiten.Image
	row, column int
	left, top   int
}

func NewCellWarp(c CellWarpConfig) (*CellWarp, error) {
	if c.Cell.X <= 0 || c.Cell.Y <= 0 || c.Gap.X <= -c.Cell.X || c.Gap.Y <= -c.Cell.Y {
		return nil, fmt.Errorf("composite: invalid cell dimensions or gap")
	}
	return &CellWarp{config: c}, nil
}

// DrawAt places the cropped image's top-left cell at x,y before its transform.
// Cropped source padding does not become destination padding; add it to x,y when
// wanted. Changing source reuses fragment storage but refreshes borrowed crops.
func (w *CellWarp) DrawAt(dst, source *ebiten.Image, frame kit.Frame, x, y float64) {
	if w == nil || dst == nil || source == nil {
		return
	}
	if source != w.source {
		w.prepare(source)
	}
	for _, cell := range w.cells {
		pose := CellTransform{}
		if w.config.Sample != nil {
			pose = w.config.Sample(cell.row, cell.column, frame)
		}
		if pose.Hidden {
			continue
		}
		op := ebiten.DrawImageOptions{GeoM: pose.GeoM, ColorScale: pose.ColorScale, Filter: w.config.Filter, Blend: w.config.Blend}
		op.GeoM.Translate(x+float64(cell.left), y+float64(cell.top))
		dst.DrawImage(cell.image, &op)
	}
}

func (w *CellWarp) prepare(source *ebiten.Image) {
	// Clear references to a previous atlas when the next image has fewer cells.
	clear(w.cells)
	w.cells = w.cells[:0]
	w.source = source
	bounds := source.Bounds()
	if !w.config.Source.Empty() {
		bounds = bounds.Intersect(w.config.Source)
	}
	if bounds.Empty() {
		return
	}
	cell := w.config.Cell
	for row, top := 0, bounds.Min.Y; top < bounds.Max.Y; row, top = row+1, top+cell.Y {
		for column, left := 0, bounds.Min.X; left < bounds.Max.X; column, left = column+1, left+cell.X {
			crop := image.Rect(left, top, min(left+cell.X, bounds.Max.X), min(top+cell.Y, bounds.Max.Y))
			w.cells = append(w.cells, cellFragment{
				image: source.SubImage(crop).(*ebiten.Image), row: row, column: column,
				left: column * (cell.X + w.config.Gap.X), top: row * (cell.Y + w.config.Gap.Y),
			})
		}
	}
}

// Close drops cached references without deallocating caller-owned source images.
// It is safe to call repeatedly. DrawAt can prepare a new cache after Close.
func (w *CellWarp) Close() error {
	if w != nil {
		w.source = nil
		w.cells = nil
	}
	return nil
}
