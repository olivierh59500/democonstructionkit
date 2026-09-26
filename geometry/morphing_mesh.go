package geometry

import (
	"cmp"
	"fmt"
	"math"
	"slices"
)

// MorphFace keeps topology and a renderer-defined material index independent
// of image, palette and blend choices.
type MorphFace struct {
	Vertices [3]int
	Material int
}

type ProjectedPoint struct{ X, Y float32 }
type SortedMorphFace struct {
	Index int
	Depth float64
}

// MorphingMeshConfig describes equal-sized shape targets, cyclic interpolation,
// sequential X/Y/Z rotation, depth sorting and a perspective camera. A target
// becomes the next source when CycleTicks elapse; MorphTicks controls how long
// the interpolation takes before the hold. A zero RotationPeriod selects 2π.
type MorphingMeshConfig struct {
	Shapes         [][]Vec3
	Faces          []MorphFace
	StartShape     int
	MorphTicks     float64
	CycleTicks     float64
	RotationStart  Vec3
	RotationStep   Vec3
	RotationPeriod float64
	CenterX        float32
	CenterY        float32
	Focal          float64
	CameraZ        float64
}

// MorphingMesh owns all animation and projected geometry without a graphics
// backend. Borrowed sample slices are valid until the next Step or Reset.
type MorphingMesh struct {
	config      MorphingMeshConfig
	current     []Vec3
	transformed []Vec3
	projected   []ProjectedPoint
	faces       []SortedMorphFace
	rotation    Vec3
	timer       float64
	shape       int
	target      int
	tick        int
}

func NewMorphingMesh(c MorphingMeshConfig) (*MorphingMesh, error) {
	if len(c.Shapes) == 0 || len(c.Shapes[0]) == 0 || len(c.Faces) == 0 ||
		c.StartShape < 0 || c.StartShape >= len(c.Shapes) ||
		!finiteMesh(c.MorphTicks) || !finiteMesh(c.CycleTicks) ||
		c.MorphTicks <= 0 || c.CycleTicks <= c.MorphTicks ||
		!finiteMesh(c.Focal) || c.Focal <= 0 || !finiteMesh(c.CameraZ) ||
		!finiteMesh(c.RotationStart.X) || !finiteMesh(c.RotationStart.Y) || !finiteMesh(c.RotationStart.Z) ||
		!finiteMesh(c.RotationStep.X) || !finiteMesh(c.RotationStep.Y) || !finiteMesh(c.RotationStep.Z) ||
		math.IsNaN(float64(c.CenterX)) || math.IsInf(float64(c.CenterX), 0) ||
		math.IsNaN(float64(c.CenterY)) || math.IsInf(float64(c.CenterY), 0) {
		return nil, fmt.Errorf("geometry: invalid morphing mesh")
	}
	if c.RotationPeriod == 0 {
		c.RotationPeriod = 2 * math.Pi
	}
	if !finiteMesh(c.RotationPeriod) || c.RotationPeriod <= 0 {
		return nil, fmt.Errorf("geometry: invalid rotation period")
	}
	n := len(c.Shapes[0])
	shapes := make([][]Vec3, len(c.Shapes))
	for i, shape := range c.Shapes {
		if len(shape) != n {
			return nil, fmt.Errorf("geometry: unequal morph shape %d", i)
		}
		shapes[i] = append([]Vec3(nil), shape...)
		for _, v := range shape {
			if !finiteMesh(v.X) || !finiteMesh(v.Y) || !finiteMesh(v.Z) {
				return nil, fmt.Errorf("geometry: invalid morph vertex")
			}
		}
	}
	for _, face := range c.Faces {
		for _, index := range face.Vertices {
			if index < 0 || index >= n {
				return nil, fmt.Errorf("geometry: invalid morph face index")
			}
		}
		if face.Material < 0 {
			return nil, fmt.Errorf("geometry: invalid morph material index")
		}
	}
	c.Shapes, c.Faces = shapes, append([]MorphFace(nil), c.Faces...)
	m := &MorphingMesh{config: c, current: make([]Vec3, n), transformed: make([]Vec3, n),
		projected: make([]ProjectedPoint, n), faces: make([]SortedMorphFace, len(c.Faces))}
	m.Reset()
	return m, nil
}

