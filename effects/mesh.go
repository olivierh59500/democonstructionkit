package effects

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

// Triangle indexes positions and supplies per-corner texture coordinates in pixels.
type Triangle struct {
	Indices [3]int
	UV      [3]geometry.Vec2
	Color   color.NRGBA
}
type Mesh struct {
	Points    []geometry.Vec3
	Triangles []Triangle
}

// Cube provides independent UVs for each face. TextureSize may be (1,1) for solids.
func Cube(size float64, textureSize geometry.Vec2, tint color.NRGBA) Mesh {
	s := size / 2
	m := Mesh{Points: []geometry.Vec3{{-s, -s, -s}, {s, -s, -s}, {s, s, -s}, {-s, s, -s}, {-s, -s, s}, {s, -s, s}, {s, s, s}, {-s, s, s}}}
	uv := [4]geometry.Vec2{{0, 0}, {textureSize.X, 0}, {textureSize.X, textureSize.Y}, {0, textureSize.Y}}
	for _, face := range [][4]int{{0, 3, 2, 1}, {4, 5, 6, 7}, {0, 4, 7, 3}, {1, 2, 6, 5}, {0, 1, 5, 4}, {3, 7, 6, 2}} {
		m.Triangles = append(m.Triangles, Triangle{Indices: [3]int{face[0], face[1], face[2]}, UV: [3]geometry.Vec2{uv[0], uv[1], uv[2]}, Color: tint}, Triangle{Indices: [3]int{face[0], face[2], face[3]}, UV: [3]geometry.Vec2{uv[0], uv[2], uv[3]}, Color: tint})
	}
	return m
}

type Transform struct {
	Position, Rotation geometry.Vec3
	Scale              float64
}

// MeshEffect draws depth-sorted triangles, clipping rather than discarding near
// plane intersections. Texture interpolation is affine; subdivide large textured
// surfaces for perspective effects. Transparent materials enable Glenz rendering.
type MeshEffect struct {
	State
	Mesh          Mesh
	Camera        geometry.Camera
	Transform     Transform
	Animate       func(float64) Transform
	Deform        func(index int, p geometry.Vec3, seconds float64) geometry.Vec3
	CullBackFaces bool
	Light         geometry.Vec3
	Ambient       float64
	Blend         ebiten.Blend
	texture       *ebiten.Image
	owned         bool
	points        []geometry.Vec3
	faces         []meshFace
	batch         *render.Batch
}
type meshFace struct {
	triangle [3]geometry.Vertex
	color    color.NRGBA
	depth    float64
}

func NewMesh(mesh Mesh, texture *ebiten.Image, camera geometry.Camera) (*MeshEffect, error) {
	if camera.Near <= 0 || camera.Focal <= 0 || len(mesh.Points) == 0 {
		return nil, fmt.Errorf("effects: invalid mesh camera or points")
	}
	for _, t := range mesh.Triangles {
		for _, i := range t.Indices {
			if i < 0 || i >= len(mesh.Points) {
				return nil, fmt.Errorf("effects: invalid mesh index")
			}
		}
	}
	m := &MeshEffect{Mesh: Mesh{Points: append([]geometry.Vec3(nil), mesh.Points...), Triangles: append([]Triangle(nil), mesh.Triangles...)}, Camera: camera, Transform: Transform{Scale: 1}, texture: texture, points: make([]geometry.Vec3, len(mesh.Points)), faces: make([]meshFace, 0, len(mesh.Triangles)), batch: render.NewBatch(2048), Ambient: 1}
	if texture == nil {
		m.texture = ebiten.NewImage(1, 1)
		m.texture.Fill(color.White)
		m.owned = true
	}
	return m, nil
}
func (m *MeshEffect) Update(f kit.Frame) error {
	m.Frame = f
	p := m.Transform
	if m.Animate != nil {
		p = m.Animate(f.Time)
	}
	rot := geometry.RotateXYZ(p.Rotation)
	for i, v := range m.Mesh.Points {
		if m.Deform != nil {
			v = m.Deform(i, v, f.Time)
		}
		m.points[i] = rot.Apply(v.Scale(p.Scale)).Add(p.Position)
	}
	m.faces = m.faces[:0]
	light := m.Light.Unit()
	for _, t := range m.Mesh.Triangles {
		a, b, c := m.points[t.Indices[0]], m.points[t.Indices[1]], m.points[t.Indices[2]]
		normal := b.Sub(a).Cross(c.Sub(a)).Unit()
		if m.CullBackFaces && normal.Dot(a) >= 0 {
			continue
		}
		clr := t.Color
		intensity := math.Max(0, math.Min(1, m.Ambient+(1-m.Ambient)*math.Max(0, normal.Dot(light))))
		clr.R = uint8(float64(clr.R) * intensity)
		clr.G = uint8(float64(clr.G) * intensity)
		clr.B = uint8(float64(clr.B) * intensity)
		m.faces = append(m.faces, meshFace{triangle: [3]geometry.Vertex{{Position: a, UV: t.UV[0]}, {Position: b, UV: t.UV[1]}, {Position: c, UV: t.UV[2]}}, color: clr, depth: (a.Z + b.Z + c.Z) / 3})
	}
	sort.SliceStable(m.faces, func(i, j int) bool { return m.faces[i].depth > m.faces[j].depth })
	return nil
}
func (m *MeshEffect) Draw(dst *ebiten.Image) {
	m.batch.Options.Blend = m.Blend
	m.batch.Begin(dst, m.texture)
	var clipped [4]geometry.Vertex
	for _, f := range m.faces {
		polygon := geometry.ClipNear(clipped[:0], f.triangle, m.Camera.Near)
		for j := 1; j+1 < len(polygon); j++ {
			var v [3]ebiten.Vertex
			for i, p := range [3]geometry.Vertex{polygon[0], polygon[j], polygon[j+1]} {
				pos, _, _ := m.Camera.Project(p.Position)
				v[i] = render.Vertex(pos.X, pos.Y, p.UV.X, p.UV.Y, f.color)
			}
			m.batch.Triangle(v[0], v[1], v[2])
		}
	}
	m.batch.Flush()
}
func (m *MeshEffect) Close() error {
	if m.owned {
		m.texture.Deallocate()
	}
	return nil
}
