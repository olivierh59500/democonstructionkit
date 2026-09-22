package motion

import "math"

// NestedOrbit combines two orbital frequencies with a shared cosine modulation.
// Phase is caller-owned: use seconds for independent-time motion, or feed an
// existing fixed-tick phase to retain an authored animation exactly. Radius signs
// mirror an axis, while the modulation controls how the orbit folds over itself.
type NestedOrbit struct {
	Center, Radius                                   Point
	XRate, YRate                                     float64
	ModulationRate, ModulationPhase, ModulationDepth float64
}

// DefaultNestedOrbit returns the folded oval shared by several classic screens.
// Radius and Center remain entirely independent of image dimensions.
func DefaultNestedOrbit(center, radius Point) NestedOrbit {
	return NestedOrbit{Center: center, Radius: radius, XRate: 4, YRate: 2.3, ModulationRate: 1, ModulationPhase: -.1, ModulationDepth: 1}
}

func (o NestedOrbit) At(phase float64) Point {
	modulation := math.Cos(phase*o.ModulationRate+o.ModulationPhase) * o.ModulationDepth
	return Point{X: o.Center.X + o.Radius.X*math.Cos(phase*o.XRate-modulation), Y: o.Center.Y + o.Radius.Y*-math.Sin(phase*o.YRate-modulation)}
}

// Weave arranges independently phased sprites along a breathing horizontal sine
// and two vertical harmonics. Spacing is measured in phase units per item, not
// pixels; use Path.At for constant-distance spacing. Periods are divisors in
// phase units per radian. A zero period disables its component safely.
type Weave struct {
	Center                                                                 Point
	HorizontalAmplitude, VerticalAmplitude, VerticalSecondAmplitude        float64
	HorizontalPeriod, EnvelopePeriod, VerticalPeriod, VerticalSecondPeriod float64
	Spacing                                                                float64
}

func DefaultWeave(center, amplitude Point) Weave {
	return Weave{Center: center, HorizontalAmplitude: amplitude.X, VerticalAmplitude: amplitude.Y, VerticalSecondAmplitude: amplitude.Y, HorizontalPeriod: 25, EnvelopePeriod: 300, VerticalPeriod: 37, VerticalSecondPeriod: 17, Spacing: 5}
}

func (w Weave) At(phase float64, index int) Point {
	t := phase + float64(index)*w.Spacing
	p := w.Center
	if w.HorizontalPeriod != 0 {
		x := w.HorizontalAmplitude * math.Sin(t/w.HorizontalPeriod)
		if w.EnvelopePeriod != 0 {
			x *= math.Cos(t / w.EnvelopePeriod)
		}
		p.X += x
	}
	if w.VerticalPeriod != 0 {
		p.Y += w.VerticalAmplitude * math.Sin(t/w.VerticalPeriod)
	}
	if w.VerticalSecondPeriod != 0 {
		p.Y += w.VerticalSecondAmplitude * math.Cos(t/w.VerticalSecondPeriod)
	}
	return p
}
