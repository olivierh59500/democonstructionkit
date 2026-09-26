package geometry

import (
	"fmt"
	"math"
)

// SinusGridConfig deforms a row-major point grid along Z. Envelope and travel
// phases advance only after each complete point update.
type SinusGridConfig struct {
	Columns, Rows                    int
	Amplitude, SpatialStep           float64
	SpatialNumerator, SpatialDivisor float64 // Optional ordered index*numerator/divisor sampling.
	TravelPhase, TravelStep          float64
	EnvelopePhase, EnvelopeStep      float64
}

type SinusGrid struct {
	config             SinusGridConfig
	phaseSin, phaseCos []float64
	travel, envelope   float64
}

func NewSinusGrid(c SinusGridConfig) (*SinusGrid, error) {
	if c.Columns < 1 || c.Rows < 1 || c.Columns > 256 || c.Rows > 256 || c.Columns*c.Rows > 65_536 ||
		!finiteAnimation(c.Amplitude, c.SpatialStep, c.SpatialNumerator, c.SpatialDivisor, c.TravelPhase, c.TravelStep, c.EnvelopePhase, c.EnvelopeStep) || c.SpatialDivisor < 0 {
		return nil, fmt.Errorf("geometry: invalid sinus grid configuration")
	}
	count := c.Columns + c.Rows - 1
	g := &SinusGrid{config: c, phaseSin: make([]float64, count), phaseCos: make([]float64, count), travel: c.TravelPhase, envelope: c.EnvelopePhase}
	for i := range g.phaseSin {
		phase := float64(i) * c.SpatialStep
		if c.SpatialDivisor > 0 {
			phase = float64(i) * c.SpatialNumerator / c.SpatialDivisor
		}
		g.phaseSin[i], g.phaseCos[i] = math.Sincos(phase)
	}
	return g, nil
}

func (g *SinusGrid) Step(points PointWriter) {
	if g == nil || points == nil {
		return
	}
	amplitude := g.config.Amplitude * math.Sin(g.envelope)
	sinTravel, cosTravel := math.Sincos(g.travel)
	index := 0
	for row := 0; row < g.config.Rows; row++ {
		for column := 0; column < g.config.Columns; column++ {
			if index < points.Len() {
				phase := row + column
				p := points.XYZ(index)
				p.Z = amplitude * (cosTravel*g.phaseCos[phase] - sinTravel*g.phaseSin[phase])
				points.SetXYZ(index, p)
			}
			index++
		}
	}
	g.travel += g.config.TravelStep
	g.envelope += g.config.EnvelopeStep
}

// RotorConfig moves point pairs around the Y axis while leaving their Y and
// artwork fields unchanged. RequireFull delays the clock for incomplete pairs.
type RotorConfig struct {
	Radii       []float64
	Offset      float64
	Phase       float64
	PhaseStep   float64
	RequireFull bool
}

type Rotors struct {
	config RotorConfig
	phase  float64
}

func NewRotors(c RotorConfig) (*Rotors, error) {
	if len(c.Radii) == 0 || len(c.Radii) > 256 || !finiteAnimation(c.Offset, c.Phase, c.PhaseStep) {
		return nil, fmt.Errorf("geometry: invalid rotor configuration")
	}
	for _, radius := range c.Radii {
		if !finiteAnimation(radius) {
			return nil, fmt.Errorf("geometry: nonfinite rotor radius")
		}
	}
	c.Radii = append([]float64(nil), c.Radii...)
	return &Rotors{config: c, phase: c.Phase}, nil
}

func (r *Rotors) Step(points PointWriter) {
	if r == nil || points == nil || r.config.RequireFull && points.Len() < 2*len(r.config.Radii) {
		return
	}
	sinPhase, cosPhase := math.Sincos(r.phase)
	for i, radius := range r.config.Radii {
		x, z := radius*sinPhase, radius*cosPhase
		if index := i * 2; index < points.Len() {
			p := points.XYZ(index)
			p.X, p.Z = x, z-r.config.Offset
			points.SetXYZ(index, p)
		}
		if index := i*2 + 1; index < points.Len() {
			p := points.XYZ(index)
			p.X, p.Z = -x, -z-r.config.Offset
			points.SetXYZ(index, p)
		}
	}
	r.phase += r.config.PhaseStep
}

// YOrbitConfig moves a full model position after the scene's own linear
// translation. RatioStep stops at the chosen bound without clamping overshoot.
type YOrbitConfig struct {
	Center                     Vec3
	Radius                     float64
	Phase, PhaseStep           float64
	Ratio, RatioStep, Min, Max float64
}

type YOrbit struct {
	config    YOrbitConfig
	phase     float64
	ratio     float64
	ratioStep float64
}