func finiteMesh(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (m *MorphingMesh) Reset() {
	m.shape = m.config.StartShape
	m.target = (m.shape + 1) % len(m.config.Shapes)
	m.rotation = m.config.RotationStart
	m.timer, m.tick = 0, 0
	copy(m.current, m.config.Shapes[m.shape])
	clear(m.transformed)
	clear(m.projected)
	clear(m.faces)
}

// Step advances rotation and morph before preparing sorted, projected faces.
// The X, Y and Z rotations are applied in that order without collapsing them
// into a matrix, which preserves the arithmetic of authored meshes.
func (m *MorphingMesh) Step() {
	c := m.config
	m.tick++
	m.rotation.X += c.RotationStep.X
	m.rotation.Y += c.RotationStep.Y
	m.rotation.Z += c.RotationStep.Z
	if m.rotation.X >= c.RotationPeriod {
		m.rotation.X -= c.RotationPeriod
	}
	if m.rotation.Y >= c.RotationPeriod {
		m.rotation.Y -= c.RotationPeriod
	}
	if m.rotation.Z >= c.RotationPeriod {
		m.rotation.Z -= c.RotationPeriod
	}
	m.timer++
	if m.timer >= c.CycleTicks {
		m.timer = 0
		m.shape = m.target
		m.target = (m.target + 1) % len(c.Shapes)
	}
	if m.timer <= c.MorphTicks {
		t := m.timer / c.MorphTicks
		for i := range m.current {
			from, to := c.Shapes[m.shape][i], c.Shapes[m.target][i]
			m.current[i] = Vec3{
				X: from.X + (to.X-from.X)*t,
				Y: from.Y + (to.Y-from.Y)*t,
				Z: from.Z + (to.Z-from.Z)*t,
			}
		}
	}
	cosX, sinX := math.Cos(m.rotation.X), math.Sin(m.rotation.X)
	cosY, sinY := math.Cos(m.rotation.Y), math.Sin(m.rotation.Y)
	cosZ, sinZ := math.Cos(m.rotation.Z), math.Sin(m.rotation.Z)
	for i, v := range m.current {
		x, y, z := v.X, v.Y, v.Z
		newY := y*cosX - z*sinX
		newZ := y*sinX + z*cosX
		y, z = newY, newZ
		newX := x*cosY + z*sinY
		newZ = -x*sinY + z*cosY
		x, z = newX, newZ
		newX = x*cosZ - y*sinZ
		newY = x*sinZ + y*cosZ
		x, y = newX, newY
		m.transformed[i] = Vec3{X: x, Y: y, Z: z}
	}
	for i, face := range c.Faces {
		v := face.Vertices
		m.faces[i] = SortedMorphFace{Index: i,
			Depth: (m.transformed[v[0]].Z + m.transformed[v[1]].Z + m.transformed[v[2]].Z) / 3.0}
	}
	slices.SortFunc(m.faces, func(a, b SortedMorphFace) int { return cmp.Compare(a.Depth, b.Depth) })
	for i, v := range m.transformed {
		z := v.Z + c.CameraZ
		if z <= 0 {
			z = 1
		}
		scale := float32(c.Focal / z)
		m.projected[i] = ProjectedPoint{X: float32(v.X)*scale + c.CenterX,
			Y: float32(v.Y)*scale + c.CenterY}
	}
}

func (m *MorphingMesh) Current() []Vec3                { return m.current }
func (m *MorphingMesh) Transformed() []Vec3            { return m.transformed }
func (m *MorphingMesh) Projected() []ProjectedPoint    { return m.projected }
func (m *MorphingMesh) SortedFaces() []SortedMorphFace { return m.faces }
func (m *MorphingMesh) Faces() []MorphFace             { return m.config.Faces }
func (m *MorphingMesh) Rotation() Vec3                 { return m.rotation }
func (m *MorphingMesh) Shape() (current, target int)   { return m.shape, m.target }
func (m *MorphingMesh) Timer() float64                 { return m.timer }
func (m *MorphingMesh) Tick() int                      { return m.tick }
