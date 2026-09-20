package outline

import (
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
	"math"
)

// Point32 and Line32 retain the arithmetic precision of classic software renderers.
type Point32 struct{ X, Y float32 }
type Line32 struct{ A, B Point32 }

// FlattenConfig preserves original font-unit scaling and adaptive curve sampling.
type FlattenConfig struct {
	Scale, CurveStep, CloseEpsilon float32
	MaxSegments                    int
}

// Flatten32 closes contours and uses the explicit Bernstein formulas in float32.
// Curve subdivision and duplicate-point tolerance are data, not fixed font rules.
func Flatten32(segments sfnt.Segments, c FlattenConfig) []Line32 {
	if c.CurveStep <= 0 || c.MaxSegments < 1 {
		return nil
	}
	point := func(p fixed.Point26_6) Point32 { return Point32{float32(p.X) * c.Scale, float32(p.Y) * c.Scale} }
	var lines []Line32
	var start, previous Point32
	contour := false
	add := func(a, b Point32) {
		dx, dy := a.X-b.X, a.Y-b.Y
		if dx*dx+dy*dy <= c.CloseEpsilon*c.CloseEpsilon {
			return
		}
		lines = append(lines, Line32{a, b})
	}
	distance := func(a, b Point32) float32 {
		dx, dy := a.X-b.X, a.Y-b.Y
		return float32(math.Sqrt(float64(dx*dx + dy*dy)))
	}
	for _, segment := range segments {
		switch segment.Op {
		case sfnt.SegmentOpMoveTo:
			if contour {
				add(previous, start)
			}
			start = point(segment.Args[0])
			previous = start
			contour = true
		case sfnt.SegmentOpLineTo:
			p := point(segment.Args[0])
			add(previous, p)
			previous = p
		case sfnt.SegmentOpQuadTo, sfnt.SegmentOpCubeTo:
			p0, p1, p2 := previous, point(segment.Args[0]), point(segment.Args[1])
			p3 := point(segment.Args[2])
			length := distance(p0, p1) + distance(p1, p2)
			if segment.Op == sfnt.SegmentOpCubeTo {
				length += distance(p2, p3)
			}
			steps := max(1, min(c.MaxSegments, int(length/c.CurveStep)+1))
			for i := 1; i <= steps; i++ {
				t := float32(i) / float32(steps)
				inv := 1 - t
				p := Point32{inv*inv*p0.X + 2*inv*t*p1.X + t*t*p2.X, inv*inv*p0.Y + 2*inv*t*p1.Y + t*t*p2.Y}
				if segment.Op == sfnt.SegmentOpCubeTo {
					inv2, t2 := inv*inv, t*t
					p = Point32{inv2*inv*p0.X + 3*inv2*t*p1.X + 3*inv*t2*p2.X + t2*t*p3.X, inv2*inv*p0.Y + 3*inv2*t*p1.Y + 3*inv*t2*p2.Y + t2*t*p3.Y}
				}
				add(previous, p)
				previous = p
			}
		}
	}
	if contour {
		add(previous, start)
	}
	return lines
}
