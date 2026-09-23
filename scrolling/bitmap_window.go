package scrolling

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// DrawWindow draws only glyphs intersecting the width-pixel interval beginning
// at dst.Bounds().Min.X. The destination supplies clipping; a subimage therefore
// selects an interior viewport without changing the text's absolute origin.
// Text keeps its original origin x, without substring allocation or layout work.
func (t *BitmapText) DrawWindow(dst *ebiten.Image, x, y, width float64) error {
	if dst == nil {
		return nil
	}
	if t.windowBatch == nil {
		t.windowBatch = composite.NewQuadBatch(min(256, max(1, t.Len())))
	}
	t.windowBatch.Options.Filter = t.grid.Filter
	t.windowBatch.Options.Address = ebiten.AddressClampToZero
	t.windowBatch.Begin(dst, t.grid.Image)
	err := t.appendWindowWithin(t.windowBatch, x, y, float64(dst.Bounds().Min.X), math.Min(width, float64(dst.Bounds().Dx())), false)
	t.windowBatch.Flush()
	return err
}

func (t *BitmapText) appendWindow(batch *composite.QuadBatch, x, y, width float64, repeat bool) error {
	return t.appendWindowWithin(batch, x, y, 0, width, repeat)
}
func (t *BitmapText) appendWindowWithin(batch *composite.QuadBatch, x, y, left, width float64, repeat bool) error {
	if !finite(x) || !finite(y) || !finite(width) || width < 0 {
		return fmt.Errorf("scrolling: invalid bitmap window")
	}
	if t.Len() == 0 || width == 0 {
		return nil
	}
	first := math.Max(0, math.Floor((left-x-t.grid.Width)/t.advance)+1)
	end := math.Ceil((left + width - x) / t.advance)
	if !repeat {
		end = math.Min(end, float64(t.Len()))
	}
	if end <= first {
		return nil
	}
	if !finite(first) || !finite(end) || end-first > 65536 || first > min(float64(1<<52), float64(int(^uint(0)>>1))) || end > min(float64(1<<52), float64(int(^uint(0)>>1))) {
		return fmt.Errorf("scrolling: bitmap window exceeds glyph budget")
	}
	// Rebase around the first visible pen to keep geometry precise and retain
	// the same operation order as printing a substring at an adjusted origin.
	origin := x + first*t.advance
	for i := int(first); i < int(end); i++ {
		index := i
		if repeat {
			index %= t.Len()
		}
		if !t.valid[index] {
			continue
		}
		r := t.regions[index]
		left := origin + float64(i-int(first))*t.advance
		vertices := [4]ebiten.Vertex{
			{DstX: float32(left), DstY: float32(y), SrcX: float32(r.X), SrcY: float32(r.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(left + r.Width), DstY: float32(y), SrcX: float32(r.X + r.Width), SrcY: float32(r.Y), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(left + r.Width), DstY: float32(y + r.Height), SrcX: float32(r.X + r.Width), SrcY: float32(r.Y + r.Height), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: float32(left), DstY: float32(y + r.Height), SrcX: float32(r.X), SrcY: float32(r.Y + r.Height), ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		}
		batch.Quad(vertices)
	}
	return nil
}
