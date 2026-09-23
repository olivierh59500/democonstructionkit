package effects

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// SolidCubeConfig supplies an outlined, flat-colored cube. Face order is back,
// front, bottom, top, left, right. EdgeWidth=0 disables outlines. Perspective
// places the camera that far from the cube's center and must exceed its radius.
// FarToNear uses the usual painter order; false retains near-first artwork.
// For textured or deformable arbitrary meshes, use NewMesh instead.
type SolidCubeConfig struct {
	Size, Perspective      float64
	FaceColors, EdgeColors [6]color.RGBA
	EdgeWidth              float32
	FarToNear              bool
}

// DefaultSolidCubeConfig returns an opaque white cube with dark gray outlines.
func DefaultSolidCubeConfig(size float64) SolidCubeConfig {
	c := SolidCubeConfig{Size: size, Perspective: max(200, size*2), EdgeWidth: 1, FarToNear: true}
	for i := range c.FaceColors {
		c.FaceColors[i] = color.RGBA{R: 255, G: 255, B: 255, A: 255}
		c.EdgeColors[i] = color.RGBA{R: 64, G: 64, B: 64, A: 255}
	}
	return c
}

// SolidCube caches all geometry buffers and its white texture. Geometry
// allocates no Go storage after construction. Rotation uses XYZ radians;
// keep it continuous between scenes instead of recreating a cube at transitions.
type SolidCube struct {
	Rotation     geometry.Vec3
	config       SolidCubeConfig
	texture      *ebiten.Image
	drawVertices [cubeFaceCount * cubeVerticesPerFace]ebiten.Vertex
	drawIndices  [cubeFaceCount * cubeIndicesPerFace]uint16
}

const (
	cubeFaceCount       = 6
	cubeVerticesPerFace = 20
	cubeIndicesPerFace  = 30
)

var cubeFaces = [cubeFaceCount][4]int{
	{0, 1, 2, 3}, {4, 5, 6, 7}, {0, 1, 5, 4}, {2, 3, 7, 6}, {0, 3, 7, 4}, {1, 2, 6, 5},
}

