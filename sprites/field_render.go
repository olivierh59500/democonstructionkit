package sprites

import (
	"image"
	"image/color"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/olivierh59500/democonstructionkit/render"
)

// FieldAppearance describes a particle in destination pixels. Tint is a
// premultiplied multiplier; its zero value is white. Anchor uses 0..1 fractions.
type FieldAppearance struct {
	Width, Height, Angle  float64
	ScaleX, ScaleY        float64 // Zero defaults to one before Sample; Sample may set zero to hide.
	AnchorX, AnchorY      float64
	Tint                  ebiten.ColorScale
	FillColor, TrailColor color.RGBA
	TrailWidth            float64
}

// FieldOutline draws an inset box border instead of sprite pixels. Width is
// measured in destination units before rotation and defaults to one. A nil
// Color selects white; its premultiplied result is multiplied by Tint.
type FieldOutline struct {
	Width float64
	Color color.Color
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
	// Outline takes precedence over image/vector skins, retaining image metrics,
	// anchor, scale, rotation, depth modulation and source-order overlap.
	Outline *FieldOutline
	// VectorRects draws each point as a filled rectangle and an optional trail
	// in source order, retaining vector pixel coverage and overlap behavior.
	VectorRects bool
	// VectorLines draws only connected projected endpoints as variable-width
	// anti-aliased lines; it does not draw a point rectangle.
	VectorLines bool
	// VectorTrailPaint overrides the per-sample TrailColor, retaining any
	// color.Color supplied by an existing vector-line effect.
	VectorTrailPaint color.Color
	// DrawImages preserves DrawImage's transform arithmetic and source clipping.
	// Use it when reproducing an existing nearest-filtered sprite renderer.
	DrawImages bool
	Sample     func(FieldSample, *FieldAppearance) bool
}

// FieldRenderer batches pixels, image sprites and trails using persistent
// bounded geometry. Assets are borrowed; only the fallback white pixel is owned.
type FieldRenderer struct {
	white       *ebiten.Image
	batch       *render.Batch
	frameSource *ebiten.Image
	frameRects  []image.Rectangle
	frameImages []*ebiten.Image
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
	material := source
	outlineWidth, outlineColor := 1.0, [4]float32{1, 1, 1, 1}
	if c.Outline != nil {
		if c.Outline.Width < 0 || !finiteField(c.Outline.Width) {
			return
		}
		if c.Outline.Width > 0 {
			outlineWidth = c.Outline.Width
		}
		if c.Outline.Color != nil {
			red, green, blue, alpha := c.Outline.Color.RGBA()
			outlineColor = [4]float32{float32(red) / 65535, float32(green) / 65535, float32(blue) / 65535, float32(alpha) / 65535}
		}
		material = r.white
	}
	if c.Outline == nil && c.DrawImages && len(c.Frames) > 0 {
		if r.frameSource != source || !slices.Equal(r.frameRects, c.Frames) {
			r.frameSource = source
			r.frameRects = append(r.frameRects[:0], c.Frames...)
			clear(r.frameImages)
			if cap(r.frameImages) < len(c.Frames) {
				r.frameImages = make([]*ebiten.Image, len(c.Frames))
			} else {
				r.frameImages = r.frameImages[:len(c.Frames)]
			}
			for i, rect := range c.Frames {
				if !rect.Empty() && rect.In(source.Bounds()) {
					r.frameImages[i] = source.SubImage(rect).(*ebiten.Image)
				}
			}
		}
	} else {
		r.frameSource = nil
		r.frameRects = r.frameRects[:0]
		clear(r.frameImages)
		r.frameImages = r.frameImages[:0]
	}
	r.batch.Options.Filter, r.batch.Options.Blend, r.batch.Options.AntiAlias = c.Filter, c.Blend, c.Antialias
	r.batch.Begin(dst, material)
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
		if c.Outline == nil && (c.VectorRects || c.VectorLines) {
			if c.VectorRects {
				vector.FillRect(dst, float32(p.X-a.Width/2), float32(p.Y-a.Height/2), float32(a.Width), float32(a.Height), a.FillColor, c.Antialias)
			}
			if a.TrailWidth > 0 && p.Connected && p.PreviousX != 0 && p.PreviousY != 0 {
				paint := color.Color(a.TrailColor)
				if c.VectorTrailPaint != nil {
					paint = c.VectorTrailPaint
				}
				vector.StrokeLine(dst, float32(p.PreviousX), float32(p.PreviousY), float32(p.X), float32(p.Y), float32(a.TrailWidth), paint, c.Antialias)
			}
			continue
		}
		if c.Outline == nil && c.DrawImages && !c.Streak {
			r.batch.Flush()
			img := source
			if len(c.Frames) > 0 {
				img = r.frameImages[p.Image]
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
		if c.Outline != nil {
			r.drawOutline(x, y, left, top, a, outlineWidth, outlineColor, sin, cos)
			continue
		}
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

func (r *FieldRenderer) drawOutline(x, y, left, top float64, a FieldAppearance, width float64, paint [4]float32, sin, cos float64) {
	if math.Abs(a.Width) <= 2*width || math.Abs(a.Height) <= 2*width {
		return
	}
	colors := [4]float32{paint[0] * a.Tint.R(), paint[1] * a.Tint.G(), paint[2] * a.Tint.B(), paint[3] * a.Tint.A()}
	w, h := math.Copysign(width, a.Width), math.Copysign(width, a.Height)
	for _, rect := range [4][4]float64{
		{left, top, a.Width, h}, {left, top + a.Height - h, a.Width, h},
		{left, top + h, w, a.Height - 2*h}, {left + a.Width - w, top + h, w, a.Height - 2*h},
	} {
		var vertices [4]ebiten.Vertex
		for i, p := range [4][4]float64{
			{rect[0], rect[1], 0, 0}, {rect[0] + rect[2], rect[1], 1, 0},
			{rect[0] + rect[2], rect[1] + rect[3], 1, 1}, {rect[0], rect[1] + rect[3], 0, 1},
		} {
			vertices[i] = ebiten.Vertex{DstX: float32(x + p[0]*cos - p[1]*sin), DstY: float32(y + p[0]*sin + p[1]*cos),
				SrcX: float32(p[2]), SrcY: float32(p[3]), ColorR: colors[0], ColorG: colors[1], ColorB: colors[2], ColorA: colors[3]}
		}
		r.batch.Quad(vertices)
	}
}

func (r *FieldRenderer) Close() error {
	if r.white != nil {
		r.white.Deallocate()
		r.white = nil
	}
	r.frameSource = nil
	r.frameRects = nil
	r.frameImages = nil
	r.batch = nil
	return nil
}
