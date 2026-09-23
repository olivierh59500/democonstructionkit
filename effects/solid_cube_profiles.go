package effects

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

func multiplyCubeColors(vertices []ebiten.Vertex, clr color.RGBA) {
	for i := range vertices {
		v := &vertices[i]
		v.ColorR = float32(clr.R) * (1.0 / 255)
		v.ColorG = float32(clr.G) * (1.0 / 255)
		v.ColorB = float32(clr.B) * (1.0 / 255)
		v.ColorA = float32(clr.A) * (1.0 / 255)
	}
}

func appendPreciseCubeLine(vertices []ebiten.Vertex, indices []uint16, from, to [2]float64, width float64, clr color.RGBA) ([]ebiten.Vertex, []uint16) {
	dx, dy := to[0]-from[0], to[1]-from[1]
	length := math.Hypot(dx, dy)
	if length == 0 {
		return vertices, indices
	}
	ox, oy := -dy/length/2*width, dx/length/2*width
	points := [4]point2D{{float32(from[0] + ox), float32(from[1] + oy)}, {float32(to[0] + ox), float32(to[1] + oy)}, {float32(from[0] - ox), float32(from[1] - oy)}, {float32(to[0] - ox), float32(to[1] - oy)}}
	return appendColoredQuad(vertices, indices, points, clr)
}
func appendSquaredCubeLine(vertices []ebiten.Vertex, indices []uint16, from, to point2D, width float32, clr color.RGBA) ([]ebiten.Vertex, []uint16) {
	dx, dy := to.x-from.x, to.y-from.y
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if length == 0 {
		return vertices, indices
	}
	half := width / (2 * length)
	ox, oy := -dy*half, dx*half
	return appendColoredQuad(vertices, indices, [4]point2D{{from.x + ox, from.y + oy}, {to.x + ox, to.y + oy}, {to.x - ox, to.y - oy}, {from.x - ox, from.y - oy}}, clr)
}

// SolidCubeBatch accumulates independently configured cubes and submits them in
// one draw call. Capacity is fixed; Add reports overflow instead of allocating.
// Reset preserves all buffers. Each cube remains caller-owned.
type SolidCubeBatch struct {
	vertices []ebiten.Vertex
	indices  []uint16
	texture  *ebiten.Image
}

func NewSolidCubeBatch(capacity int) *SolidCubeBatch {
	capacity = max(1, min(capacity, 65535/(cubeFaceCount*cubeVerticesPerFace)))
	b := &SolidCubeBatch{vertices: make([]ebiten.Vertex, 0, capacity*cubeFaceCount*cubeVerticesPerFace), indices: make([]uint16, 0, capacity*cubeFaceCount*cubeIndicesPerFace), texture: ebiten.NewImage(3, 3)}
	b.texture.Fill(color.White)
	return b
}
func (b *SolidCubeBatch) Reset() { b.vertices = b.vertices[:0]; b.indices = b.indices[:0] }
func (b *SolidCubeBatch) Add(cube *SolidCube, x, y float64) bool {
	if cube == nil || b.texture == nil {
		return false
	}
	vertices, indices := cube.Geometry(x, y)
	if len(b.vertices)+len(vertices) > cap(b.vertices) || len(b.indices)+len(indices) > cap(b.indices) {
		return false
	}
	base := uint16(len(b.vertices))
	b.vertices = append(b.vertices, vertices...)
	for _, i := range indices {
		b.indices = append(b.indices, base+i)
	}
	return true
}
func (b *SolidCubeBatch) Draw(dst *ebiten.Image) {
	if dst != nil && b.texture != nil && len(b.indices) > 0 {
		dst.DrawTriangles(b.vertices, b.indices, b.texture, nil)
	}
}
func (b *SolidCubeBatch) Close() error {
	if b.texture != nil {
		b.texture.Deallocate()
		b.texture = nil
	}
	return nil
}