func NewSolidCube(c SolidCubeConfig) (*SolidCube, error) {
	finite := func(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
	if !finite(c.Size) || !finite(c.Perspective) || !finite(float64(c.EdgeWidth)) || c.Size <= 0 || c.Perspective <= math.Sqrt(3)*c.Size/2 || c.EdgeWidth < 0 {
		return nil, fmt.Errorf("effects: invalid solid cube configuration")
	}
	cube := &SolidCube{config: c, texture: ebiten.NewImage(3, 3)}
	cube.texture.Fill(color.White)
	return cube, nil
}

// Rotate advances caller-controlled angles without changing position or size.
func (c *SolidCube) Rotate(dx, dy, dz float64) {
	c.Rotation.X += dx
	c.Rotation.Y += dy
	c.Rotation.Z += dz
}

func (c *SolidCube) DrawAt(dst *ebiten.Image, centerX, centerY float64) {
	if dst == nil || c.texture == nil {
		return
	}
	vertices, indices := c.Geometry(centerX, centerY)
	dst.DrawTriangles(vertices, indices, c.texture, &ebiten.DrawTrianglesOptions{ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
}
func (c *SolidCube) Close() error {
	if c.texture != nil {
		c.texture.Deallocate()
		c.texture = nil
	}
	return nil
}

type point2D struct {
	x float32
	y float32
}

type faceDepth struct {
	index int
	depth float64
}

// Geometry returns borrowed render buffers, valid until the next call. Each
// face is filled and outlined together, so equal-depth ordering stays stable.
func (c *SolidCube) Geometry(centerX, centerY float64) ([]ebiten.Vertex, []uint16) {
	vertices := [8][3]float64{
		{-c.config.Size / 2, -c.config.Size / 2, -c.config.Size / 2}, // 0
		{c.config.Size / 2, -c.config.Size / 2, -c.config.Size / 2},  // 1
		{c.config.Size / 2, c.config.Size / 2, -c.config.Size / 2},   // 2
		{-c.config.Size / 2, c.config.Size / 2, -c.config.Size / 2},  // 3
		{-c.config.Size / 2, -c.config.Size / 2, c.config.Size / 2},  // 4
		{c.config.Size / 2, -c.config.Size / 2, c.config.Size / 2},   // 5
		{c.config.Size / 2, c.config.Size / 2, c.config.Size / 2},    // 6
		{-c.config.Size / 2, c.config.Size / 2, c.config.Size / 2},   // 7
	}

	var rotated [8][3]float64
	var projected [8]point2D
	sinX, cosX := math.Sincos(c.Rotation.X)
	sinY, cosY := math.Sincos(c.Rotation.Y)
	sinZ, cosZ := math.Sincos(c.Rotation.Z)
	for i, v := range vertices {
		x, y, z := v[0], v[1], v[2]

		y1 := y*cosX - z*sinX
		z1 := y*sinX + z*cosX
		y, z = y1, z1

		x1 := x*cosY + z*sinY
		z2 := -x*sinY + z*cosY
		x, z = x1, z2

		x2 := x*cosZ - y*sinZ
		y2 := x*sinZ + y*cosZ
		x, y = x2, y2

		rotated[i] = [3]float64{x, y, z}
		factor := c.config.Perspective / (c.config.Perspective + z)
		x2d, y2d := x*factor, y*factor
		projected[i] = point2D{float32(centerX + x2d), float32(centerY + y2d)}
	}

	var depths [cubeFaceCount]faceDepth
	for i, face := range cubeFaces {
		centerZ := 0.0
		for _, vi := range face {
			centerZ += rotated[vi][2]
		}
		depths[i] = faceDepth{index: i, depth: centerZ / 4}
	}

	for i := 0; i < len(depths)-1; i++ {
		for j := i + 1; j < len(depths); j++ {
			if (!c.config.FarToNear && depths[i].depth > depths[j].depth) || (c.config.FarToNear && depths[i].depth < depths[j].depth) {
				depths[i], depths[j] = depths[j], depths[i]
			}
		}
	}

	drawVertices := c.drawVertices[:0]
	drawIndices := c.drawIndices[:0]
	for _, fd := range depths {
		face := cubeFaces[fd.index]
		faceColor := c.config.FaceColors[fd.index]
		points := [4]point2D{
			projected[face[0]], projected[face[1]], projected[face[2]], projected[face[3]],
		}
		drawVertices, drawIndices = appendColoredQuad(drawVertices, drawIndices, points, faceColor)

		if c.config.EdgeWidth == 0 {
			continue
		}
		edgeColor := c.config.EdgeColors[fd.index]
		for i := 0; i < 4; i++ {
			drawVertices, drawIndices = appendColoredLine(
				drawVertices, drawIndices, points[i], points[(i+1)%4], c.config.EdgeWidth, edgeColor,
			)
		}
	}

	return drawVertices, drawIndices
}

func appendColoredQuad(vertices []ebiten.Vertex, indices []uint16, points [4]point2D, clr color.RGBA) ([]ebiten.Vertex, []uint16) {
	base := uint16(len(vertices))
	r := float32(clr.R) / 255
	g := float32(clr.G) / 255
	b := float32(clr.B) / 255
	a := float32(clr.A) / 255
	for _, point := range points {
		vertices = append(vertices, ebiten.Vertex{
			DstX: point.x, DstY: point.y,
			SrcX: 1, SrcY: 1,
			ColorR: r, ColorG: g, ColorB: b, ColorA: a,
		})
	}
	indices = append(indices, base, base+1, base+2, base, base+2, base+3)
	return vertices, indices
}

func appendColoredLine(vertices []ebiten.Vertex, indices []uint16, from, to point2D, width float32, clr color.RGBA) ([]ebiten.Vertex, []uint16) {
	dx := to.x - from.x
	dy := to.y - from.y
	length := float32(math.Hypot(float64(dx), float64(dy)))
	if length == 0 {
		return vertices, indices
	}
	halfWidth := width / (2 * length)
	nx, ny := -dy*halfWidth, dx*halfWidth
	points := [4]point2D{
		{x: from.x + nx, y: from.y + ny},
		{x: to.x + nx, y: to.y + ny},
		{x: to.x - nx, y: to.y - ny},
		{x: from.x - nx, y: from.y - ny},
	}
	return appendColoredQuad(vertices, indices, points, clr)
}
