package composite

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// ContourBankConfig describes retained image slots, each with independent
// layers. Image storage is fixed at construction. FillRule and Blend apply to
// geometry written into a layer, not to later composition of the image.
type ContourBankConfig struct {
	Width, Height  int
	Slots, Layers  int
	BatchTriangles int
	FillRule       ebiten.FillRule
	Blend          ebiten.Blend
}

// ContourBank owns a fixed set of retained mask images, a white texture and a
// reusable geometry batch. A caller's authored controller chooses write/display
// indices and whether a write clears or accumulates its previous image. Fan,
// outline and parity-edge geometry share the same render.Batch API. Borrowed
// Image handles can feed a palette, raster, reflection or another image pass.
type ContourBank struct {
	slots, layers int
	images        []*ebiten.Image
	white         *ebiten.Image
	batch         *render.Batch
}

func NewContourBank(c ContourBankConfig) (*ContourBank, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Width > 8192 || c.Height > 8192 || c.Slots < 1 || c.Slots > 64 || c.Layers < 1 || c.Layers > 8 ||
		int64(c.Width)*int64(c.Height)*int64(c.Slots)*int64(c.Layers) > 16*1024*1024 {
		return nil, fmt.Errorf("composite: invalid contour bank dimensions or image budget")
	}
	if c.BatchTriangles == 0 {
		c.BatchTriangles = 4096
	}
	if c.BatchTriangles < 2 || c.BatchTriangles > 20000 ||
		c.FillRule != ebiten.FillRuleFillAll && c.FillRule != ebiten.FillRuleNonZero && c.FillRule != ebiten.FillRuleEvenOdd {
		return nil, fmt.Errorf("composite: invalid contour triangle budget or fill rule")
	}
	b := &ContourBank{slots: c.Slots, layers: c.Layers, white: ebiten.NewImage(1, 1), batch: render.NewBatch(c.BatchTriangles)}
	b.white.Fill(color.White)
	b.batch.Options.FillRule, b.batch.Options.Blend = c.FillRule, c.Blend
	for i := 0; i < c.Slots*c.Layers; i++ {
		b.images = append(b.images, render.NewSurface(c.Width, c.Height))
	}
	return b, nil
}

// Image borrows a retained slot/layer. Invalid indices or a closed bank return
// nil. Keeping the same handle across writes avoids per-frame image wrappers.
func (b *ContourBank) Image(slot, layer int) *ebiten.Image {
	if b == nil || slot < 0 || slot >= b.slots || layer < 0 || layer >= b.layers || len(b.images) == 0 {
		return nil
	}
	return b.images[slot*b.layers+layer]
}

// Paint optionally clears one layer, then synchronously submits the callback's
// contours. A nil callback can clear a layer without drawing another pose.
// The bank flushes after the callback; the caller does not need to manage GPU
// batch lifetime. Calls and callbacks belong on the graphics goroutine.
func (b *ContourBank) Paint(slot, layer int, clear bool, draw func(*render.Batch)) error {
	dst := b.Image(slot, layer)
	if dst == nil {
		return fmt.Errorf("composite: invalid or closed contour slot/layer")
	}
	if clear {
		dst.Clear()
	}
	if draw != nil {
		b.batch.Begin(dst, b.white)
		draw(b.batch)
		b.batch.Flush()
	}
	return nil
}

// Clear resets every layer. ClearSlot resets a working pose's layers together.
func (b *ContourBank) Clear() {
	if b != nil {
		for _, img := range b.images {
			img.Clear()
		}
	}
}
func (b *ContourBank) ClearSlot(slot int) error {
	if b.Image(slot, 0) == nil {
		return fmt.Errorf("composite: invalid or closed contour slot")
	}
	for layer := 0; layer < b.layers; layer++ {
		b.Image(slot, layer).Clear()
	}
	return nil
}

func (b *ContourBank) Close() error {
	if b != nil {
		for _, img := range b.images {
			img.Deallocate()
		}
		if b.white != nil {
			b.white.Deallocate()
			b.white = nil
		}
		b.images, b.batch = nil, nil
	}
	return nil
}
