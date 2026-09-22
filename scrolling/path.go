package scrolling

import (
	"fmt"
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
	Vertical                       bool // Read the pen distance from Y for a vertical scroller.
	Clip                           bool // Hide glyph origins outside an open Path; closed paths always wrap.
}

func AlongPath(c PathConfig) (Mode, error) {
	if (c.Path == nil) == (c.Sample == nil) || !finite(c.Offset) || !finite(c.NormalOffset) || !finite(c.Rotation) {
		return Mode{}, fmt.Errorf("scrolling: choose one valid path or sampler")
	}
	return Mode{Map: func(s Sample, op *ebiten.DrawImageOptions) bool {
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
	}}, nil
}
