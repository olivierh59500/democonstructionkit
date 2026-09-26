package sprites

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// ProjectedObjectConfig selects one point source and its independent animation,
// camera and sprite material. Images passed to Draw are borrowed. Flag may
// deform a plane or custom XY points while the unchanged rest shape is retained.
type ProjectedObjectConfig struct {
	Cube           *CubeConfig
	Pyramid        *PyramidConfig
	Plane          *PlaneConfig
	Points         []Point
	Flag           *Flag
	Rotation       geometry.Vec3
	RotationStep   geometry.Vec3
	Position       geometry.Vec3
	Scale          float64
	Focal          float64
	CenterX        float64
	CenterY        float64
	TimeStart      float64
	YUp            bool
	AscendingDepth bool
	ScaleImages    bool
	Options        ebiten.DrawImageOptions
}

// ProjectedObject owns one animated point shape, rotation clock, model matrix
// and projector. Multiple instances can render different objects in layer order
// without sharing scratch buffers or allocating during ordinary updates.
type ProjectedObject struct {
	config    ProjectedObjectConfig
	rest      []Point
	points    []Point
	rotation  geometry.Vec3
	matrix    [9]float64
	projector Projector
}

func projectedObjectFinite(values ...float64) bool {
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return false
		}
	}
	return true
}

func NewProjectedObject(config ProjectedObjectConfig) (*ProjectedObject, error) {
	sources := 0
	if config.Cube != nil {
		sources++
	}
	if config.Pyramid != nil {
		sources++
	}
	if config.Plane != nil {
		sources++
	}
	if len(config.Points) > 0 {
		sources++
	}
	if sources != 1 || config.Flag != nil && (config.Cube != nil || config.Pyramid != nil) ||
		config.Scale <= 0 || config.Focal <= 0 ||
		!projectedObjectFinite(config.Scale, config.Focal, config.CenterX, config.CenterY, config.TimeStart,
			config.Rotation.X, config.Rotation.Y, config.Rotation.Z,
			config.RotationStep.X, config.RotationStep.Y, config.RotationStep.Z,
			config.Position.X, config.Position.Y, config.Position.Z) {
		return nil, fmt.Errorf("sprites: invalid projected object source or transform")
	}
	if config.Flag != nil {
		f := *config.Flag
		if !projectedObjectFinite(f.Width, f.RowPhase, f.Wave.Amplitude, f.Wave.Spatial,
			f.Wave.Speed, f.Wave.Phase, f.Wave.Offset) || f.PinLeft && f.Width <= 0 {
			return nil, fmt.Errorf("sprites: invalid projected flag")
		}
		config.Flag = &f
	}
	var points []Point
	var err error
	switch {
	case config.Cube != nil:
		points, err = Cube(*config.Cube)
	case config.Pyramid != nil:
		points, err = Pyramid(*config.Pyramid)
	case config.Plane != nil:
		points, err = Plane(*config.Plane)
	default:
		points = append([]Point(nil), config.Points...)
	}
	if err != nil {
		return nil, err
	}
	for _, point := range points {
		if !projectedObjectFinite(point.X, point.Y, point.Z) {
			return nil, fmt.Errorf("sprites: invalid projected object point")
		}
	}
	config.Cube, config.Pyramid, config.Plane, config.Points = nil, nil, nil, nil
	object := &ProjectedObject{config: config, rest: points, points: make([]Point, len(points)), rotation: config.Rotation}
	copy(object.points, object.rest)
	if object.config.Flag != nil {
		object.config.Flag.Apply(object.points, object.rest, config.TimeStart)
	}
	object.matrix = [9]float64(geometry.RotateXYZScaled(object.rotation, config.Scale))
	return object, nil
}

// Update applies the flag at an absolute scene time and then advances each
// Euler angle once. This order retains the source's pre-draw rotation step.
func (object *ProjectedObject) Update(frame kit.Frame) error {
	if !projectedObjectFinite(frame.Time) {
		return fmt.Errorf("sprites: nonfinite projected object time")
	}
	next := geometry.Vec3{
		X: object.rotation.X + object.config.RotationStep.X,
		Y: object.rotation.Y + object.config.RotationStep.Y,
		Z: object.rotation.Z + object.config.RotationStep.Z,
	}
	if !projectedObjectFinite(next.X, next.Y, next.Z) {
		return fmt.Errorf("sprites: projected object rotation overflow")
	}
	if object.config.Flag != nil {
		object.config.Flag.Apply(object.points, object.rest, frame.Time)
		for _, point := range object.points {
			if !projectedObjectFinite(point.X, point.Y, point.Z) {
				return fmt.Errorf("sprites: projected object deformation overflow")
			}
		}
	}
	object.rotation = next
	object.matrix = [9]float64(geometry.RotateXYZScaled(next, object.config.Scale))
	return nil
}

// Draw projects the prepared points in depth order. The destination and sprite
// bank are borrowed; repeated Draw calls do not advance motion.
func (object *ProjectedObject) Draw(dst *ebiten.Image, images []*ebiten.Image) {
	if object == nil || dst == nil {
		return
	}
	c := object.config
	object.projector.Draw(dst, object.points, images, Projection{
		Matrix:    object.matrix,
		Translate: Point{X: c.Position.X, Y: c.Position.Y, Z: c.Position.Z},
		Focal:     c.Focal, CenterX: c.CenterX, CenterY: c.CenterY,
		YUp: c.YUp, AscendingDepth: c.AscendingDepth, ScaleImages: c.ScaleImages,
		Options: c.Options,
	})
}

// Points returns the current borrowed model points for inspection or masking.
func (object *ProjectedObject) Points() []Point { return object.points }

func (object *ProjectedObject) Rotation() geometry.Vec3 { return object.rotation }

func (object *ProjectedObject) Position() geometry.Vec3 { return object.config.Position }

// SetPosition changes projection placement without resetting flag or rotation.
func (object *ProjectedObject) SetPosition(position geometry.Vec3) error {
	if !projectedObjectFinite(position.X, position.Y, position.Z) {
		return fmt.Errorf("sprites: nonfinite projected object position")
	}
	object.config.Position = position
	return nil
}

// SetRotationStep changes all three angular speeds without moving this pose.
func (object *ProjectedObject) SetRotationStep(step geometry.Vec3) error {
	if !projectedObjectFinite(step.X, step.Y, step.Z) {
		return fmt.Errorf("sprites: nonfinite projected object rotation step")
	}
	object.config.RotationStep = step
	return nil
}

// SetScale changes the next and current model matrix without resetting phase.
func (object *ProjectedObject) SetScale(scale float64) error {
	if scale <= 0 || !projectedObjectFinite(scale) {
		return fmt.Errorf("sprites: invalid projected object scale")
	}
	object.config.Scale = scale
	object.matrix = [9]float64(geometry.RotateXYZScaled(object.rotation, scale))
	return nil
}

// Reset restores the rest shape, flag time and initial rotation while retaining
// live position, scale and angular-speed edits.
func (object *ProjectedObject) Reset() {
	object.rotation = object.config.Rotation
	copy(object.points, object.rest)
	if object.config.Flag != nil {
		object.config.Flag.Apply(object.points, object.rest, object.config.TimeStart)
	}
	object.matrix = [9]float64(geometry.RotateXYZScaled(object.rotation, object.config.Scale))
}
