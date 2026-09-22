package geometry

import (
	"fmt"
	"math"
)

// DeformKind selects a coordinate operation. Stages run in their supplied order.
// The string values and data-only stage configuration can be stored by editors.
type DeformKind string

const (
	DeformScale     DeformKind = "scale"
	DeformReference DeformKind = "reference"
	DeformWobble    DeformKind = "wobble"
	DeformRipple    DeformKind = "ripple"
	DeformTwist     DeformKind = "twist"
	DeformRotate    DeformKind = "rotate_xyz"
	DeformTranslate DeformKind = "translate"
)

// Harmonic is one sine or cosine term. Phase and spatial coefficients are in
// radians; Speed is radians per second. Gain scales the enclosing field Amount.
type Harmonic struct {
	Gain, Speed, Spatial, Phase float64
	Cosine                      bool
}

// WobbleField displaces each axis by its own harmonic sum. Key is dotted with
// the original model point, independent of prior displacement. Radius measures
// the most recently captured reference point; zero disables radial influence.
// The influence is BaseInfluence + normalizedRadius*RadialInfluence.
type WobbleField struct {
	Key                                            Vec3
	X, Y, Z                                        []Harmonic
	Amount, Radius, BaseInfluence, RadialInfluence float64
}

// RippleField adds a radial wave along three combinations of original axes.
// DirectionX/Y/Z are rows of the displacement matrix. Coordinates are divided
// by CoordinateScale before applying these rows. Radius uses the captured
// reference; a preceding DeformReference stage can select its coordinate space.
type RippleField struct {
	Amplitude, Speed, Spatial, Phase   float64
	Radius, CoordinateScale            float64
	DirectionX, DirectionY, DirectionZ Vec3
}

// TwistField rotates around Axis (0=X, 1=Y, 2=Z). The angle depends on the
// original position along that axis, normalized by Radius. Time adds rotation
// through Speed; Phase is a uniform angular offset.
type TwistField struct {
	Axis                         int
	Amount, Radius, Speed, Phase float64
}

// DeformStage holds parameters only for its Kind. Vector means scale factors,
// Euler XYZ angles, or translation. DeformReference captures the current point
// for subsequent radial operations without changing it. Original coordinates
// always remain available to the spatial wave and twist fields.
type DeformStage struct {
	Kind   DeformKind
	Vector Vec3
	Wobble WobbleField
	Ripple RippleField
	Twist  TwistField
}

// DeformProgram applies an ordered deformation to arbitrary mesh/vectorball
// positions. It owns its configuration, borrows input/output slices, and never
// allocates during Apply. Source and destination may be the same slice.
// One program is intended for one animation/update thread.
type DeformProgram struct {
	stages   []deformStage
	prepared bool
	seconds  float64
}
type deformStage struct {
	config                             DeformStage
	sinX, cosX, sinY, cosY, sinZ, cosZ float64
}

func NewDeformProgram(stages []DeformStage) (*DeformProgram, error) {
	p := &DeformProgram{stages: make([]deformStage, len(stages))}
	for i, stage := range stages {
		if err := stage.Validate(); err != nil {
			return nil, fmt.Errorf("geometry: stage %d: %w", i, err)
		}
		stage.Wobble.X = append([]Harmonic(nil), stage.Wobble.X...)
		stage.Wobble.Y = append([]Harmonic(nil), stage.Wobble.Y...)
		stage.Wobble.Z = append([]Harmonic(nil), stage.Wobble.Z...)
		p.stages[i].config = stage
	}
	return p, nil
}

// SetVector updates a scale, rotation or translation stage without rebuilding
// storage. SetWobbleAmount and SetTwistAmount animate the respective strengths.
func (p *DeformProgram) SetVector(index int, value Vec3) error {
	if index < 0 || index >= len(p.stages) || !deformFiniteVec(value) {
		return fmt.Errorf("geometry: invalid vector stage update")
	}
	s := &p.stages[index].config
	if s.Kind != DeformScale && s.Kind != DeformRotate && s.Kind != DeformTranslate {
		return fmt.Errorf("geometry: stage does not accept a vector")
	}
	if s.Kind == DeformRotate && s.Vector != value {
		p.prepared = false
	}
	s.Vector = value
	return nil
}
func (p *DeformProgram) SetWobbleAmount(index int, amount float64) error {
	if index < 0 || index >= len(p.stages) || p.stages[index].config.Kind != DeformWobble || !deformFinite(amount) {
		return fmt.Errorf("geometry: invalid wobble update")
	}
	p.stages[index].config.Wobble.Amount = amount
	return nil
}
func (p *DeformProgram) SetTwistAmount(index int, amount float64) error {
	if index < 0 || index >= len(p.stages) || p.stages[index].config.Kind != DeformTwist || !deformFinite(amount) {
		return fmt.Errorf("geometry: invalid twist update")
	}
	p.stages[index].config.Twist.Amount = amount
	return nil
}