func NewYOrbit(c YOrbitConfig) (*YOrbit, error) {
	if !finiteAnimation(c.Center.X, c.Center.Y, c.Center.Z, c.Radius, c.Phase, c.PhaseStep, c.Ratio, c.RatioStep, c.Min, c.Max) || c.Max < c.Min {
		return nil, fmt.Errorf("geometry: invalid Y orbit configuration")
	}
	return &YOrbit{config: c, phase: c.Phase, ratio: c.Ratio, ratioStep: c.RatioStep}, nil
}

func (o *YOrbit) Step() Vec3 {
	if o == nil {
		return Vec3{}
	}
	o.phase += o.config.PhaseStep
	if o.ratioStep > 0 {
		o.ratio += o.ratioStep
		if o.ratio >= o.config.Max {
			o.ratioStep = 0
		}
	} else if o.ratioStep < 0 {
		o.ratio += o.ratioStep
		if o.ratio <= o.config.Min {
			o.ratioStep = 0
		}
	}
	return o.Pose()
}

func (o *YOrbit) Pose() Vec3 {
	if o == nil {
		return Vec3{}
	}
	sinPhase, cosPhase := math.Sincos(o.phase)
	return Vec3{X: o.config.Center.X + o.ratio*o.config.Radius*cosPhase,
		Y: o.config.Center.Y, Z: o.config.Center.Z + o.ratio*o.config.Radius*sinPhase}
}

// BounceCurveConfig accepts an authored Y table or generates repeated cosine
// sweeps followed by one final forward sweep. Every dimension and decay is
// independent, so the same controller can move a logo, sprite or point scene.
type BounceCurveConfig struct {
	Samples                                      []float64
	Radius, Offset, RadiusDecay, OffsetDecay     float64
	StepDegrees, BackwardStart, FinalStepDegrees float64
	Loops                                        int
	Divisor                                      float64
}

type BounceCurve struct {
	samples []float64
	index   int
	y       float64
}

func NewBounceCurve(c BounceCurveConfig) (*BounceCurve, error) {
	if len(c.Samples) > 1_000_000 {
		return nil, fmt.Errorf("geometry: bounce table exceeds budget")
	}
	if len(c.Samples) > 0 {
		for _, value := range c.Samples {
			if !finiteAnimation(value) {
				return nil, fmt.Errorf("geometry: nonfinite bounce sample")
			}
		}
		return &BounceCurve{samples: append([]float64(nil), c.Samples...)}, nil
	}
	if c.Loops < 0 || c.Loops > 1000 || c.StepDegrees <= 0 || c.FinalStepDegrees <= 0 || c.Divisor == 0 ||
		!finiteAnimation(c.Radius, c.Offset, c.RadiusDecay, c.OffsetDecay, c.StepDegrees, c.BackwardStart, c.FinalStepDegrees, c.Divisor) {
		return nil, fmt.Errorf("geometry: invalid bounce curve configuration")
	}
	curve := &BounceCurve{samples: make([]float64, 0, 540)}
	radius, offset := c.Radius, c.Offset
	appendSample := func(angle float64) error {
		if len(curve.samples) >= 1_000_000 {
			return fmt.Errorf("geometry: bounce table exceeds budget")
		}
		sample := -(offset - radius*math.Cos(angle*math.Pi/180)) / c.Divisor
		if !finiteAnimation(sample) {
			return fmt.Errorf("geometry: nonfinite generated bounce sample")
		}
		curve.samples = append(curve.samples, sample)
		radius -= c.RadiusDecay
		offset -= c.OffsetDecay
		return nil
	}
	for loop := 0; loop < c.Loops; loop++ {
		for angle := 0.0; angle < 90; angle += c.StepDegrees {
			if err := appendSample(angle); err != nil {
				return nil, err
			}
		}
		for angle := c.BackwardStart; angle >= 0; angle -= c.StepDegrees {
			if err := appendSample(angle); err != nil {
				return nil, err
			}
		}
	}
	for angle := 0.0; angle < 90; angle += c.FinalStepDegrees {
		if err := appendSample(angle); err != nil {
			return nil, err
		}
	}
	if len(curve.samples) == 0 {
		return nil, fmt.Errorf("geometry: empty bounce curve")
	}
	return curve, nil
}

// Step returns the current authored Y position and advances the table cursor.
func (b *BounceCurve) Step() float64 {
	if b == nil || len(b.samples) == 0 {
		return 0
	}
	b.y = b.samples[b.index]
	b.index++
	if b.index == len(b.samples) {
		b.index = 0
	}
	return b.y
}

func (b *BounceCurve) At() float64 { return b.y }
func (b *BounceCurve) Len() int    { return len(b.samples) }

func finiteAnimation(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}
