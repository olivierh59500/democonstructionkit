package effects

import (
	"fmt"
	"image/color"
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
	w := &Warp{Source: source, canvas: ebiten.NewImage(width, height), columns: columns, rows: rows, batch: render.NewBatch(2048)}
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
		for i, p := range [4]geometry.Vec2{{c.x, c.y}, {c.x + c.w, c.y}, {c.x + c.w, c.y + c.h}, {c.x, c.y + c.h}} {
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
}

func NewMask(content, alpha kit.Effect, width, height int) (*Mask, error) {
	if content == nil || alpha == nil || width <= 0 || height <= 0 {
		return nil, fmt.Errorf("effects: invalid mask")
	}
	return &Mask{Content: content, Alpha: alpha, canvas: ebiten.NewImage(width, height), mask: ebiten.NewImage(width, height)}, nil
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
	m.canvas.DrawImage(m.mask, &ebiten.DrawImageOptions{Blend: ebiten.BlendDestinationIn})
	dst.DrawImage(m.canvas, nil)
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
	return &Reflection{Source: source, Horizon: horizon, Scale: 1, Alpha: .35, canvas: ebiten.NewImage(width, height)}, nil
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
	dst.DrawImage(r.canvas, &op)
}
func (r *Reflection) Close() error { r.canvas.Deallocate(); return kit.Close(r.Source) }
