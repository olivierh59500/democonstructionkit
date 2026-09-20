package sprites

import (
	"fmt"
	"math"

	"github.com/olivierh59500/democonstructionkit/motion"
)

// Fill selects which points of a volume become balls. Edges has no duplicate
// corner balls; Surface includes the faces; Solid also fills the interior.
type Fill uint8

const (
	Edges Fill = iota
	Surface
	Solid
)

type CubeConfig struct {
	Size     float64
	Segments int // Number of intervals along each edge, at least one.
	Fill     Fill
	Image    int
}

func Cube(c CubeConfig) ([]Point, error) {
	if !dimension(c.Size) || !segments(c.Segments) || c.Fill > Solid {
		return nil, fmt.Errorf("sprites: invalid cube dimensions, segments or fill")
	}
	var points []Point
	n := c.Segments
	for z := 0; z <= n; z++ {
		for y := 0; y <= n; y++ {
			for x := 0; x <= n; x++ {
				boundary := 0
				for _, i := range []int{x, y, z} {
					if i == 0 || i == n {
						boundary++
					}
				}
				if c.Fill == Edges && boundary < 2 || c.Fill == Surface && boundary == 0 {
					continue
				}
				points = append(points, Point{X: coordinate(x, n, c.Size), Y: coordinate(y, n, c.Size), Z: coordinate(z, n, c.Size), Image: c.Image})
			}
		}
	}
	return points, nil
}

type PyramidConfig struct {
	Width, Height float64
	Segments      int
	Fill          Fill
	Image         int
}

// Pyramid has a square base and its apex on positive Y. Each level shares the
// same horizontal point spacing; the apex is emitted exactly once.
func Pyramid(c PyramidConfig) ([]Point, error) {
	if !dimension(c.Width) || !dimension(c.Height) || !segments(c.Segments) || c.Fill > Solid {
		return nil, fmt.Errorf("sprites: invalid pyramid dimensions, segments or fill")
	}
	var points []Point
	n := c.Segments
	for level := 0; level <= n; level++ {
		m := n - level
		for z := 0; z <= m; z++ {
			for x := 0; x <= m; x++ {
				xEdge, zEdge := x == 0 || x == m, z == 0 || z == m
				if c.Fill == Edges && !(xEdge && zEdge || level == 0 && (xEdge || zEdge)) {
					continue
				}
				if c.Fill == Surface && level != 0 && !xEdge && !zEdge {
					continue
				}
				points = append(points, Point{X: (float64(x) - float64(m)/2) * c.Width / float64(n), Y: coordinate(level, n, c.Height), Z: (float64(z) - float64(m)/2) * c.Width / float64(n), Image: c.Image})
			}
		}
	}
	return points, nil
}

type PlaneConfig struct {
	Width, Height float64
	Columns, Rows int // Intervals, so the point count is (Columns+1)*(Rows+1).
	Image         int
}

func Plane(c PlaneConfig) ([]Point, error) {
	if !dimension(c.Width) || !dimension(c.Height) || !segments(c.Columns) || !segments(c.Rows) {
		return nil, fmt.Errorf("sprites: invalid plane dimensions or segments")
	}
	points := make([]Point, 0, (c.Columns+1)*(c.Rows+1))
	for y := 0; y <= c.Rows; y++ {
		for x := 0; x <= c.Columns; x++ {
			points = append(points, Point{X: coordinate(x, c.Columns, c.Width), Y: coordinate(y, c.Rows, c.Height), Image: c.Image})
		}
	}
	return points, nil
}

// Flag deforms an existing XY plane along Z, without modifying the rest shape.
// Spatial is in radians per model unit. PinLeft keeps X=-Width/2 stationary.
// dst and rest may alias; always pass an unchanged rest shape to avoid drift.
type Flag struct {
	Width    float64
	Wave     motion.Wave
	RowPhase float64
	PinLeft  bool
}

func (f Flag) Apply(dst, rest []Point, seconds float64) {
	for i := 0; i < min(len(dst), len(rest)); i++ {
		p := rest[i]
		gain := 1.0
		if f.PinLeft {
			gain = 0
			if f.Width > 0 {
				gain = motion.Linear((p.X + f.Width/2) / f.Width)
			}
		}
		w := f.Wave
		w.Phase += p.Y * f.RowPhase
		p.Z += gain * w.At(p.X, seconds)
		dst[i] = p
	}
}

func dimension(x float64) bool                  { return x > 0 && !math.IsNaN(x) && !math.IsInf(x, 0) }
func segments(n int) bool                       { return n >= 1 && n <= 64 }
func coordinate(i, n int, size float64) float64 { return (float64(i)/float64(n) - .5) * size }
