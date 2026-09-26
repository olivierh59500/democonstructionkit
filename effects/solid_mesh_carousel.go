package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
)

// SolidFace is a triangle when Indices[3] is negative and a quad otherwise.
// Color is a packed 0xRRGGBB value; every face retains its own material.
type SolidFace struct {
	Indices [4]int
	Color   uint32
}

// SolidMesh converts grouped solid faces without changing their input order.
func SolidMesh(points []geometry.Vec3, faces []SolidFace) Mesh {
	mesh := Mesh{Points: points, Triangles: make([]Triangle, 0, len(faces)*2)}
	for _, face := range faces {
		clr := face.Color
		shade := color.NRGBA{R: uint8(clr >> 16), G: uint8(clr >> 8), B: uint8(clr), A: 255}
		index := face.Indices
		mesh.Triangles = append(mesh.Triangles, Triangle{Indices: [3]int{index[0], index[1], index[2]}, Color: shade})
		if index[3] >= 0 {
			mesh.Triangles = append(mesh.Triangles, Triangle{Indices: [3]int{index[0], index[2], index[3]}, Color: shade})
		}
	}
	return mesh
}

// SolidMeshModel leaves points, face groups and colors as production data.
type SolidMeshModel struct {
	Points []geometry.Vec3
	Groups [][]SolidFace
}

type SolidMeshCarouselConfig struct {
	Models        []SolidMeshModel
	Motion        motion.ModelCarouselConfig
	Camera        geometry.Camera
	RotationOrder [3]uint8 // 0=X, 1=Y, 2=Z; first entry is applied first.
	Mirror        geometry.Vec3
	CameraSign    geometry.Vec3
	CullBackFaces bool
}

// SolidMeshCarousel owns its pure selection clock and all grouped MeshEffects.
// Each group keeps its own face order while sharing the current rotation pose.
type SolidMeshCarousel struct {
	config   SolidMeshCarouselConfig
	motion   *motion.ModelCarousel
	models   [][]*MeshEffect
	rotation *geometry.OrderedEuler
	source   *ebiten.Image
}

func NewSolidMeshCarousel(c SolidMeshCarouselConfig) (*SolidMeshCarousel, error) {
	if len(c.Models) == 0 || len(c.Models) != len(c.Motion.Models) || !finiteSolidVec(c.CameraSign) {
		return nil, fmt.Errorf("effects: invalid solid mesh carousel")
	}
	rotation, err := geometry.NewOrderedEuler(c.RotationOrder, c.Mirror)
	if err != nil {
		return nil, err
	}
	clock, err := motion.NewModelCarousel(c.Motion)
	if err != nil {
		return nil, err
	}
	result := &SolidMeshCarousel{config: c, motion: clock, rotation: rotation,
		models: make([][]*MeshEffect, len(c.Models))}
	result.source = ebiten.NewImage(1, 1)
	result.source.Fill(color.White)
	for i, model := range c.Models {
		if len(model.Points) == 0 || len(model.Groups) == 0 {
			result.Close()
			return nil, fmt.Errorf("effects: empty solid model %d", i)
		}
		for _, group := range model.Groups {
			if len(group) == 0 {
				result.Close()
				return nil, fmt.Errorf("effects: empty solid model face group")
			}
			mesh, err := NewMesh(SolidMesh(model.Points, group), result.source, c.Camera)
			if err != nil {
				result.Close()
				return nil, err
			}
			mesh.CullBackFaces = c.CullBackFaces
			mesh.Deform = result.deform
			result.models[i] = append(result.models[i], mesh)
		}
	}
	return result, nil
}

func finiteSolidVec(v geometry.Vec3) bool {
	return !math.IsNaN(v.X) && !math.IsInf(v.X, 0) &&
		!math.IsNaN(v.Y) && !math.IsInf(v.Y, 0) &&
		!math.IsNaN(v.Z) && !math.IsInf(v.Z, 0)
}

func (c *SolidMeshCarousel) deform(_ int, p geometry.Vec3, _ float64) geometry.Vec3 {
	return c.rotation.Apply(p)
}

// Step prepares only the active model's mesh groups. Draw does not move time.
func (c *SolidMeshCarousel) Step() error {
	c.motion.Step()
	pose := c.motion.Pose()
	c.rotation.SetAngles(geometry.Vec3{X: pose.Rotation.X, Y: pose.Rotation.Y, Z: pose.Rotation.Z})
	position := geometry.Vec3{
		X: pose.Camera.X * c.config.CameraSign.X,
		Y: pose.Camera.Y * c.config.CameraSign.Y,
		Z: pose.Camera.Z*c.config.CameraSign.Z + pose.Distance,
	}
	for _, mesh := range c.models[pose.Active] {
		mesh.Transform.Position = position
		if err := mesh.Update(kit.Frame{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *SolidMeshCarousel) Update(kit.Frame) error { return c.Step() }

func (c *SolidMeshCarousel) Draw(dst *ebiten.Image) {
	if c == nil || dst == nil {
		return
	}
	for _, mesh := range c.models[c.motion.Pose().Active] {
		mesh.Draw(dst)
	}
}

func (c *SolidMeshCarousel) Select(index int) error        { return c.motion.Select(index) }
func (c *SolidMeshCarousel) Motion() *motion.ModelCarousel { return c.motion }

func (c *SolidMeshCarousel) Close() error {
	if c == nil {
		return nil
	}
	var first error
	for _, group := range c.models {
		for _, mesh := range group {
			if err := mesh.Close(); err != nil && first == nil {
				first = err
			}
		}
	}
	if c.source != nil {
		c.source.Deallocate()
		c.source = nil
	}
	return first
}
