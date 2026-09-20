package scrolling

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// PlaneShaderSource masks a vertical raster by a glyph's alpha. Raster sampling
// remains in unscaled scene coordinates, independent of the source font's atlas.
const PlaneShaderSource = `//kage:unit pixels
package main
func Fragment(dstPos vec4, srcPos vec2, custom vec4) vec4 {
 glyph := imageSrc0UnsafeAt(srcPos)
 rasterY := floor(custom.r) + 0.5
 raster := imageSrc1UnsafeAt(imageSrc0Origin() + vec2(0.5, rasterY))
 return vec4(raster.rgb * glyph.a, raster.a * glyph.a)
}`

type PlaneRenderer struct {
	face   Face
	raster *ebiten.Image
	shader *ebiten.Shader
	batch  *composite.QuadBatch
}
type PlaneDraw struct{ OriginX, OriginY, ScaleX, ScaleY float32 }

func NewPlaneRenderer(face Face, raster *ebiten.Image) (*PlaneRenderer, error) {
	if face.Atlas == nil || face.Metrics == nil || !face.Metrics.Bounds().In(face.Atlas.Bounds()) {
		return nil, fmt.Errorf("scrolling: invalid plane font")
	}
	if face.ScaleX == 0 {
		face.ScaleX = 1
	}
	if face.ScaleY == 0 {
		face.ScaleY = 1
	}
	r := &PlaneRenderer{face: face, raster: raster, batch: composite.NewQuadBatch(64)}
	r.batch.AlternateDiagonal = true
	if raster != nil {
		var err error
		r.shader, err = ebiten.NewShader([]byte(PlaneShaderSource))
		if err != nil {
			return nil, err
		}
		r.batch.Shader = r.shader
		r.batch.ShaderOptions.Images[0] = face.Atlas
		r.batch.ShaderOptions.Images[1] = raster
	}
	return r, nil
}

func (r *PlaneRenderer) Draw(dst *ebiten.Image, points []PlanePoint, c PlaneDraw) {
	r.batch.Begin(dst, r.face.Atlas)
	for _, p := range points {
		if p.Rune == 0 || p.Scale <= 0 || !finite(p.Scale) {
			continue
		}
		g, ok := r.face.Metrics.Glyph(p.Rune)
		if !ok || g.Rect.Empty() {
			continue
		}
		sx, sy := float32(p.Scale*r.face.ScaleX), float32(p.Scale*r.face.ScaleY)
		width, height := float32(g.Rect.Dx()), float32(g.Rect.Dy())
		localY := float32(p.Y) + (float32(g.OffsetY)-height/2)*sy
		x, y := c.OriginX+c.ScaleX*(float32(p.X)+(float32(g.OffsetX)-width/2)*sx), c.OriginY+c.ScaleY*localY
		w, h := width*sx*c.ScaleX, height*sy*c.ScaleY
		u, v := float32(g.Rect.Min.X), float32(g.Rect.Min.Y)
		top, bottom := float32(1), float32(1)
		if r.raster != nil {
			top = localY
			bottom = localY + height*sy
		}
		r.batch.Quad([4]ebiten.Vertex{
			{DstX: x, DstY: y, SrcX: u, SrcY: v, ColorR: top, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: x + w, DstY: y, SrcX: u + width, SrcY: v, ColorR: top, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: x, DstY: y + h, SrcX: u, SrcY: v + height, ColorR: bottom, ColorG: 1, ColorB: 1, ColorA: 1},
			{DstX: x + w, DstY: y + h, SrcX: u + width, SrcY: v + height, ColorR: bottom, ColorG: 1, ColorB: 1, ColorA: 1},
		})
	}
	r.batch.Flush()
}

func (r *PlaneRenderer) Close() error {
	if r.shader != nil {
		r.shader.Deallocate()
		r.shader = nil
	}
	return nil
}
