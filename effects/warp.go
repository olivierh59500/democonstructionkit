package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/render"
)

// Warp maps a reusable effect surface through a grid. One row per source scanline
// implements MegaTwist; one column per pixel implements a sine/DNA scroller.
// Map may project onto a 3D plane. Depth sorts cells back-to-front when provided.
type Warp struct {
	State
	Source        kit.Effect
	Map           func(x, y, seconds float64) geometry.Vec2
	Depth         func(x, y, seconds float64) float64
	Tint          func(x, y, seconds float64) color.Color
	canvas        *ebiten.Image
	columns, rows int
	batch         *render.Batch
	cells         []warpCell
}
type warpCell struct{ x, y, w, h, depth float64 }

func NewWarp(source kit.Effect, width, height, columns, rows int) (*Warp, error) {
	if source == nil || width <= 0 || height <= 0 || columns <= 0 || rows <= 0 || columns > 65536/rows {
		return nil, fmt.Errorf("effects: invalid warp grid")
	}
	w := &Warp{Source: source, canvas: render.NewSurface(width, height), columns: columns, rows: rows, batch: render.NewBatch(2048)}
	for y := 0; y < rows; y++ {
		for x := 0; x < columns; x++ {
			w.cells = append(w.cells, warpCell{x: float64(x*width) / float64(columns), y: float64(y*height) / float64(rows), w: float64(width) / float64(columns), h: float64(height) / float64(rows)})
		}
	}
	return w, nil
}
func (w *Warp) Update(f kit.Frame) error {
	w.Frame = f
	if w.Depth != nil {
		for i := range w.cells {
			c := &w.cells[i]
			c.depth = w.Depth(c.x+c.w/2, c.y+c.h/2, f.Time)
		}
		sort.SliceStable(w.cells, func(i, j int) bool { return w.cells[i].depth > w.cells[j].depth })
	}
	return w.Source.Update(f)
}
func (w *Warp) Draw(dst *ebiten.Image) {
	w.canvas.Clear()
	w.Source.Draw(w.canvas)
	w.batch.Begin(dst, w.canvas)
	for _, c := range w.cells {
		var quad [4]ebiten.Vertex
		for i, p := range [4]geometry.Vec2{{X: c.x, Y: c.y}, {X: c.x + c.w, Y: c.y}, {X: c.x + c.w, Y: c.y + c.h}, {X: c.x, Y: c.y + c.h}} {
			q := p
			if w.Map != nil {
				q = w.Map(p.X, p.Y, w.Frame.Time)
			}
			tint := color.Color(color.White)
			if w.Tint != nil {
				tint = w.Tint(p.X, p.Y, w.Frame.Time)
			}
			quad[i] = render.Vertex(q.X, q.Y, p.X, p.Y, tint)
		}
		w.batch.Quad(quad)
	}
	w.batch.Flush()
}
func (w *Warp) Close() error { w.canvas.Deallocate(); return kit.Close(w.Source) }

// Mask combines a color/raster effect with the alpha channel of another effect.
type Mask struct {
	Content, Alpha kit.Effect
	canvas, mask   *ebiten.Image
	config         MaskConfig
}

// MaskConfig combines two complete effects with an editable Porter-Duff blend.
// MaskX/Y place the alpha effect within the intermediate surface; OutputX/Y
// place the result on the destination. ClearTop removes an authored top band.
type MaskConfig struct {
	Content, Alpha   kit.Effect
	Width, Height    int
	Blend            ebiten.Blend
	MaskX, MaskY     float64
	OutputX, OutputY float64
	ClearTop         int
}

func NewMask(content, alpha kit.Effect, width, height int) (*Mask, error) {
	return NewMaskWith(MaskConfig{Content: content, Alpha: alpha, Width: width, Height: height,
		Blend: ebiten.BlendDestinationIn})
}

// NewMaskWith owns both input effects and its two reusable working surfaces.
func NewMaskWith(c MaskConfig) (*Mask, error) {
	if c.Content == nil || c.Alpha == nil || c.Width <= 0 || c.Height <= 0 ||
		c.ClearTop < 0 || c.ClearTop > c.Height {
		return nil, fmt.Errorf("effects: invalid mask")
	}
	for _, value := range [...]float64{c.MaskX, c.MaskY, c.OutputX, c.OutputY} {
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, fmt.Errorf("effects: nonfinite mask placement")
		}
	}
	if c.Blend == (ebiten.Blend{}) {
		c.Blend = ebiten.BlendDestinationIn
	}
	return &Mask{Content: c.Content, Alpha: c.Alpha, config: c,
		canvas: render.NewSurface(c.Width, c.Height), mask: render.NewSurface(c.Width, c.Height)}, nil
}
func (m *Mask) Update(f kit.Frame) error {
	if err := m.Content.Update(f); err != nil {
		return err
	}
	return m.Alpha.Update(f)
}
func (m *Mask) Draw(dst *ebiten.Image) {
	m.canvas.Clear()
	m.mask.Clear()
	m.Content.Draw(m.canvas)
	m.Alpha.Draw(m.mask)
	maskOptions := ebiten.DrawImageOptions{Blend: m.config.Blend}
	maskOptions.GeoM.Translate(m.config.MaskX, m.config.MaskY)
	m.canvas.DrawImage(m.mask, &maskOptions)
	if m.config.ClearTop > 0 {
		m.canvas.SubImage(image.Rect(0, 0, m.config.Width, m.config.ClearTop)).(*ebiten.Image).Clear()
	}
	var output ebiten.DrawImageOptions
	output.GeoM.Translate(m.config.OutputX, m.config.OutputY)
	dst.DrawImage(m.canvas, &output)
}
func (m *Mask) Close() error {
	m.canvas.Deallocate()
	m.mask.Deallocate()
	return kit.Group{m.Content, m.Alpha}.Close()
}

// Reflection mirrors a source layer below a configurable horizon.
type Reflection struct {
	State
	Source         kit.Effect
	Horizon, Scale float64
	Alpha          float32
	canvas         *ebiten.Image
}

func NewReflection(source kit.Effect, width, height int, horizon float64) (*Reflection, error) {
	if source == nil || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("effects: invalid reflection")
	}
	return &Reflection{Source: source, Horizon: horizon, Scale: 1, Alpha: .35, canvas: render.NewSurface(width, height)}, nil
}
func (r *Reflection) Update(f kit.Frame) error { r.Frame = f; return r.Source.Update(f) }
func (r *Reflection) Draw(dst *ebiten.Image) {
	r.canvas.Clear()
	r.Source.Draw(r.canvas)
	dst.DrawImage(r.canvas, nil)
	op := ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, -r.Scale)
	op.GeoM.Translate(0, r.Horizon*(1+r.Scale))
	op.ColorScale.ScaleAlpha(r.Alpha)
	// Only content above the horizon contributes to the reflected image.
	height := max(0, min(r.canvas.Bounds().Dy(), int(math.Floor(r.Horizon))))
	if height > 0 {
		source := r.canvas.SubImage(image.Rect(0, 0, r.canvas.Bounds().Dx(), height)).(*ebiten.Image)
		dst.DrawImage(source, &op)
	}
}
func (r *Reflection) Close() error { r.canvas.Deallocate(); return kit.Close(r.Source) }
