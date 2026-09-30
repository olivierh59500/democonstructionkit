package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

// MeshOutlineConfig describes polygon boundaries over an existing mesh.
// Empty Faces uses its triangles; CubeFaces supplies undivided cube quads.
// Width defaults to one and Color to white. White optionally borrows a one-
// pixel material; a solid mesh reuses its own fallback when none is supplied.
// Polygon order is retained and shared edges are drawn once per incident face.
type MeshOutlineConfig struct {
	Faces [][]int
	Width float64
	Color color.Color
	White *ebiten.Image
	Blend ebiten.Blend
}

type meshOutline struct {
	faces     [][]int
	points    []int
	projected []geometry.Vec2
	visible   []bool
	width     float64
	paint     color.Color
	white     *ebiten.Image
	owned     bool
	blend     ebiten.Blend
}

// SetOutline prepares an independently styled outline without resetting the
// mesh's current pose. Faces are copied and bounded to 65,536 total corners.
// DrawOutline uses the same transformed/deformed points as filled drawing.
func (m *MeshEffect) SetOutline(c MeshOutlineConfig) error {
	if m == nil || m.closed || c.Width < 0 || math.IsNaN(c.Width) || math.IsInf(c.Width, 0) {
		return fmt.Errorf("effects: invalid mesh outline target or width")
	}
	if c.White != nil && (c.White.Bounds().Min.X != 0 || c.White.Bounds().Min.Y != 0 || c.White.Bounds().Dx() != 1 || c.White.Bounds().Dy() != 1) {
		return fmt.Errorf("effects: mesh outline needs a one-pixel white material")
	}
	faces := c.Faces
	if len(faces) == 0 {
		if len(m.Mesh.Triangles) == 0 || len(m.Mesh.Triangles) > 65536/3 {
			return fmt.Errorf("effects: mesh outline needs bounded polygon faces")
		}
		faces = make([][]int, len(m.Mesh.Triangles))
		for i, triangle := range m.Mesh.Triangles {
			face := triangle.Indices
			faces[i] = face[:]
		}
	}
	if len(faces) > 65536/3 {
		return fmt.Errorf("effects: too many mesh outline faces")
	}
	corners := 0
	for _, face := range faces {
		if len(face) < 3 || len(face) > 65536-corners {
			return fmt.Errorf("effects: invalid mesh outline face size")
		}
		corners += len(face)
		for _, index := range face {
			if index < 0 || index >= len(m.points) {
				return fmt.Errorf("effects: invalid mesh outline point %d", index)
			}
		}
	}
	if c.Width == 0 {
		c.Width = 1
	}
	if c.Color == nil {
		c.Color = color.White
	}
	o := &meshOutline{width: c.Width, paint: c.Color, blend: c.Blend, white: c.White, faces: make([][]int, len(faces))}
	// Compact the referenced point set once; unused model points need no outline
	// projection or per-draw scratch reset.
	slots := make(map[int]int, min(corners, len(m.points)))
	for i, face := range faces {
		o.faces[i] = make([]int, len(face))
		for j, index := range face {
			slot, ok := slots[index]
			if !ok {
				slot = len(o.points)
				slots[index] = slot
				o.points = append(o.points, index)
			}
			o.faces[i][j] = slot
		}
	}
	o.projected = make([]geometry.Vec2, len(o.points))
	o.visible = make([]bool, len(o.points))
	if o.white == nil {
		if m.owned {
			o.white = m.texture
		} else {
			o.white = ebiten.NewImage(1, 1)
			o.white.Fill(color.White)
			o.owned = true
		}
	}
	if m.outline != nil {
		m.outline.close()
	}
	m.outline = o
	return nil
}

// DrawOutline projects each referenced point once and draws ordered polygon
// edges. Back-face culling follows CullBackFaces; near-plane intersections are
// clipped per segment. Draw does not update animation or allocate point slices.
func (m *MeshEffect) DrawOutline(dst *ebiten.Image) {
	if m == nil || m.closed || m.outline == nil || dst == nil {
		return
	}
	o := m.outline
	for slot, index := range o.points {
		o.projected[slot], _, o.visible[slot] = m.Camera.Project(m.points[index])
	}
	m.batch.Options.Blend = o.blend
	m.batch.Begin(dst, o.white)
	for _, face := range o.faces {
		a, b, c := m.points[o.points[face[0]]], m.points[o.points[face[1]]], m.points[o.points[face[2]]]
		if m.CullBackFaces && b.Sub(a).Cross(c.Sub(a)).Dot(a) >= 0 {
			continue
		}
		for i, start := range face {
			end := face[(i+1)%len(face)]
			segment := [2]geometry.Vec2{o.projected[start], o.projected[end]}
			if !o.visible[start] || !o.visible[end] {
				first, second, visible := geometry.ClipNearSegment(m.points[o.points[start]], m.points[o.points[end]], m.Camera.Near)
				if !visible {
					continue
				}
				var firstOK, secondOK bool
				segment[0], _, firstOK = m.Camera.Project(first)
				segment[1], _, secondOK = m.Camera.Project(second)
				if !firstOK || !secondOK {
					continue
				}
			}
			m.batch.StrokePath(segment[:], render.PathStroke{Width: o.width, Open: true}, o.paint)
		}
	}
	m.batch.Flush()
}

func (o *meshOutline) close() {
	if o.owned && o.white != nil {
		o.white.Deallocate()
	}
	o.white = nil
}
