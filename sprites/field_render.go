package sprites

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/render"
)

// FieldAppearance describes a particle in destination pixels. Tint is a
// premultiplied multiplier; its zero value is white. Anchor uses 0..1 fractions.
type FieldAppearance struct {
	Width, Height, Angle float64
	ScaleX, ScaleY       float64 // Zero defaults to one before Sample; Sample may set zero to hide.
	AnchorX, AnchorY     float64
	Tint                 ebiten.ColorScale
}

// FieldStyle is a skin for the common projected field. A nil Image selects a
// white pixel, so stars and image sprites use exactly the same projection and
// drawing path. Optional Frames are absolute atlas rectangles selected by Image.
// Sample can modulate size, color and rotation by depth, index or music signals.
type FieldStyle struct {
	Image        *ebiten.Image
	Frames       []image.Rectangle
	Appearance   FieldAppearance
	ScaleByDepth bool
	Streak       bool // Connect the last two sampled positions with a width-sized strip.
	Filter       ebiten.Filter
	Blend        ebiten.Blend
	Antialias    bool
	// DrawImages preserves DrawImage's transform arithmetic and source clipping.
	// Use it when reproducing an existing nearest-filtered sprite renderer.
	DrawImages bool
	Sample     func(FieldSample, *FieldAppearance) bool
}

// FieldRenderer batches pixels, image sprites and trails using persistent
// bounded geometry. Assets are borrowed; only the fallback white pixel is owned.
type FieldRenderer struct {
	white  *ebiten.Image
	batch  *render.Batch
	frames map[fieldFrame]*ebiten.Image
}

type fieldFrame struct {
	image *ebiten.Image
	rect  image.Rectangle
}

func NewFieldRenderer(capacity int) *FieldRenderer {
	white := ebiten.NewImage(1, 1)
	white.Fill(color.White)
	return &FieldRenderer{white: white, batch: render.NewBatch(max(1, min(capacity, 10000)) * 2)}
}

func (r *FieldRenderer) Draw(dst *ebiten.Image, samples []FieldSample, c FieldStyle) {
	if r.white == nil || dst == nil {
		return
	}
	source := c.Image
	if source == nil {
		source = r.white
	}
	r.batch.Options.Filter, r.batch.Options.Blend, r.batch.Options.AntiAlias = c.Filter, c.Blend, c.Antialias
	r.batch.Begin(dst, source)
	for _, p := range samples {
		rect := source.Bounds()
		if len(c.Frames) > 0 {
			if p.Image < 0 || p.Image >= len(c.Frames) {
				continue
			}
			rect = c.Frames[p.Image]
			if rect.Empty() || !rect.In(source.Bounds()) {
				continue
			}
		}
		a := c.Appearance
		if a.Width == 0 {
			a.Width = float64(rect.Dx())
		}
		if a.Height == 0 {
			a.Height = float64(rect.Dy())
		}
		if a.ScaleX == 0 {
			a.ScaleX = 1
		}
		if a.ScaleY == 0 {
			a.ScaleY = 1
		}
		if c.ScaleByDepth {
			a.Width *= p.Scale
			a.Height *= p.Scale
		}
		if c.Sample != nil && !c.Sample(p, &a) {
			continue
		}
		if a.Width <= 0 || a.Height <= 0 || !finiteField(a.Width) || !finiteField(a.Height) || !finiteField(a.Angle) || !finiteField(a.ScaleX) || !finiteField(a.ScaleY) || !finiteField(a.AnchorX) || !finiteField(a.AnchorY) {
			continue
		}
		if c.DrawImages && !c.Streak {
			r.batch.Flush()
			img := source
			if rect != source.Bounds() {
				key := fieldFrame{source, rect}
				img = r.frames[key]
				if img == nil {
					if r.frames == nil {
						r.frames = make(map[fieldFrame]*ebiten.Image)
					}
					img = source.SubImage(rect).(*ebiten.Image)
					r.frames[key] = img
				}
			}
			op := ebiten.DrawImageOptions{Filter: c.Filter, Blend: c.Blend, ColorScale: a.Tint}
			op.GeoM.Translate(-a.AnchorX*float64(rect.Dx()), -a.AnchorY*float64(rect.Dy()))
			op.GeoM.Scale(a.Width/float64(rect.Dx())*a.ScaleX, a.Height/float64(rect.Dy())*a.ScaleY)
			op.GeoM.Rotate(a.Angle)
			op.GeoM.Translate(p.X, p.Y)
			dst.DrawImage(img, &op)
			continue
		}
		a.Width *= a.ScaleX
		a.Height *= a.ScaleY
		x, y := p.X, p.Y
		if c.Streak {
			if !p.Connected {
				continue
			}
			dx, dy := p.X-p.PreviousX, p.Y-p.PreviousY
			a.Height = math.Hypot(dx, dy)
			if a.Height <= 0 {
				continue
			}
			a.Angle = math.Atan2(dy, dx) - math.Pi/2
			a.AnchorX, a.AnchorY = .5, 0
			x, y = p.PreviousX, p.PreviousY
		}
		sin, cos := math.Sincos(a.Angle)
		left, top := -a.AnchorX*a.Width, -a.AnchorY*a.Height
		corners := [4][4]float64{
			{left, top, float64(rect.Min.X), float64(rect.Min.Y)},
			{left + a.Width, top, float64(rect.Max.X), float64(rect.Min.Y)},
			{left + a.Width, top + a.Height, float64(rect.Max.X), float64(rect.Max.Y)},
			{left, top + a.Height, float64(rect.Min.X), float64(rect.Max.Y)},
		}
		var vertices [4]ebiten.Vertex
		for i, v := range corners {
			vertices[i] = ebiten.Vertex{DstX: float32(x + v[0]*cos - v[1]*sin), DstY: float32(y + v[0]*sin + v[1]*cos),
				SrcX: float32(v[2]), SrcY: float32(v[3]), ColorR: a.Tint.R(), ColorG: a.Tint.G(), ColorB: a.Tint.B(), ColorA: a.Tint.A()}
		}
		r.batch.Quad(vertices)
	}
	r.batch.Flush()
}

func (r *FieldRenderer) Close() error {
	if r.white != nil {
		r.white.Deallocate()
		r.white = nil
	}
	clear(r.frames)
	return nil
}
