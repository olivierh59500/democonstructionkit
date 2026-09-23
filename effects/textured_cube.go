package effects

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
)

// TexturedCubeConfig controls an affine-textured, depth-sorted cube. Size is its
// full edge length, Depth is an additional camera offset, and angles are XYZ
// radians. UVs are normalized per face; nil maps the full image onto every face.
type TexturedCubeConfig struct {
	Size, Focal, Depth, X, Y  float64
	Rotation, AngularVelocity geometry.Vec3
	FarToNear, CullBackFaces  bool
	FrontClockwise            bool // Positive projected winding; false selects negative winding.
	UV                        *[6][4]geometry.Vec2
	Filter                    ebiten.Filter
	Blend                     ebiten.Blend
}

func DefaultTexturedCubeConfig(size float64) TexturedCubeConfig {
	return TexturedCubeConfig{Size: size, Focal: 300, Depth: math.Max(300, size), FarToNear: true, CullBackFaces: true}
}

// TexturedCube borrows its texture, which may be a live plasma or another
// effect's surface. Instances own only their pose and fixed geometry buffers.
// Update samples absolute-time rotation; Rotate is the fixed-step alternative.
type TexturedCube struct {
	Rotation geometry.Vec3
	config   TexturedCubeConfig
	texture  *ebiten.Image
	points   [8]geometry.Vec3
	depths   [6]texturedCubeDepth
	vertices [24]ebiten.Vertex
	indices  [36]uint16
}
type texturedCubeDepth struct {
	face  int
	depth float64
}

var texturedCubeFaces = [6][4]int{{4, 5, 6, 7}, {1, 0, 3, 2}, {5, 1, 2, 6}, {0, 4, 7, 3}, {7, 6, 2, 3}, {0, 1, 5, 4}}
var texturedCubeCorners = [8]geometry.Vec3{{X: -1, Y: -1, Z: -1}, {X: 1, Y: -1, Z: -1}, {X: 1, Y: 1, Z: -1}, {X: -1, Y: 1, Z: -1}, {X: -1, Y: -1, Z: 1}, {X: 1, Y: -1, Z: 1}, {X: 1, Y: 1, Z: 1}, {X: -1, Y: 1, Z: 1}}

func NewTexturedCube(texture *ebiten.Image, c TexturedCubeConfig) (*TexturedCube, error) {
	for _, v := range []float64{c.Size, c.Focal, c.Depth, c.X, c.Y, c.Rotation.X, c.Rotation.Y, c.Rotation.Z, c.AngularVelocity.X, c.AngularVelocity.Y, c.AngularVelocity.Z} {
		if !finite(v) {
			return nil, fmt.Errorf("effects: nonfinite textured cube parameter")
		}
	}
	camera, radius := c.Focal+c.Depth, math.Sqrt(3)*(c.Size/2)
	if texture == nil || texture.Bounds().Empty() || c.Size <= 0 || c.Focal <= 0 || !finite(camera) || !finite(radius) || camera <= radius {
		return nil, fmt.Errorf("effects: invalid textured cube image or camera")
	}
	projectedRadius := radius * (c.Focal / (camera - radius))
	if !finite(projectedRadius) || math.Abs(c.X)+projectedRadius > math.MaxFloat32/2 || math.Abs(c.Y)+projectedRadius > math.MaxFloat32/2 {
		return nil, fmt.Errorf("effects: cube projection exceeds GPU coordinate range")
	}
	if c.UV != nil {
		copy := *c.UV
		for _, face := range copy {
			for _, uv := range face {
				b := texture.Bounds()
				u, v := float64(b.Min.X)+uv.X*float64(b.Dx()), float64(b.Min.Y)+uv.Y*float64(b.Dy())
				if !finite(u) || !finite(v) || math.Abs(u) > math.MaxFloat32 || math.Abs(v) > math.MaxFloat32 {
					return nil, fmt.Errorf("effects: nonfinite cube UV")
				}
			}
		}
		c.UV = &copy
	}
	return &TexturedCube{Rotation: c.Rotation, config: c, texture: texture}, nil
}

