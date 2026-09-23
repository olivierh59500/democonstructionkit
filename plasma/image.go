package plasma

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// HarmonicImage owns the CPU pixels and GPU surface for one independently
// animated harmonic field. A dirty frame is rendered and uploaded only once,
// even when several layers sample Image during the same simulation tick.
type HarmonicImage struct {
	kernel *Harmonic
	image  *ebiten.Image
	pixels []byte
	width  int
	step   float64
	time   float64
	dirty  bool
	err    error
}

func NewHarmonicImage(config HarmonicConfig, step float64) (*HarmonicImage, error) {
	if math.IsNaN(step) || math.IsInf(step, 0) {
		return nil, fmt.Errorf("plasma: invalid harmonic-image time step")
	}
	kernel, err := NewHarmonic(config)
	if err != nil {
		return nil, err
	}
	return &HarmonicImage{
		kernel: kernel, image: ebiten.NewImageWithOptions(
			image.Rect(0, 0, config.Width, config.Height), &ebiten.NewImageOptions{Unmanaged: true}),
		pixels: make([]byte, config.Width*config.Height*4), width: config.Width,
		step: step, dirty: true,
	}, nil
}

// Update advances time once. Draw and Image never change the animation clock.
func (p *HarmonicImage) Update(kit.Frame) error {
	if p == nil || p.image == nil {
		return nil
	}
	next := p.time + p.step
	if math.IsNaN(next) || math.IsInf(next, 0) {
		return fmt.Errorf("plasma: harmonic-image time overflows")
	}
	p.time, p.dirty = next, true
	return p.err
}

// SetTime supports seeking and synchronizing several independent images.
func (p *HarmonicImage) SetTime(seconds float64) error {
	if math.IsNaN(seconds) || math.IsInf(seconds, 0) {
		return fmt.Errorf("plasma: invalid harmonic-image time")
	}
	p.time, p.dirty = seconds, true
	return nil
}

func (p *HarmonicImage) Time() float64 { return p.time }
func (p *HarmonicImage) Err() error    { return p.err }

// Image returns the borrowed, live surface. It uploads only when time changed.
func (p *HarmonicImage) Image() *ebiten.Image {
	if p == nil || p.image == nil || p.err != nil {
		return nil
	}
	if p.dirty {
		if err := p.kernel.RenderRGBA(p.pixels, p.width*4, p.time); err != nil {
			p.err = err
			return nil
		}
		p.image.WritePixels(p.pixels)
		p.dirty = false
	}
	return p.image
}

func (p *HarmonicImage) Draw(dst *ebiten.Image) {
	if dst != nil {
		if source := p.Image(); source != nil {
			dst.DrawImage(source, nil)
		}
	}
}

func (p *HarmonicImage) Close() error {
	if p != nil && p.image != nil {
		p.image.Deallocate()
		p.image = nil
		p.pixels = nil
	}
	return nil
}
