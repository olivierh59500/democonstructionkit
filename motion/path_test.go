package motion

import (
	"math"
	"testing"
)

func TestPathDistanceAndClosedWrapping(t *testing.T) {
	points := []Point{{0, 0}, {30, 0}, {30, 0}, {30, 40}}
	p, err := NewPolyline(points, false)
	if err != nil {
		t.Fatal(err)
	}
	points[1] = Point{999, 999}
	if p.Length() != 70 {
		t.Fatal(p.Length())
	}
	for _, tt := range []struct {
		d    float64
		p, t Point
	}{{-1, Point{0, 0}, Point{1, 0}}, {15, Point{15, 0}, Point{1, 0}}, {35, Point{30, 5}, Point{0, 1}}, {99, Point{30, 40}, Point{0, 1}}} {
		got, tangent := p.At(tt.d)
		if got != tt.p || tangent != tt.t {
			t.Fatal(tt, got, tangent)
		}
	}
	c, err := NewPolyline([]Point{{0, 0}, {30, 0}, {30, 40}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if c.Length() != 120 {
		t.Fatal(c.Length())
	}
	a, _ := c.At(-10)
	b, _ := c.At(110)
	if a != b {
		t.Fatal(a, b)
	}
	if n := testing.AllocsPerRun(1000, func() { p.At(37) }); n != 0 {
		t.Fatalf("path sampling allocated %g", n)
	}
}
func TestSampledCurvesAndSpline(t *testing.T) {
	circle, err := SamplePath(EllipseCurve(Point{5, 6}, Point{30, 30}, 0), 256, true)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(circle.Length()-60*math.Pi) > .02 {
		t.Fatal(circle.Length())
	}
	p, _ := circle.At(circle.Length() / 4)
	if math.Hypot(p.X-5, p.Y-36) > 1e-8 {
		t.Fatal(p)
	}
	for _, curve := range []Curve{SineCurve(Point{}, 200, 20, 2, 0), FigureEightCurve(Point{}, Point{100, 50}), LissajousCurve(Point{}, Point{100, 60}, Point{3, 2}, math.Pi/2), BezierCurve(Point{}, Point{50, 100}, Point{100, -50}, Point{200, 0})} {
		p, err := SamplePath(curve, 128, false)
		if err != nil || p.Length() == 0 {
			t.Fatal(p, err)
		}
	}
	s, err := NewSpline([]Point{{0, 0}, {50, 20}, {100, 0}}, false, 32)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := s.At(-1)
	b, _ := s.At(s.Length() + 1)
	if a != (Point{}) || b != (Point{100, 0}) {
		t.Fatal(a, b)
	}
	for _, points := range [][]Point{nil, {{1, 2}}, {{1, 2}, {1, 2}}, {{math.NaN(), 0}, {1, 0}}} {
		if _, err := NewPolyline(points, false); err == nil {
			t.Fatal("invalid path accepted")
		}
	}
	if _, err := SamplePath(nil, 10, false); err == nil {
		t.Fatal("nil curve accepted")
	}
	if _, err := NewSpline([]Point{{0, 0}, {1, 1}}, true, 0); err == nil {
		t.Fatal("zero sampling accepted")
	}
}