func (c *TexturedCube) Update(f kit.Frame) error {
	if !finite(f.Time) {
		return fmt.Errorf("effects: nonfinite cube time")
	}
	rotation := c.config.Rotation.Add(c.config.AngularVelocity.Scale(f.Time))
	if !finite(rotation.X) || !finite(rotation.Y) || !finite(rotation.Z) {
		return fmt.Errorf("effects: cube rotation overflows")
	}
	c.Rotation = rotation
	return nil
}
func (c *TexturedCube) Rotate(x, y, z float64) {
	c.Rotation.X += x
	c.Rotation.Y += y
	c.Rotation.Z += z
}
func (c *TexturedCube) Draw(dst *ebiten.Image) { c.DrawAt(dst, c.config.X, c.config.Y) }
func (c *TexturedCube) DrawAt(dst *ebiten.Image, x, y float64) {
	if dst == nil || c.texture == nil {
		return
	}
	v, i := c.Geometry(x, y)
	dst.DrawTriangles(v, i, c.texture, &ebiten.DrawTrianglesOptions{Filter: c.config.Filter, Blend: c.config.Blend})
}

// Geometry returns borrowed arrays in stable face drawing order. Face grouping,
// projected winding and float32 camera-origin addition are retained explicitly.
func (c *TexturedCube) Geometry(centerX, centerY float64) ([]ebiten.Vertex, []uint16) {
	if c.texture == nil || !finite(centerX) || !finite(centerY) || !finite(c.Rotation.X) || !finite(c.Rotation.Y) || !finite(c.Rotation.Z) {
		return nil, nil
	}
	sx, cx := math.Sincos(c.Rotation.X)
	sy, cy := math.Sincos(c.Rotation.Y)
	sz, cz := math.Sincos(c.Rotation.Z)
	for i, v := range texturedCubeCorners {
		x, y, z := v.X*c.config.Size/2, v.Y*c.config.Size/2, v.Z*c.config.Size/2
		y2, z2 := y*cx-z*sx, y*sx+z*cx
		y, z = y2, z2
		x2, z2 := x*cy+z*sy, -x*sy+z*cy
		x = x2
		x2, y2 = x*cz-y*sz, x*sz+y*cz
		c.points[i] = geometry.Vec3{X: x2, Y: y2, Z: z2}
	}
	for i, f := range texturedCubeFaces {
		c.depths[i] = texturedCubeDepth{i, (c.points[f[0]].Z + c.points[f[1]].Z + c.points[f[2]].Z + c.points[f[3]].Z) / 4}
	}
	for i := 1; i < len(c.depths); i++ {
		item, j := c.depths[i], i
		for j > 0 && ((c.config.FarToNear && item.depth > c.depths[j-1].depth) || (!c.config.FarToNear && item.depth < c.depths[j-1].depth)) {
			c.depths[j] = c.depths[j-1]
			j--
		}
		c.depths[j] = item
	}
	vertices, indices := c.vertices[:0], c.indices[:0]
	b := c.texture.Bounds()
	tw, th := float32(b.Dx()), float32(b.Dy())
	for _, depth := range c.depths {
		face := texturedCubeFaces[depth.face]
		var points [4][2]float32
		for i, index := range face {
			v := c.points[index]
			scale := c.config.Focal / (c.config.Focal + v.Z + c.config.Depth)
			points[i] = [2]float32{float32(centerX) + float32(v.X*scale), float32(centerY) + float32(v.Y*scale)}
			if !finite(float64(points[i][0])) || !finite(float64(points[i][1])) {
				return nil, nil
			}
		}
		x1, y1 := points[1][0]-points[0][0], points[1][1]-points[0][1]
		x2, y2 := points[2][0]-points[0][0], points[2][1]-points[0][1]
		cross := x1*y2 - y1*x2
		if c.config.CullBackFaces && ((c.config.FrontClockwise && cross < 0) || (!c.config.FrontClockwise && cross > 0)) {
			continue
		}
		uvs := [4]geometry.Vec2{{}, {X: 1}, {X: 1, Y: 1}, {Y: 1}}
		if c.config.UV != nil {
			uvs = c.config.UV[depth.face]
		}
		base := uint16(len(vertices))
		for i, p := range points {
			vertices = append(vertices, ebiten.Vertex{DstX: p[0], DstY: p[1], SrcX: float32(b.Min.X) + float32(uvs[i].X)*tw, SrcY: float32(b.Min.Y) + float32(uvs[i].Y)*th, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1})
		}
		indices = append(indices, base, base+1, base+2, base, base+2, base+3)
	}
	return vertices, indices
}

// Close forgets the borrowed texture without destroying the caller's image.
func (c *TexturedCube) Close() error { c.texture = nil; return nil }
