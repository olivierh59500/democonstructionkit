package motion

import (
	"fmt"
	"math"
	"sort"
)

// Point is a position or direction in image pixels. Motion stays independent of
// the renderer and of the geometry package, whose handoff code uses motion.
type Point struct{ X, Y float64 }

// Curve samples a normalized parameter in [0,1]. SamplePath turns any user curve
// into a reusable distance-based path, avoiding uneven speed between samples.
type Curve func(float64) Point

// Path is an immutable polyline with an arc-length index. Sampling allocates
// nothing; create paths once, not once per glyph or frame.
type Path struct {
	points, tangents []Point
	ends             []float64
	length           float64
	closed           bool
}

func NewPolyline(points []Point, closed bool) (*Path, error) {
	if len(points) < 2 {
		return nil, fmt.Errorf("motion: a path needs at least two points")
	}
	p := &Path{closed: closed}
	for _, v := range points {
		if !pathFinite(v.X) || !pathFinite(v.Y) {
			return nil, fmt.Errorf("motion: nonfinite path point")
		}
		if len(p.points) == 0 || p.points[len(p.points)-1] != v {
			p.points = append(p.points, v)
		}
	}
	if len(p.points) < 2 {
		return nil, fmt.Errorf("motion: path has no length")
	}
	if closed && p.points[len(p.points)-1] != p.points[0] {
		p.points = append(p.points, p.points[0])
	}
	for i := 1; i < len(p.points); i++ {
		a, b := p.points[i-1], p.points[i]
		dx, dy := b.X-a.X, b.Y-a.Y
		length := math.Hypot(dx, dy)
		if !pathFinite(length) || !pathFinite(p.length+length) {
			return nil, fmt.Errorf("motion: path length overflow")
		}
		p.length += length
		p.ends = append(p.ends, p.length)
		p.tangents = append(p.tangents, Point{dx / length, dy / length})
	}
	return p, nil
}

// SamplePath approximates a curve with segments equal parameter intervals.
// More segments improve shape and speed accuracy, at construction-time cost.
func SamplePath(curve Curve, segments int, closed bool) (*Path, error) {
	if curve == nil || segments < 1 || segments > 1<<20 {
		return nil, fmt.Errorf("motion: invalid curve or segment count")
	}
	points := make([]Point, segments+1)
	for i := range points {
		points[i] = curve(float64(i) / float64(segments))
	}
	return NewPolyline(points, closed)
}

func (p *Path) Length() float64 {
	if p == nil {
		return 0
	}
	return p.length
}
func (p *Path) Closed() bool { return p != nil && p.closed }

// At returns a position and unit tangent at a distance in pixels. Open paths
// clamp to their endpoints; closed paths wrap positive and negative distances.
func (p *Path) At(distance float64) (Point, Point) {
	if p == nil || len(p.ends) == 0 {
		return Point{}, Point{}
	}
	if math.IsNaN(distance) {
		distance = 0
	}
	if p.closed {
		distance = Wrap(distance, p.length)
	} else {
		distance = math.Max(0, math.Min(p.length, distance))
	}
	i := sort.Search(len(p.ends), func(i int) bool { return p.ends[i] > distance })
	if i == len(p.ends) {
		return p.points[len(p.points)-1], p.tangents[len(p.tangents)-1]
	}
	start := 0.0
	if i > 0 {
		start = p.ends[i-1]
	}
	u := (distance - start) / (p.ends[i] - start)
	a, b := p.points[i], p.points[i+1]
	return Point{Lerp(a.X, b.X, u), Lerp(a.Y, b.Y, u)}, p.tangents[i]
}

func SineCurve(origin Point, width, amplitude, turns, phase float64) Curve {
	return func(u float64) Point {
		return Point{origin.X + u*width, origin.Y + amplitude*math.Sin(phase+u*turns*2*math.Pi)}
	}
}
func EllipseCurve(center, radius Point, phase float64) Curve {
	return func(u float64) Point {
		a := phase + u*2*math.Pi
		return Point{center.X + radius.X*math.Cos(a), center.Y + radius.Y*math.Sin(a)}
	}
}
func LissajousCurve(center, radius, frequency Point, phase float64) Curve {
	return func(u float64) Point {
		a := u * 2 * math.Pi
		return Point{center.X + radius.X*math.Sin(a*frequency.X+phase), center.Y + radius.Y*math.Sin(a*frequency.Y)}
	}
}
func FigureEightCurve(center, radius Point) Curve {
	return LissajousCurve(center, radius, Point{1, 2}, 0)
}
func BezierCurve(a, b, c, d Point) Curve {
	return func(u float64) Point {
		v := 1 - u
		return Point{v*v*v*a.X + 3*v*v*u*b.X + 3*v*u*u*c.X + u*u*u*d.X, v*v*v*a.Y + 3*v*v*u*b.Y + 3*v*u*u*c.Y + u*u*u*d.Y}
	}
}

// NewSpline samples a Catmull-Rom curve through supplied coordinates. Uniform
// splines can overshoot sharp corners; NewPolyline preserves hard corners.
func NewSpline(points []Point, closed bool, samplesPerSegment int) (*Path, error) {
	if len(points) < 2 || samplesPerSegment < 1 || len(points) > 1<<20/samplesPerSegment {
		return nil, fmt.Errorf("motion: invalid spline dimensions")
	}
	for _, p := range points {
		if !pathFinite(p.X) || !pathFinite(p.Y) {
			return nil, fmt.Errorf("motion: nonfinite spline point")
		}
	}
	segments := len(points) - 1
	if closed {
		segments = len(points)
	}
	at := func(i int) Point {
		if closed {
			i = ((i % len(points)) + len(points)) % len(points)
		} else {
			i = max(0, min(len(points)-1, i))
		}
		return points[i]
	}
	curve := func(u float64) Point {
		position := u * float64(segments)
		i := min(int(position), segments-1)
		t := position - float64(i)
		a, b, c, d := at(i-1), at(i), at(i+1), at(i+2)
		sample := func(a, b, c, d float64) float64 {
			return .5 * ((2 * b) + (-a+c)*t + (2*a-5*b+4*c-d)*t*t + (-a+3*b-3*c+d)*t*t*t)
		}
		return Point{sample(a.X, b.X, c.X, d.X), sample(a.Y, b.Y, c.Y, d.Y)}
	}
	return SamplePath(curve, segments*samplesPerSegment, closed)
}
func pathFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
