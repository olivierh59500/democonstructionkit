package scrolling

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

type Painter func(*ebiten.Image, Sample, ebiten.DrawImageOptions)

// Mode combines an independent geometry transform, optional fragment renderer
// and optional depth ordering. Larger Depth values are painted first.
type Mode struct {
	Prepare func([]Glyph) error
	Map     Mapper
	Paint   Painter
	Depth   func(Sample) float64
}

// Chain applies transforms in the requested order, allowing a bouncing, zooming
// sine scroll, for example. A hidden glyph stops the remainder of the chain.
func Chain(mappers ...Mapper) Mapper {
	return func(s Sample, op *ebiten.DrawImageOptions) bool {
		for _, m := range mappers {
			if m != nil && !m(s, op) {
				return false
			}
		}
		return true
	}
}

func Normal() Mode { return Mode{} }

// Sine displaces glyphs vertically. Spatial is in radians per text pixel, so
// changing font width or using proportional/mixed faces does not change units.
func Sine(w motion.Wave) Mode {
	return Mode{Map: func(s Sample, op *ebiten.DrawImageOptions) bool {
		op.GeoM.Translate(0, w.At(s.Glyph.Offset, s.Time))
		return true
	}}
}

// Bounce moves the complete baseline; Amplitude is in pixels and Speed in
// radians per second. Phase allows independent scrollers to share a clock.
func Bounce(w motion.Wave) Mode { w.Spatial = 0; return Sine(w) }

type ZoomConfig struct {
	BaseX, BaseY   float64
	Wave           motion.Wave
	PivotX, PivotY float64
}

// Zoom scales the whole text around an explicit screen-space pivot. Set both
// bases to one for a unit-size rest state. Negative scale mirrors the text.
func Zoom(c ZoomConfig) Mode {
	return Mode{Map: func(s Sample, op *ebiten.DrawImageOptions) bool {
		v := c.Wave.At(s.Glyph.Offset, s.Time)
		op.GeoM.Translate(-c.PivotX, -c.PivotY)
		op.GeoM.Scale(c.BaseX+v, c.BaseY+v)
		op.GeoM.Translate(c.PivotX, c.PivotY)
		return true
	}}
}

type PerspectiveConfig struct {
	Focal, Near, Depth      float64
	CenterX, CenterY        float64
	DepthWave, VerticalWave motion.Wave
}

// Perspective projects the complete glyph around a camera center, then paints
// distant glyphs first. Near is a positive clipping distance from the camera.
// Its waves are expressed in text pixels, independent of an atlas or alphabet.
func Perspective(c PerspectiveConfig) (Mode, error) {
	if !finite(c.Focal) || !finite(c.Near) || c.Focal <= 0 || c.Near <= 0 {
		return Mode{}, fmt.Errorf("scrolling: invalid perspective camera")
	}
	depth := func(s Sample) float64 { return c.Depth + c.DepthWave.At(s.Glyph.Offset, s.Time) }
	return Mode{Depth: depth, Map: func(s Sample, op *ebiten.DrawImageOptions) bool {
		z := c.Focal + depth(s)
		if !finite(z) || z < c.Near {
			return false
		}
		op.GeoM.Translate(-c.CenterX, -c.CenterY+c.VerticalWave.At(s.Glyph.Offset, s.Time))
		op.GeoM.Scale(c.Focal/z, c.Focal/z)
		op.GeoM.Translate(c.CenterX, c.CenterY)
		return true
	}}, nil
}

type Cue struct {
	At   float64
	Mode string
}

// ModeSequence is an absolute-time alternative to {shape:name} controls.
// Times are in seconds. A zero period holds the last cue indefinitely;
// a positive period repeats the sequence without resetting effect phase.
type ModeSequence struct {
	cues   []Cue
	period float64
}

func NewModeSequence(cues []Cue, period float64) (*ModeSequence, error) {
	if len(cues) == 0 || cues[0].At != 0 || !finite(period) || period < 0 {
		return nil, fmt.Errorf("scrolling: a mode sequence must start at zero")
	}
	for i, c := range cues {
		if !finite(c.At) || c.Mode == "" || (i > 0 && c.At <= cues[i-1].At) {
			return nil, fmt.Errorf("scrolling: invalid mode cue")
		}
	}
	if period > 0 && period <= cues[len(cues)-1].At {
		return nil, fmt.Errorf("scrolling: period must exceed the last cue")
	}
	return &ModeSequence{append([]Cue(nil), cues...), period}, nil
}

func (s *ModeSequence) At(seconds float64) string {
	if s == nil {
		return ""
	}
	if !finite(seconds) {
		seconds = 0
	}
	seconds = math.Max(0, seconds)
	if s.period > 0 {
		seconds = motion.Wrap(seconds, s.period)
	}
	name := s.cues[0].Mode
	for _, c := range s.cues {
		if c.At > seconds {
			break
		}
		name = c.Mode
	}
	return name
}
