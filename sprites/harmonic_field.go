package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// HarmonicFieldConfig supplies a fixed sprite population and a sampled motion
// program. ClockScale multiplies Frame.Time; an all-zero value defaults to
// {1, 0}. PixelSnap rounds each final position, without changing the oscillators.
// Artwork in Style is borrowed and may be replaced independently of motion.
type HarmonicFieldConfig struct {
	Count      int
	Motion     motion.HarmonicFormationConfig
	ClockScale [2]float64
	PixelSnap  bool
	Style      FieldStyle
}

// HarmonicField owns its motion, cached samples and batched sprite renderer.
// Sampling uses absolute clocks rather than accumulated phase, retaining exact
// authored poses after a seek. Draw does not resample or advance the formation.
type HarmonicField struct {
	Style      FieldStyle
	formation  *motion.HarmonicFormation
	clockScale [2]float64
	pixelSnap  bool
	samples    []FieldSample
	renderer   *FieldRenderer
}

func NewHarmonicField(c HarmonicFieldConfig) (*HarmonicField, error) {
	if c.Count < 0 || c.Count > 65536 || !finiteField(c.ClockScale[0]) || !finiteField(c.ClockScale[1]) {
		return nil, fmt.Errorf("sprites: invalid harmonic field count or clock scale")
	}
	formation, err := motion.NewHarmonicFormation(c.Motion)
	if err != nil {
		return nil, err
	}
	if formation.IndexOffsetCount() > 0 && formation.IndexOffsetCount() < c.Count {
		return nil, fmt.Errorf("sprites: harmonic field phases are shorter than population")
	}
	if c.ClockScale == [2]float64{} {
		c.ClockScale[0] = 1
	}
	f := &HarmonicField{Style: c.Style, formation: formation, clockScale: c.ClockScale,
		pixelSnap: c.PixelSnap, samples: make([]FieldSample, c.Count), renderer: NewFieldRenderer(c.Count)}
	if err := f.Sample([2]float64{}, 1); err != nil {
		f.Close()
		return nil, err
	}
	return f, nil
}

func (f *HarmonicField) Update(frame kit.Frame) error {
	if f == nil || f.renderer == nil || !finiteField(frame.Time) {
		return fmt.Errorf("sprites: invalid harmonic field update")
	}
	return f.Sample([2]float64{frame.Time * f.clockScale[0], frame.Time * f.clockScale[1]}, 1)
}

// Sample prepares explicit clock/envelope values for cues or music signals.
// Samples expose the same retained positions for an additional visual skin.
func (f *HarmonicField) Sample(clocks [2]float64, envelope float64) error {
	if f == nil || f.renderer == nil || !finiteField(clocks[0]) || !finiteField(clocks[1]) || !finiteField(envelope) {
		return fmt.Errorf("sprites: invalid harmonic field clocks")
	}
	for i := range f.samples {
		p := f.formation.At(i, clocks, envelope)
		if !finiteField(p.X) || !finiteField(p.Y) {
			return fmt.Errorf("sprites: nonfinite harmonic field pose")
		}
		if f.pixelSnap {
			p.X, p.Y = math.Round(p.X), math.Round(p.Y)
		}
		f.samples[i] = FieldSample{Index: i, X: p.X, Y: p.Y, Scale: 1}
	}
	return nil
}

func (f *HarmonicField) Samples() []FieldSample { return f.samples }
func (f *HarmonicField) Draw(dst *ebiten.Image) {
	if f != nil {
		f.DrawStyle(dst, f.Style)
	}
}

// DrawStyle draws the same cached positions with another borrowed material,
// including box outlines or solid pixels. It does not change the default Style.
func (f *HarmonicField) DrawStyle(dst *ebiten.Image, style FieldStyle) {
	if f != nil && f.renderer != nil {
		f.renderer.Draw(dst, f.samples, style)
	}
}

// Close releases the renderer's fallback pixel; caller-owned images survive.
func (f *HarmonicField) Close() error {
	if f != nil && f.renderer != nil {
		f.renderer.Close()
		f.renderer = nil
	}
	return nil
}

var _ kit.Effect = (*HarmonicField)(nil)
