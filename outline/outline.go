// Package outline converts TrueType/OpenType outlines into reusable line segments.
package outline

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/geometry"
	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

type Segment struct{ A, B geometry.Vec2 }
type Glyph struct {
	Segments []Segment
	Advance  float64
}
type Font struct {
	source *sfnt.Font
	size   fixed.Int26_6
	steps  int
}

// New flattens curves with a fixed subdivision count (1–128) at the given pixel size.
func New(data []byte, size float64, steps int) (*Font, error) {
	if size <= 0 || size > 32767 || math.IsNaN(size) || steps < 1 || steps > 128 {
		return nil, fmt.Errorf("outline: invalid size or subdivisions")
	}
	f, err := sfnt.Parse(data)
	if err != nil {
		return nil, err
	}
	return &Font{source: f, size: fixed.Int26_6(size * 64), steps: steps}, nil
}
func point(p fixed.Point26_6) geometry.Vec2 {
	return geometry.Vec2{X: float64(p.X) / 64, Y: float64(p.Y) / 64}
}
func lerp(a, b geometry.Vec2, t float64) geometry.Vec2 {
	return geometry.Vec2{X: a.X + (b.X-a.X)*t, Y: a.Y + (b.Y-a.Y)*t}
}
func (f *Font) Glyph(r rune) (Glyph, error) {
	var buffer sfnt.Buffer
	index, err := f.source.GlyphIndex(&buffer, r)
	if err != nil {
		return Glyph{}, err
	}
	segments, err := f.source.LoadGlyph(&buffer, index, f.size, nil)
	if err != nil {
		return Glyph{}, err
	}
	g := Glyph{}
	current := geometry.Vec2{}
	for _, s := range segments {
		switch s.Op {
		case sfnt.SegmentOpMoveTo:
			current = point(s.Args[0])
		case sfnt.SegmentOpLineTo:
			next := point(s.Args[0])
			g.Segments = append(g.Segments, Segment{current, next})
			current = next
		case sfnt.SegmentOpQuadTo, sfnt.SegmentOpCubeTo:
			start := current
			for i := 1; i <= f.steps; i++ {
				t := float64(i) / float64(f.steps)
				a, b := lerp(start, point(s.Args[0]), t), lerp(point(s.Args[0]), point(s.Args[1]), t)
				next := lerp(a, b, t)
				if s.Op == sfnt.SegmentOpCubeTo {
					c := lerp(point(s.Args[1]), point(s.Args[2]), t)
					next = lerp(lerp(a, b, t), lerp(b, c, t), t)
				}
				g.Segments = append(g.Segments, Segment{current, next})
				current = next
			}
		}
	}
	advance, err := f.source.GlyphAdvance(&buffer, index, f.size, font.HintingNone)
	if err != nil {
		return Glyph{}, err
	}
	g.Advance = float64(advance) / 64
	return g, nil
}

// Text returns independent line endpoints and edge indices for a wireframe effect.
func (f *Font) Text(message string, spacing float64) ([]geometry.Vec3, [][2]int, error) {
	if math.IsNaN(spacing) || math.IsInf(spacing, 0) {
		return nil, nil, fmt.Errorf("outline: spacing must be finite")
	}
	var points []geometry.Vec3
	var edges [][2]int
	x, y := 0.0, 0.0
	for _, r := range message {
		if r == '\n' {
			x = 0
			y += float64(f.size) / 64
			continue
		}
		g, err := f.Glyph(r)
		if err != nil {
			return nil, nil, err
		}
		for _, s := range g.Segments {
			i := len(points)
			points = append(points, geometry.Vec3{X: x + s.A.X, Y: y + s.A.Y}, geometry.Vec3{X: x + s.B.X, Y: y + s.B.Y})
			edges = append(edges, [2]int{i, i + 1})
		}
		x += g.Advance + spacing
	}
	return points, edges, nil
}
