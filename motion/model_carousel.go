package motion

import (
	"fmt"
	"math"
)

// Vector3 is a graphics-independent XYZ pose for model choreography.
type Vector3 struct{ X, Y, Z float64 }

func (v Vector3) Add(b Vector3) Vector3 { return Vector3{X: v.X + b.X, Y: v.Y + b.Y, Z: v.Z + b.Z} }

// ModelCarouselEntry keeps a model's initial rotation, camera offset and
// per-tick Euler steps independent of its mesh and material.
type ModelCarouselEntry struct {
	Rotation, Camera, RotationStep Vector3
}

type ModelCarouselConfig struct {
	Models                   []ModelCarouselEntry
	Initial                  int
	StartCamera              Vector3
	EntranceStep, EntranceAt float64
	ExitStep, ExitAt         float64
}

// ModelCarouselPose is the pose prepared for the current draw. Rotating is
// sampled before the entrance distance advances, preserving the first held
// rotation frame and a source's handoff tick.
type ModelCarouselPose struct {
	Active, Pending int
	Camera          Vector3
	Distance        float64
	Rotation        Vector3
	Changing        bool
	Rotating        bool
}

// ModelCarousel owns entry, rotation and recessional handoff timing without
// constructing any GPU resource. Draw reads Pose after Step.
type ModelCarousel struct {
	config   ModelCarouselConfig
	active   int
	pending  int
	camera   Vector3
	distance float64
	rotation Vector3
	changing bool
	pose     ModelCarouselPose
}

func NewModelCarousel(c ModelCarouselConfig) (*ModelCarousel, error) {
	if len(c.Models) == 0 || len(c.Models) > 1024 || c.Initial < 0 || c.Initial >= len(c.Models) ||
		!finiteModelVec(c.StartCamera) ||
		!finiteModelValue(c.EntranceStep) || !finiteModelValue(c.EntranceAt) ||
		!finiteModelValue(c.ExitStep) || !finiteModelValue(c.ExitAt) ||
		c.EntranceStep <= 0 || c.EntranceAt < 0 || c.ExitStep <= 0 || c.ExitAt <= 0 {
		return nil, fmt.Errorf("motion: invalid model carousel")
	}
	for _, model := range c.Models {
		if !finiteModelVec(model.Rotation) || !finiteModelVec(model.Camera) || !finiteModelVec(model.RotationStep) {
			return nil, fmt.Errorf("motion: nonfinite model carousel pose")
		}
	}
	c.Models = append([]ModelCarouselEntry(nil), c.Models...)
	m := &ModelCarousel{config: c}
	m.Reset()
	return m, nil
}

func finiteModelValue(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func finiteModelVec(v Vector3) bool {
	return finiteModelValue(v.X) && finiteModelValue(v.Y) && finiteModelValue(v.Z)
}

func (m *ModelCarousel) Reset() {
	m.active, m.pending = m.config.Initial, m.config.Initial
	m.camera, m.distance = m.config.StartCamera, 0
	m.rotation = m.config.Models[m.active].Rotation
	m.changing = false
	m.pose = ModelCarouselPose{Active: m.active, Pending: m.pending, Camera: m.camera,
		Distance: m.distance, Rotation: m.rotation}
}

// Select starts or retargets a recession without changing the current pose.
func (m *ModelCarousel) Select(index int) error {
	if index < 0 || index >= len(m.config.Models) {
		return fmt.Errorf("motion: invalid model carousel selection")
	}
	m.pending, m.changing = index, true
	return nil
}

func (m *ModelCarousel) Step() {
	rotating := m.distance >= m.config.EntranceAt
	if !rotating {
		m.distance += m.config.EntranceStep
	}
	if m.changing {
		if m.camera.Z < m.config.ExitAt {
			m.camera.Z += m.config.ExitStep
		} else {
			m.active = m.pending
			m.distance = 0
			m.camera = m.config.Models[m.active].Camera
			m.rotation = m.config.Models[m.active].Rotation
			m.changing = false
		}
	}
	m.pose = ModelCarouselPose{Active: m.active, Pending: m.pending,
		Camera: m.camera, Distance: m.distance, Rotation: m.rotation,
		Changing: m.changing, Rotating: rotating}
	if rotating {
		m.rotation = m.rotation.Add(m.config.Models[m.active].RotationStep)
	}
}

func (m *ModelCarousel) Pose() ModelCarouselPose { return m.pose }
func (m *ModelCarousel) Count() int              { return len(m.config.Models) }
