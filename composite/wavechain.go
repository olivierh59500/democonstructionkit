package composite

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// WavePass selects a working size and origin independently of its source image.
// Increase Wave.Thickness for coarse block/strip deformation, or use one pixel
// for scanline waves. Row and column passes can be combined in either order.
type WavePass struct {
	Size image.Point
	X, Y float64
	Wave WaveStrips
}
type WaveChain struct {
	passes   []WavePass
	surfaces []*ebiten.Image
}

func NewWaveChain(passes ...WavePass) (*WaveChain, error) {
	if len(passes) == 0 {
		return nil, fmt.Errorf("composite: empty wave chain")
	}
	for _, p := range passes {
		if p.Size.X < 1 || p.Size.Y < 1 {
			return nil, fmt.Errorf("composite: invalid wave pass size")
		}
	}
	c := &WaveChain{passes: append([]WavePass(nil), passes...)}
	for i := range c.passes {
		c.passes[i].Wave.Waves = append([]StripWave(nil), c.passes[i].Wave.Waves...)
		c.surfaces = append(c.surfaces, render.NewSurface(c.passes[i].Size.X, c.passes[i].Size.Y))
	}
	return c, nil
}

// Pass exposes this instance's parameters for live phase/amplitude/band changes.
// Its Size is fixed at construction; source images remain caller-owned.
func (c *WaveChain) Pass(index int) *WaveStrips { return &c.passes[index].Wave }
func (c *WaveChain) Advance() {
	for i := range c.passes {
		c.passes[i].Wave.Advance()
	}
}

// Render returns a borrowed surface, valid until the next Render or Close.
// It never advances phases, so drawing at any display refresh rate is safe.
func (c *WaveChain) Render(src *ebiten.Image) *ebiten.Image {
	if src == nil || len(c.surfaces) == 0 {
		return nil
	}
	for i, p := range c.passes {
		dst := c.surfaces[i]
		dst.Clear()
		p.Wave.DrawAt(dst, src, p.X, p.Y)
		src = dst
	}
	return src
}
func (c *WaveChain) Close() error {
	for _, s := range c.surfaces {
		s.Deallocate()
	}
	c.surfaces = nil
	return nil
}
