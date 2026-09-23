package motion

import "math"

// HarmonicTerm contributes one sinusoidal displacement on an axis.
type HarmonicTerm struct {
	Amplitude, Rate, Phase float64
	Cos                    bool
}

// HarmonicTranslation sums independent terms on X and Y. Sampling once per
// group Update keeps synchronized sprites in the same position offset.
type HarmonicTranslation struct {
	X, Y []HarmonicTerm
}

func (h HarmonicTranslation) At(phase float64) Point {
	p := Point{}
	for _, term := range h.X {
		v := math.Sin(phase*term.Rate + term.Phase)
		if term.Cos {
			v = math.Cos(phase*term.Rate + term.Phase)
		}
		p.X += term.Amplitude * v
	}
	for _, term := range h.Y {
		v := math.Sin(phase*term.Rate + term.Phase)
		if term.Cos {
			v = math.Cos(phase*term.Rate + term.Phase)
		}
		p.Y += term.Amplitude * v
	}
	return p
}