// Apply returns the number of written points. Invalid time writes no points.
// Rotations are sampled once per batch. Sequential XYZ arithmetic preserves the
// same rounding as separate Euler rotations, unlike a multiplied matrix.
func (p *DeformProgram) Apply(dst, source []Vec3, seconds float64) int {
	if p == nil || !deformFinite(seconds) {
		return 0
	}
	if !p.prepared || p.seconds != seconds {
		for i := range p.stages {
			s := &p.stages[i]
			if s.config.Kind == DeformRotate {
				s.sinX, s.cosX = math.Sincos(s.config.Vector.X)
				s.sinY, s.cosY = math.Sincos(s.config.Vector.Y)
				s.sinZ, s.cosZ = math.Sincos(s.config.Vector.Z)
			}
		}
		p.prepared = true
		p.seconds = seconds
	}
	n := min(len(dst), len(source))
	for i, original := range source[:n] {
		point, reference := original, original
		for j := range p.stages {
			stage := &p.stages[j]
			c := &stage.config
			switch c.Kind {
			case DeformReference:
				reference = point
			case DeformScale:
				point = Vec3{point.X * c.Vector.X, point.Y * c.Vector.Y, point.Z * c.Vector.Z}
			case DeformTranslate:
				point = point.Add(c.Vector)
			case DeformRotate:
				y, z := point.Y*stage.cosX-point.Z*stage.sinX, point.Y*stage.sinX+point.Z*stage.cosX
				x, z := point.X*stage.cosY+z*stage.sinY, -point.X*stage.sinY+z*stage.cosY
				point = Vec3{x*stage.cosZ - y*stage.sinZ, x*stage.sinZ + y*stage.cosZ, z}
			case DeformWobble:
				w := c.Wobble
				key := original.Dot(w.Key)
				influence := w.BaseInfluence
				if w.Radius > 0 {
					influence += deformRadius(reference, w.Radius) * w.RadialInfluence
				}
				point.X += deformHarmonics(w.X, key, seconds, w.Amount) * influence
				point.Y += deformHarmonics(w.Y, key, seconds, w.Amount) * influence
				point.Z += deformHarmonics(w.Z, key, seconds, w.Amount) * influence
			case DeformRipple:
				r := c.Ripple
				wave := math.Sin(seconds*r.Speed+deformRadius(reference, r.Radius)*r.Spatial+r.Phase) * r.Amplitude
				normalized := Vec3{original.X / r.CoordinateScale, original.Y / r.CoordinateScale, original.Z / r.CoordinateScale}
				point.X += wave * normalized.Dot(r.DirectionX)
				point.Y += wave * normalized.Dot(r.DirectionY)
				point.Z += wave * normalized.Dot(r.DirectionZ)
			case DeformTwist:
				t := c.Twist
				if t.Amount == 0 && t.Speed == 0 && t.Phase == 0 {
					continue
				}
				axis := [3]float64{original.X, original.Y, original.Z}[t.Axis]
				angle := t.Amount*(axis/t.Radius) + seconds*t.Speed + t.Phase
				s, co := math.Sincos(angle)
				switch t.Axis {
				case 0:
					point.Y, point.Z = point.Y*co-point.Z*s, point.Y*s+point.Z*co
				case 1:
					point.X, point.Z = point.X*co-point.Z*s, point.X*s+point.Z*co
				case 2:
					point.X, point.Y = point.X*co-point.Y*s, point.X*s+point.Y*co
				}
			}
		}
		dst[i] = point
	}
	return n
}

// Deform adapts the program to effects.MeshEffect.Deform. Rotation preparation
// is cached by time, so all vertices in one update share the same sampled frame.
// Prefer Apply for point-cloud batches; it also supports in-place deformation.
func (p *DeformProgram) Deform(_ int, point Vec3, seconds float64) Vec3 {
	single := [1]Vec3{point}
	p.Apply(single[:], single[:], seconds)
	return single[0]
}

func deformHarmonics(waves []Harmonic, key, seconds, amount float64) float64 {
	value := 0.0
	for _, wave := range waves {
		phase := seconds*wave.Speed + key*wave.Spatial + wave.Phase
		var sample float64
		if wave.Cosine {
			sample = math.Cos(phase)
		} else {
			sample = math.Sin(phase)
		}
		value += sample * amount * wave.Gain
	}
	return value
}
func deformRadius(point Vec3, radius float64) float64 {
	return math.Sqrt(point.X*point.X+point.Y*point.Y+point.Z*point.Z) / radius
}
func deformFinite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func deformFiniteVec(v Vec3) bool { return deformFinite(v.X) && deformFinite(v.Y) && deformFinite(v.Z) }
func (s DeformStage) Validate() error {
	switch s.Kind {
	case DeformReference:
		return nil
	case DeformScale, DeformRotate, DeformTranslate:
		if deformFiniteVec(s.Vector) {
			return nil
		}
	case DeformWobble:
		w := s.Wobble
		if !deformFiniteVec(w.Key) || !deformFinite(w.Amount) || !deformFinite(w.Radius) || w.Radius < 0 || !deformFinite(w.BaseInfluence) || !deformFinite(w.RadialInfluence) {
			break
		}
		for _, list := range [][]Harmonic{w.X, w.Y, w.Z} {
			for _, h := range list {
				if !deformFinite(h.Gain) || !deformFinite(h.Speed) || !deformFinite(h.Spatial) || !deformFinite(h.Phase) {
					return fmt.Errorf("geometry: nonfinite harmonic")
				}
			}
		}
		return nil
	case DeformRipple:
		r := s.Ripple
		if deformFinite(r.Amplitude) && deformFinite(r.Speed) && deformFinite(r.Spatial) && deformFinite(r.Phase) && deformFinite(r.Radius) && r.Radius > 0 && deformFinite(r.CoordinateScale) && r.CoordinateScale > 0 && deformFiniteVec(r.DirectionX) && deformFiniteVec(r.DirectionY) && deformFiniteVec(r.DirectionZ) {
			return nil
		}
	case DeformTwist:
		t := s.Twist
		if t.Axis >= 0 && t.Axis < 3 && deformFinite(t.Amount) && deformFinite(t.Radius) && t.Radius > 0 && deformFinite(t.Speed) && deformFinite(t.Phase) {
			return nil
		}
	}
	return fmt.Errorf("geometry: invalid %q deformation parameters", s.Kind)
}
