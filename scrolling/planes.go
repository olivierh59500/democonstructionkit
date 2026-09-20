package scrolling

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
)

// PlaneForm describes independent vertical and depth oscillations. Spatial
// steps are radians per character, speeds are radians per phase unit. A zero
// depth amplitude gives a flat scroll; a zero vertical step gives a bounce.
// Vertical uses cosine, matching the TCB production's phase convention.
type PlaneForm struct {
	DepthAmplitude, DepthStep, DepthSpeed, DepthPhase  float64
	Height, VerticalStep, VerticalSpeed, VerticalPhase float64
}

// PlaneSlot is one position in a scroll. Form=-1 keeps the active form. A slot
// may be invisible (Rune=0), preserving control-byte spacing when required by
// an original demo. Advance is explicit and may come from any font's metrics.
type PlaneSlot struct {
	Rune    rune
	Advance float64
	Form    int
}

type PlanePoint struct {
	X, Y, Scale float64
	Rune        rune
	Index       int
}

type PlaneProjection struct {
	Focal, Depth              float64
	OriginX, CenterX, CenterY float64
	XBias, YBias              float64
	VerticalOffset            float64
}

type PlanesConfig struct {
	Slots      []PlaneSlot
	Forms      []PlaneForm
	Projection PlaneProjection
	Visible    int
	PhaseStep  float64
}

// Planes keeps the original TCB recurrence, visible-slot control timing and
// stable depth order. It is independent of textures, alphabets and cell sizes.
// Use the normal Scrolling Modes API for controls at the text cursor instead.
type Planes struct {
	config        PlanesConfig
	points        []PlanePoint
	first, form   int
	offset, phase float64
	length        float64
}

func NewPlanes(c PlanesConfig) (*Planes, error) {
	if len(c.Slots) == 0 || len(c.Forms) == 0 || c.Visible < 1 || c.Visible > 16383 || !finite(c.Projection.Focal) || c.Projection.Focal <= 0 || !finite(c.PhaseStep) {
		return nil, fmt.Errorf("scrolling: invalid multi-plane configuration")
	}
	for _, v := range []float64{c.Projection.Depth, c.Projection.OriginX, c.Projection.CenterX, c.Projection.CenterY, c.Projection.XBias, c.Projection.YBias, c.Projection.VerticalOffset} {
		if !finite(v) {
			return nil, fmt.Errorf("scrolling: nonfinite plane projection")
		}
	}
	for _, f := range c.Forms {
		for _, v := range []float64{f.DepthAmplitude, f.DepthStep, f.DepthSpeed, f.DepthPhase, f.Height, f.VerticalStep, f.VerticalSpeed, f.VerticalPhase} {
			if !finite(v) {
				return nil, fmt.Errorf("scrolling: nonfinite plane form")
			}
		}
	}
	p := &Planes{config: c, points: make([]PlanePoint, c.Visible)}
	for _, s := range c.Slots {
		if !finite(s.Advance) || s.Advance <= 0 || s.Form < -1 || s.Form >= len(c.Forms) {
			return nil, fmt.Errorf("scrolling: invalid multi-plane slot")
		}
		p.length += s.Advance
	}
	if !finite(p.length) {
		return nil, fmt.Errorf("scrolling: multi-plane text is too long")
	}
	p.config.Slots = append([]PlaneSlot(nil), c.Slots...)
	p.config.Forms = append([]PlaneForm(nil), c.Forms...)
	return p, nil
}

// Points borrows the last computed, depth-sorted positions. Do not modify it.
func (p *Planes) Points() []PlanePoint { return p.points }

// Step samples first, then advances the pen, preserving original frame timing.
func (p *Planes) Step(pixels float64) error {
	if !finite(pixels) || pixels < 0 {
		return fmt.Errorf("scrolling: invalid multi-plane step")
	}
	p.phase += p.config.PhaseStep
	active, previous := -1, -2
	var zs, zc, ys, yc, zss, zcs, yss, ycs float64
	x := p.config.Projection.OriginX - p.offset
	c := p.config.Projection
	for i := range p.points {
		index := (p.first + i) % len(p.config.Slots)
		slot := p.config.Slots[index]
		if slot.Form >= 0 {
			p.form = slot.Form
		}
		f := p.config.Forms[p.form]
		if active != p.form || index != previous+1 {
			if f.DepthAmplitude != 0 {
				zs, zc = math.Sincos(f.DepthPhase + float64(index)*f.DepthStep + p.phase*f.DepthSpeed)
				zss, zcs = math.Sincos(f.DepthStep)
			}
			ys, yc = math.Sincos(f.VerticalPhase + float64(index)*f.VerticalStep + p.phase*f.VerticalSpeed)
			yss, ycs = math.Sincos(f.VerticalStep)
			active = p.form
		} else {
			if f.DepthAmplitude != 0 {
				zs, zc = zs*zcs+zc*zss, zc*zcs-zs*zss
			}
			ys, yc = ys*ycs+yc*yss, yc*ycs-ys*yss
		}
		previous = index
		z := f.DepthAmplitude*zs + c.Depth
		y := f.Height*yc + c.VerticalOffset
		scale := c.Focal / (c.Focal + z)
		if c.Focal+z <= 0 {
			scale = 0
		}
		p.points[i] = PlanePoint{X: (x+c.XBias)*scale + c.CenterX, Y: (y+c.YBias)*scale + c.CenterY, Scale: scale, Rune: slot.Rune, Index: index}
		x += slot.Advance
	}
	// Stable insertion sort retains overlapping glyph order at equal depth.
	for i := 1; i < len(p.points); i++ {
		item := p.points[i]
		j := i
		for j > 0 && p.points[j-1].Scale > item.Scale {
			p.points[j] = p.points[j-1]
			j--
		}
		p.points[j] = item
	}
	p.offset += math.Mod(pixels, p.length)
	for p.offset >= p.config.Slots[p.first].Advance {
		p.offset -= p.config.Slots[p.first].Advance
		p.first = (p.first + 1) % len(p.config.Slots)
	}
	return nil
}

// Mode also exposes an individual TCB-style form through the regular Scrolling
// pipeline. Geometry uses the selected face's layout; no 32x33 assumption exists.
func (f PlaneForm) Mode(c PlaneProjection) Mode {
	depth := func(s Sample) float64 {
		return c.Depth + f.DepthAmplitude*math.Sin(f.DepthPhase+float64(s.Index)*f.DepthStep+s.Time*f.DepthSpeed)
	}
	return Mode{Depth: depth, Map: func(s Sample, op *ebiten.DrawImageOptions) bool {
		z := c.Focal + depth(s)
		if c.Focal <= 0 || z <= 0 || !finite(z) {
			return false
		}
		y := f.Height*math.Cos(f.VerticalPhase+float64(s.Index)*f.VerticalStep+s.Time*f.VerticalSpeed) + c.VerticalOffset + c.YBias
		op.GeoM.Translate(-c.CenterX+c.XBias, -c.CenterY+y)
		op.GeoM.Scale(c.Focal/z, c.Focal/z)
		op.GeoM.Translate(c.CenterX, c.CenterY)
		return true
	}}
}
