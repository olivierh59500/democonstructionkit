package scrolling

import (
	"fmt"
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// PathConfig places each glyph along a distance-based baseline. Configure either
// Path (precomputed coordinates/curve) or Sample (a time-dependent user function).
// Glyph advances remain in pixels, so mixed and proportional fonts keep spacing.
type PathConfig struct {
	Path                           *motion.Path
	Sample                         func(distance, seconds float64) (position, tangent motion.Point)
	Offset, NormalOffset, Rotation float64
	Orient                         bool
	Vertical                       bool            // Read the pen distance from Y for a vertical scroller.
	Clip                           bool            // Hide glyph origins outside an open Path; closed paths always wrap.
	Extrapolate                    bool            // Continue an open path along its endpoint tangents.
	Viewport                       image.Rectangle // Clip rendered pixels, rather than whole glyph origins.
}

func AlongPath(c PathConfig) (Mode, error) {
	if (c.Path == nil) == (c.Sample == nil) || !finite(c.Offset) || !finite(c.NormalOffset) || !finite(c.Rotation) {
		return Mode{}, fmt.Errorf("scrolling: choose one valid path or sampler")
	}
	if c.Extrapolate && (c.Clip || c.Path == nil) {
		return Mode{}, fmt.Errorf("scrolling: endpoint extrapolation requires an unclipped coordinate path")
	}
	mode := Mode{Map: func(s Sample, op *ebiten.DrawImageOptions) bool {
		distance := s.X + c.Offset
		if c.Vertical {
			distance = s.Y + c.Offset
		}
		var p, t motion.Point
		if c.Path != nil {
			if c.Clip && !c.Path.Closed() && (distance < 0 || distance > c.Path.Length()) {
				return false
			}
			p, t = c.Path.At(distance)
			if c.Extrapolate && !c.Path.Closed() {
				extra := 0.0
				if distance < 0 {
					extra = distance
				} else if distance > c.Path.Length() {
					extra = distance - c.Path.Length()
				}
				p.X += t.X * extra
				p.Y += t.Y * extra
			}
		} else {
			p, t = c.Sample(distance, s.Time)
		}
		if !finite(p.X) || !finite(p.Y) || !finite(t.X) || !finite(t.Y) {
			return false
		}
		length := math.Hypot(t.X, t.Y)
		if !finite(length) {
			return false
		}
		if length > 0 {
			t.X /= length
			t.Y /= length
		}
		angle := c.Rotation
		if c.Orient && length > 0 {
			angle += math.Atan2(t.Y, t.X)
		}
		op.GeoM.Translate(-s.X, -s.Y)
		op.GeoM.Rotate(angle)
		op.GeoM.Translate(p.X-t.Y*c.NormalOffset, p.Y+t.X*c.NormalOffset)
		return true
	}}
	if !c.Viewport.Empty() {
		mode.Paint = func(dst *ebiten.Image, sample Sample, options ebiten.DrawImageOptions) {
			if sample.Glyph.Image == nil {
				return
			}
			clip := c.Viewport.Intersect(dst.Bounds())
			if !clip.Empty() && glyphIntersects(clip, sample.Glyph.Image.Bounds(), options.GeoM) {
				dst.SubImage(clip).(*ebiten.Image).DrawImage(sample.Glyph.Image, &options)
			}
		}
	}
	return mode, nil
}
