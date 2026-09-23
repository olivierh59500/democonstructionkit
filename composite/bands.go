package composite

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// MovingBand is one independently moving crop of a shared background atlas.
// Velocity and Wrap use phase units per Step; MotionScale converts phase to
// destination pixels without changing image scale. Signed remainder preserves
// authored leftward wraps. CopyOffsets permits overlaps or deliberately finite
// repetitions; an empty list draws one copy.
type MovingBand struct {
	Source                                             image.Rectangle
	X, Y, Width, Height                                float64
	PhaseX, PhaseY, VelocityX, VelocityY, WrapX, WrapY float64
	MotionScaleX, MotionScaleY                         float64
}
type BandsConfig struct {
	Bands       []MovingBand
	CopyOffsets [][2]float64
	Filter      ebiten.Filter
	Blend       ebiten.Blend
}
type Bands struct {
	config BandsConfig
	batch  *QuadBatch
}

func NewBands(c BandsConfig) (*Bands, error) {
	if len(c.Bands) == 0 || len(c.Bands) > 16383 || len(c.CopyOffsets) > 16383 || len(c.Bands)*max(1, len(c.CopyOffsets)) > 65536 {
		return nil, fmt.Errorf("composite: invalid band count")
	}
	c.Bands = append([]MovingBand(nil), c.Bands...)
	c.CopyOffsets = append([][2]float64(nil), c.CopyOffsets...)
	if len(c.CopyOffsets) == 0 {
		c.CopyOffsets = [][2]float64{{0, 0}}
	}
	for _, offset := range c.CopyOffsets {
		if !bandFinite(offset[0]) || !bandFinite(offset[1]) {
			return nil, fmt.Errorf("composite: invalid band copy offset")
		}
	}
	for i := range c.Bands {
		b := &c.Bands[i]
		if b.Width == 0 {
			b.Width = float64(b.Source.Dx())
		}
		if b.Height == 0 {
			b.Height = float64(b.Source.Dy())
		}
		if b.MotionScaleX == 0 {
			b.MotionScaleX = 1
		}
		if b.MotionScaleY == 0 {
			b.MotionScaleY = 1
		}
		if b.Source.Empty() || b.Width <= 0 || b.Height <= 0 || b.WrapX < 0 || b.WrapY < 0 {
			return nil, fmt.Errorf("composite: invalid moving band")
		}
		for _, v := range []float64{b.X, b.Y, b.Width, b.Height, b.PhaseX, b.PhaseY, b.VelocityX, b.VelocityY, b.WrapX, b.WrapY, b.MotionScaleX, b.MotionScaleY} {
			if !bandFinite(v) {
				return nil, fmt.Errorf("composite: invalid moving band coordinate")
			}
		}
	}
	batch := NewQuadBatch(min(256, len(c.Bands)*len(c.CopyOffsets)))
	batch.Options.Filter = c.Filter
	batch.Options.Blend = c.Blend
	return &Bands{config: c, batch: batch}, nil
}
func bandFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

// Step advances independent positions once. Draw never advances the simulation.
func (b *Bands) Step() {
	for i := range b.config.Bands {
		p := &b.config.Bands[i]
		p.PhaseX += p.VelocityX
		p.PhaseY += p.VelocityY
		if p.WrapX > 0 {
			p.PhaseX = math.Mod(p.PhaseX, p.WrapX)
		}
		if p.WrapY > 0 {
			p.PhaseY = math.Mod(p.PhaseY, p.WrapY)
		}
	}
}

// DrawAt batches all bands using bounded reusable geometry and no image copies.
func (b *Bands) DrawAt(dst, atlas *ebiten.Image, x, y float64) {
	if b == nil || dst == nil || atlas == nil {
		return
	}
	b.batch.Begin(dst, atlas)
	for _, p := range b.config.Bands {
		for _, offset := range b.config.CopyOffsets {
			b.batch.Rect(p.Source, float32(x+p.X+p.PhaseX*p.MotionScaleX+offset[0]), float32(y+p.Y+p.PhaseY*p.MotionScaleY+offset[1]), float32(p.Width), float32(p.Height))
		}
	}
	b.batch.Flush()
}

// Band returns a value snapshot suitable for controls or an effect inspector.
func (b *Bands) Band(index int) (MovingBand, bool) {
	if b == nil || index < 0 || index >= len(b.config.Bands) {
		return MovingBand{}, false
	}
	return b.config.Bands[index], true
}

// SetVelocity changes motion without resetting current phases. Call from the
// update thread, for example in response to an input or music envelope.
func (b *Bands) SetVelocity(index int, x, y float64) error {
	if b == nil || index < 0 || index >= len(b.config.Bands) || !bandFinite(x) || !bandFinite(y) {
		return fmt.Errorf("composite: invalid band velocity")
	}
	b.config.Bands[index].VelocityX, b.config.Bands[index].VelocityY = x, y
	return nil
}

// SetPhase seeks one band independently; its next Step applies configured wrap.
func (b *Bands) SetPhase(index int, x, y float64) error {
	if b == nil || index < 0 || index >= len(b.config.Bands) || !bandFinite(x) || !bandFinite(y) {
		return fmt.Errorf("composite: invalid band phase")
	}
	b.config.Bands[index].PhaseX, b.config.Bands[index].PhaseY = x, y
	return nil
}
