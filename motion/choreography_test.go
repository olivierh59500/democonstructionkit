package motion

import (
	"math"
	"testing"
)

func TestNestedOrbitMatchesAuthoredCoordinates(t *testing.T) {
	for _, c := range []struct{ center, radius Point }{
		{Point{X: 320, Y: 200}, Point{X: 160, Y: 400 / 3.7}},
		{Point{X: 384, Y: 270}, Point{X: 135, Y: 200}},
		{Point{X: 384, Y: 268}, Point{X: 192, Y: 536 / 2.7}},
	} {
		o := DefaultNestedOrbit(c.center, c.radius)
		phase := 0.0
		for i := 0; i < 10000; i++ {
			want := Point{X: c.center.X + c.radius.X*math.Cos(phase*4-math.Cos(phase-.1)), Y: c.center.Y + c.radius.Y*-math.Sin(phase*2.3-math.Cos(phase-.1))}
			if got := o.At(phase); math.Abs(got.X-want.X) > 1e-10 || math.Abs(got.Y-want.Y) > 1e-10 {
				t.Fatalf("tick %d: got %v want %v", i, got, want)
			}
			phase += .008
		}
	}
}

func TestWeaveMatchesAuthoredCoordinates(t *testing.T) {
	for _, c := range []struct {
		center, amplitude Point
		step              float64
	}{
		{Point{X: 308, Y: 190}, Point{X: 308, Y: 500.0 / 6}, 1.25},
		{Point{X: 380, Y: 277}, Point{X: 380, Y: 125.5}, 1},
		{Point{X: 306, Y: 207}, Point{X: 306, Y: 90.5}, 1},
	} {
		w := DefaultWeave(c.center, c.amplitude)
		for tick := 0; tick < 1000; tick++ {
			for i := 0; i < 12; i++ {
				phase := float64(tick) * c.step
				inter := phase + float64(i*5)
				want := Point{X: c.center.X + c.amplitude.X*math.Sin(inter/25)*math.Cos(inter/300), Y: c.center.Y + c.amplitude.Y*math.Sin(inter/37) + c.amplitude.Y*math.Cos(inter/17)}
				if got := w.At(phase, i); math.Abs(got.X-want.X) > 1e-10 || math.Abs(got.Y-want.Y) > 1e-10 {
					t.Fatalf("tick %d item %d: got %v want %v", tick, i, got, want)
				}
			}
		}
	}
}

func TestChoreographyParameters(t *testing.T) {
	o := NestedOrbit{Center: Point{X: 3, Y: 4}, Radius: Point{X: 7, Y: 9}, XRate: 2, YRate: 3}
	if got := o.At(0); got != (Point{X: 10, Y: 4}) {
		t.Fatal(got)
	}
	w := Weave{Center: Point{X: 3, Y: 4}, HorizontalAmplitude: 7, HorizontalPeriod: 1, Spacing: math.Pi / 2}
	if got := w.At(0, 1); got != (Point{X: 10, Y: 4}) {
		t.Fatal(got)
	}
	if got := (Weave{Center: Point{X: 3, Y: 4}}).At(100, 3); got != (Point{X: 3, Y: 4}) {
		t.Fatal(got)
	}
}

func BenchmarkWeave(b *testing.B) {
	w := DefaultWeave(Point{}, Point{X: 320, Y: 100})
	for i := 0; i < b.N; i++ {
		w.At(float64(i), i%12)
	}
}
