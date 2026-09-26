package effects

import (
	"errors"
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// SolidCubeTrainConfig describes a batched procession of independently
// rotating cubes. X and Y sample the same per-cube phase with their own wave
// rates; Path can replace both waves with an authored trajectory. CubeConfigs
// optionally replaces the common material for individual cubes.
type SolidCubeTrainConfig struct {
	Count               int
	Cube                SolidCubeConfig
	CubeConfigs         []SolidCubeConfig
	PhaseStart          float64
	PhaseSpacing        float64
	PhaseIndexOrigin    float64 // One samples spacing*(index+1), as in Coco.
	PhaseStep           float64
	X, Y                motion.Wave
	Path                func(index int, phase float64) motion.Point
	RotationStart       geometry.Vec3
	RotationSpacing     geometry.Vec3
	RotationStep        geometry.Vec3
	RotationIndexFactor geometry.Vec3
	Speed               float64 // Zero defaults to one; SetSpeed can pause later.
}

// SolidCubeTrain owns the cube geometry and one bounded batch. Update changes
// pose once per logical tick; repeated Draw calls retain the same positions.
type SolidCubeTrain struct {
	config SolidCubeTrainConfig
	cubes  []*SolidCube
	phases []float64
	batch  *SolidCubeBatch
	speed  float64
}

func NewSolidCubeTrain(c SolidCubeTrainConfig) (*SolidCubeTrain, error) {
	if c.Count < 1 || c.Count > 256 || len(c.CubeConfigs) != 0 && len(c.CubeConfigs) != c.Count {
		return nil, fmt.Errorf("effects: invalid solid cube train count or materials")
	}
	for _, value := range [...]float64{
		c.PhaseStart, c.PhaseSpacing, c.PhaseIndexOrigin, c.PhaseStep, c.Speed,
		c.X.Amplitude, c.X.Spatial, c.X.Speed, c.X.Phase, c.X.Offset,
		c.Y.Amplitude, c.Y.Spatial, c.Y.Speed, c.Y.Phase, c.Y.Offset,
		c.RotationStart.X, c.RotationStart.Y, c.RotationStart.Z,
		c.RotationSpacing.X, c.RotationSpacing.Y, c.RotationSpacing.Z,
		c.RotationStep.X, c.RotationStep.Y, c.RotationStep.Z,
		c.RotationIndexFactor.X, c.RotationIndexFactor.Y, c.RotationIndexFactor.Z,
	} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("effects: nonfinite solid cube train setting")
		}
	}
	if c.Speed < 0 {
		return nil, fmt.Errorf("effects: negative solid cube train speed")
	}
	speed := c.Speed
	if speed == 0 {
		speed = 1
	}
	train := &SolidCubeTrain{
		config: c, cubes: make([]*SolidCube, c.Count), phases: make([]float64, c.Count),
		batch: NewSolidCubeBatch(c.Count), speed: speed,
	}
	for i := range train.cubes {
		material := c.Cube
		if len(c.CubeConfigs) != 0 {
			material = c.CubeConfigs[i]
		}
		cube, err := NewSolidCube(material)
		if err != nil {
			_ = train.Close()
			return nil, fmt.Errorf("effects: solid cube train item %d: %w", i, err)
		}
		cube.Rotation = geometry.Vec3{
			X: c.RotationStart.X + float64(i)*c.RotationSpacing.X,
			Y: c.RotationStart.Y + float64(i)*c.RotationSpacing.Y,
			Z: c.RotationStart.Z + float64(i)*c.RotationSpacing.Z,
		}
		train.cubes[i] = cube
		train.phases[i] = c.PhaseStart + c.PhaseSpacing*(float64(i)+c.PhaseIndexOrigin)
	}
	return train, nil
}

// SetSpeed changes phase and rotation rates together without resetting pose.
func (t *SolidCubeTrain) SetSpeed(speed float64) error {
	if t == nil || math.IsNaN(speed) || math.IsInf(speed, 0) || speed < 0 {
		return fmt.Errorf("effects: invalid solid cube train speed")
	}
	t.speed = speed
	return nil
}

func (t *SolidCubeTrain) Update(kit.Frame) error {
	if t == nil || t.batch == nil {
		return fmt.Errorf("effects: closed solid cube train")
	}
	if t.speed == 0 {
		return nil
	}
	c := t.config
	for i, cube := range t.cubes {
		t.phases[i] += c.PhaseStep * t.speed
		index := float64(i)
		cube.Rotate(
			c.RotationStep.X*t.speed*(1+index*c.RotationIndexFactor.X),
			c.RotationStep.Y*t.speed*(1+index*c.RotationIndexFactor.Y),
			c.RotationStep.Z*t.speed*(1+index*c.RotationIndexFactor.Z),
		)
	}
	return nil
}

// Pose reports a cube's current screen position and continuous XYZ rotation.
func (t *SolidCubeTrain) Pose(index int) (motion.Point, geometry.Vec3, bool) {
	if t == nil || t.batch == nil || index < 0 || index >= len(t.cubes) {
		return motion.Point{}, geometry.Vec3{}, false
	}
	return t.position(index), t.cubes[index].Rotation, true
}

func (t *SolidCubeTrain) position(index int) motion.Point {
	phase := t.phases[index]
	if t.config.Path != nil {
		return t.config.Path(index, phase)
	}
	return motion.Point{X: t.config.X.At(0, phase), Y: t.config.Y.At(0, phase)}
}

func (t *SolidCubeTrain) Draw(dst *ebiten.Image) {
	if t == nil || t.batch == nil || dst == nil {
		return
	}
	t.batch.Reset()
	for i, cube := range t.cubes {
		position := t.position(i)
		if math.IsNaN(position.X) || math.IsInf(position.X, 0) || math.IsNaN(position.Y) || math.IsInf(position.Y, 0) {
			continue
		}
		t.batch.Add(cube, position.X, position.Y)
	}
	t.batch.Draw(dst)
}

func (t *SolidCubeTrain) Close() error {
	if t == nil {
		return nil
	}
	var closeErr error
	for _, cube := range t.cubes {
		if cube != nil {
			closeErr = errors.Join(closeErr, cube.Close())
		}
	}
	if t.batch != nil {
		closeErr = errors.Join(closeErr, t.batch.Close())
		t.batch = nil
	}
	return closeErr
}
