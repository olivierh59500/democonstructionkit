package effects

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// MorphMaterial keeps each face's tint and blend independent of the mesh's
// geometry. An unset Blend uses source-over.
type MorphMaterial struct {
	Color color.RGBA
	Blend ebiten.Blend
}

// MorphingMeshConfig binds a pure cyclic mesh to per-face solid materials.
// CullPositive and CullNegative test projected triangle winding; both false
// retain both sides. Source is a borrowed white image, or nil for an owned one.
type MorphingMeshConfig struct {
	Geometry     geometry.MorphingMeshConfig
	Materials    []MorphMaterial
	CullPositive bool
	CullNegative bool
	Filter       ebiten.Filter
	Source       *ebiten.Image
}

// MorphingMesh renders sorted faces in order. Each face is drawn separately so
// overlapping transparent and opaque materials keep their authored blending.
type MorphingMesh struct {
	geometry  *geometry.MorphingMesh
	materials []MorphMaterial
	cull      int
	source    *ebiten.Image
	owns      bool
	vertices  [3]ebiten.Vertex
	indices   [3]uint16
	options   ebiten.DrawTrianglesOptions
}

func NewMorphingMesh(c MorphingMeshConfig) (*MorphingMesh, error) {
	if len(c.Materials) == 0 || c.CullPositive && c.CullNegative {
		return nil, fmt.Errorf("effects: invalid morphing mesh materials or culling")
	}
	shape, err := geometry.NewMorphingMesh(c.Geometry)
	if err != nil {
		return nil, err
	}
	for _, face := range shape.Faces() {
		if face.Material >= len(c.Materials) {
			return nil, fmt.Errorf("effects: morph face material out of range")
		}
	}
	m := &MorphingMesh{
		geometry: shape, materials: append([]MorphMaterial(nil), c.Materials...),
		source: c.Source, indices: [3]uint16{0, 1, 2},
	}
	for i := range m.materials {
		if m.materials[i].Blend == (ebiten.Blend{}) {
			m.materials[i].Blend = ebiten.BlendSourceOver
		}
	}
	if c.CullPositive {
		m.cull = 1
	} else if c.CullNegative {
		m.cull = -1
	}
	if m.source == nil {
		m.source = ebiten.NewImage(1, 1)
		m.source.Fill(color.White)
		m.owns = true
	}
	m.options.Filter = c.Filter
	return m, nil
}

func (m *MorphingMesh) Update(kit.Frame) error {
	m.geometry.Step()
	return nil
}

func (m *MorphingMesh) Draw(dst *ebiten.Image) {
	if m == nil || dst == nil || m.geometry.Tick() == 0 {
		return
	}
	projected := m.geometry.Projected()
	faces := m.geometry.Faces()
	for _, sample := range m.geometry.SortedFaces() {
		face := faces[sample.Index]
		p0, p1, p2 := projected[face.Vertices[0]], projected[face.Vertices[1]], projected[face.Vertices[2]]
		cross := (p1.X-p0.X)*(p2.Y-p0.Y) - (p1.Y-p0.Y)*(p2.X-p0.X)
		if m.cull > 0 && cross > 0 || m.cull < 0 && cross < 0 {
			continue
		}
		material := m.materials[face.Material]
		red := float32(material.Color.R) / 0xff
		green := float32(material.Color.G) / 0xff
		blue := float32(material.Color.B) / 0xff
		alpha := float32(material.Color.A) / 0xff
		m.vertices[0] = ebiten.Vertex{DstX: p0.X, DstY: p0.Y, SrcX: 0, SrcY: 0,
			ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha}
		m.vertices[1] = ebiten.Vertex{DstX: p1.X, DstY: p1.Y, SrcX: 1, SrcY: 0,
			ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha}
		m.vertices[2] = ebiten.Vertex{DstX: p2.X, DstY: p2.Y, SrcX: 0, SrcY: 1,
			ColorR: red, ColorG: green, ColorB: blue, ColorA: alpha}
		m.options.Blend = material.Blend
		dst.DrawTriangles(m.vertices[:], m.indices[:], m.source, &m.options)
	}
}

// Geometry exposes live projected samples and shape state for cues and tooling.
func (m *MorphingMesh) Geometry() *geometry.MorphingMesh { return m.geometry }

// Close releases only an internally created source image.
func (m *MorphingMesh) Close() {
	if m != nil && m.owns && m.source != nil {
		m.source.Deallocate()
		m.source = nil
	}
}
