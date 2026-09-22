package sprites

import (
	"fmt"
	"math"
	"slices"

	"github.com/olivierh59500/democonstructionkit/geometry"
)

// DepthPolicy selects the lifetime of particles crossing the depth interval.
type DepthPolicy uint8

const (
	DepthFree DepthPolicy = iota
	DepthWrap
	DepthRespawn
)

// FieldConfig separates particle movement from its pixel, sprite or trail skin.
// Spawn is called in index order, both at construction and on respawn. It must
// return finite coordinates. Its reset flag distinguishes the two cases.
type FieldConfig struct {
	Count     int
	Points    []Point // Optional initial model; copied, never modified in place.
	Spawn     func(index int, reset bool) Point
	Depth     DepthPolicy
	Near, Far float64
}

// FieldView projects any field through the same camera. DepthOffset is applied
// before wrapping and rotation, permitting deterministic absolute-time travel.
// Angle rotates around the view axis in radians; camera Y points downward.
type FieldView struct {
	Camera    geometry.Camera
	Offset    geometry.Vec3
	Angle     float64
	SortDepth bool
}

// FieldSample is independent of its visual representation. Image chooses an
// atlas frame. Connected is false on first sampling, resets and depth wraps,
// so a streak renderer never draws a line across the field after recycling.
type FieldSample struct {
	Index, Image         int
	X, Y, Z, Scale       float64
	PreviousX, PreviousY float64
	Connected            bool
}

type fieldHistory struct {
	x, y  float64
	cycle float64
	valid bool
}

// Field owns bounded reusable simulation and projection storage. Step and Sample
// belong in Update; Samples and drawing never advance state. The same samples
// can be drawn on several layers with different skins/blends without reprojecting.
type Field struct {
	config  FieldConfig
	points  []Point
	samples []FieldSample
	history []fieldHistory
}

func NewField(c FieldConfig) (*Field, error) {
	if c.Count < 0 || c.Count > 1_000_000 || c.Depth > DepthRespawn ||
		(c.Depth != DepthFree && (!finiteField(c.Near) || !finiteField(c.Far) || c.Far <= c.Near)) ||
		(c.Points == nil && c.Count > 0 && c.Spawn == nil) || (c.Depth == DepthRespawn && c.Spawn == nil) {
		return nil, fmt.Errorf("sprites: invalid field configuration")
	}
	if c.Points != nil {
		c.Count = len(c.Points)
		if c.Count > 1_000_000 {
			return nil, fmt.Errorf("sprites: too many field points")
		}
	}
	f := &Field{config: c, points: make([]Point, c.Count), history: make([]fieldHistory, c.Count), samples: make([]FieldSample, 0, c.Count)}
	for i := range f.points {
		if c.Points != nil {
			f.points[i] = c.Points[i]
		} else {
			f.points[i] = c.Spawn(i, false)
		}
		if !validFieldPoint(f.points[i]) {
			return nil, fmt.Errorf("sprites: nonfinite field point %d", i)
		}
	}
	return f, nil
}

// ResetCount initializes a new population while retaining allocated capacity.
// Supplying a model without Spawn makes resizing invalid; construct another
// field instead. This is suitable for interactive particle-count controls.
func (f *Field) ResetCount(count int) error {
	if count < 0 || count > 1_000_000 || f.config.Spawn == nil {
		return fmt.Errorf("sprites: invalid field reset")
	}
	if cap(f.points) < count {
		f.points = make([]Point, count)
		f.history = make([]fieldHistory, count)
		f.samples = make([]FieldSample, 0, count)
	} else {
		f.points = f.points[:count]
		f.history = f.history[:count]
		clear(f.history)
		f.samples = f.samples[:0]
	}
	for i := range f.points {
		f.points[i] = f.config.Spawn(i, false)
		if !validFieldPoint(f.points[i]) {
			return fmt.Errorf("sprites: nonfinite reset point %d", i)
		}
	}
	return nil
}

// Step integrates one caller-controlled step. A fixed delta retains an authored
// cadence; seconds with a velocity per second work equally well. Large steps
// wrap without repeated subtraction. Respawn intentionally runs once per step.
func (f *Field) Step(delta float64, velocity geometry.Vec3) {
	if !finiteField(delta) || !finiteField(velocity.X) || !finiteField(velocity.Y) || !finiteField(velocity.Z) {
		return
	}
	for i := range f.points {
		p := &f.points[i]
		p.X += velocity.X * delta
		p.Y += velocity.Y * delta
		p.Z += velocity.Z * delta
		switch f.config.Depth {
		case DepthWrap:
			z, cycle := f.wrap(p.Z)
			p.Z = z
			if cycle != 0 {
				f.history[i].valid = false
			}
		case DepthRespawn:
			if p.Z <= f.config.Near || p.Z > f.config.Far {
				*p = f.config.Spawn(i, true)
				f.history[i].valid = false
			}
		}
	}
}

// Sample projects once per simulation update and caches the result. It retains
// source order by default; optional stable depth sorting supports translucent
// sprite fields without changing spawn order or history identity.
func (f *Field) Sample(v FieldView) []FieldSample {
	f.samples = f.samples[:0]
	sin, cos := math.Sincos(v.Angle)
	for i, p := range f.points {
		z, cycle := p.Z+v.Offset.Z, 0.0
		if f.config.Depth == DepthWrap {
			z, cycle = f.wrap(z)
		}
		x, y := p.X+v.Offset.X, p.Y+v.Offset.Y
		position := geometry.Vec3{X: x*cos - y*sin, Y: x*sin + y*cos, Z: z}
		xy, scale, visible := v.Camera.Project(position)
		h := &f.history[i]
		if !visible || !finiteField(xy.X) || !finiteField(xy.Y) || !finiteField(scale) {
			h.valid = false
			continue
		}
		f.samples = append(f.samples, FieldSample{Index: i, Image: p.Image, X: xy.X, Y: xy.Y, Z: z, Scale: scale,
			PreviousX: h.x, PreviousY: h.y, Connected: h.valid && h.cycle == cycle})
		*h = fieldHistory{x: xy.X, y: xy.Y, cycle: cycle, valid: true}
	}
	if v.SortDepth {
		slices.SortStableFunc(f.samples, func(a, b FieldSample) int {
			if a.Z > b.Z {
				return -1
			}
			if a.Z < b.Z {
				return 1
			}
			return 0
		})
	}
	return f.samples
}

// Samples returns borrowed storage, valid until the next Sample or ResetCount.
func (f *Field) Samples() []FieldSample { return f.samples }

func (f *Field) wrap(z float64) (float64, float64) {
	if z >= f.config.Near && z <= f.config.Far {
		return z, 0
	}
	period := f.config.Far - f.config.Near
	cycle := math.Floor((z - f.config.Near) / period)
	return z - period*cycle, cycle
}

func finiteField(x float64) bool   { return !math.IsNaN(x) && !math.IsInf(x, 0) }
func validFieldPoint(p Point) bool { return finiteField(p.X) && finiteField(p.Y) && finiteField(p.Z) }
