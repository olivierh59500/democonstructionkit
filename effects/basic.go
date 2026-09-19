// Package effects implements configurable demoscene effect families.
package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/geometry"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// State stores the last absolute update sample for stateless drawing effects.
type State struct{ Frame kit.Frame }

func (s *State) Update(f kit.Frame) error { s.Frame = f; return nil }

// Solid fills the destination, normally as the first layer of a group.
type Solid struct {
	State
	Color color.Color
}

func (s *Solid) Draw(dst *ebiten.Image) { dst.Fill(s.Color) }

// Pose describes a sprite centered on X,Y. Zero scale hides the sprite.
type Pose struct {
	X, Y, ScaleX, ScaleY, Angle float64
	Alpha                       float32
}

func IdentityPose(x, y float64) Pose { return Pose{X: x, Y: y, ScaleX: 1, ScaleY: 1, Alpha: 1} }

// Sprite supports sprite sheets, independent animation FPS and arbitrary paths.
type Sprite struct {
	State
	Frames []*ebiten.Image
	FPS    float64
	Pose   Pose
	Path   func(float64) Pose
	Blend  ebiten.Blend
}

func (s *Sprite) Draw(dst *ebiten.Image) {
	if len(s.Frames) == 0 {
		return
	}
	i := int(motion.Wrap(math.Floor(s.Frame.Time*s.FPS), float64(len(s.Frames))))
	img := s.Frames[i]
	p := s.Pose
	if s.Path != nil {
		p = s.Path(s.Frame.Time)
	}
	b := img.Bounds()
	op := ebiten.DrawImageOptions{Blend: s.Blend}
	op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
	op.GeoM.Scale(p.ScaleX, p.ScaleY)
	op.GeoM.Rotate(p.Angle)
	op.GeoM.Translate(p.X, p.Y)
	op.ColorScale.ScaleAlpha(max(0, min(1, p.Alpha)))
	dst.DrawImage(img, &op)
}

// SplitSheet validates actual frame rectangles, including atlas gutters.
func SplitSheet(atlas *ebiten.Image, rectangles []image.Rectangle) ([]*ebiten.Image, error) {
	if atlas == nil {
		return nil, fmt.Errorf("effects: nil sprite sheet")
	}
	out := make([]*ebiten.Image, len(rectangles))
	for i, r := range rectangles {
		if r.Empty() || !r.In(atlas.Bounds()) {
			return nil, fmt.Errorf("effects: sprite %d outside sheet", i)
		}
		out[i] = atlas.SubImage(r).(*ebiten.Image)
	}
	return out, nil
}

// RasterBar samples a vertical color profile along a moving horizontal bar.
type RasterBar struct {
	Y      float64
	Motion motion.Waves
	Colors []color.NRGBA
	Height float64
}
type RasterBars struct {
	State
	Bars  []RasterBar
	white *ebiten.Image
	batch *render.Batch
}

func NewRasterBars(bars []RasterBar) *RasterBars {
	w := ebiten.NewImage(1, 1)
	w.Fill(color.White)
	return &RasterBars{Bars: bars, white: w, batch: render.NewBatch(1024)}
}
func (r *RasterBars) Draw(dst *ebiten.Image) {
	r.batch.Begin(dst, r.white)
	for i, b := range r.Bars {
		if len(b.Colors) == 0 || b.Height <= 0 {
			continue
		}
		y := b.Y + b.Motion.At(float64(i), r.Frame.Time)
		h := b.Height / float64(len(b.Colors))
		for j, c := range b.Colors {
			r.batch.Rect(0, y+float64(j)*h, float64(dst.Bounds().Dx()), h, image.Rect(0, 0, 1, 1), c)
		}
	}
	r.batch.Flush()
}
func (r *RasterBars) Close() error { r.white.Deallocate(); return nil }

// TileTransform defines a repeating texture's center, zoom, rotation and phase.
type TileTransform struct {
	Center                       geometry.Vec2
	Scale, Angle, PhaseX, PhaseY float64
}

// Tiles renders a tiled or rotozoom background with one repeating textured quad.
// Draw cost is independent of zoom and the number of visible tile repetitions.
type Tiles struct {
	State
	Texture   *ebiten.Image
	Transform TileTransform
	Animate   func(float64) TileTransform
	batch     *render.Batch
}

func (t *Tiles) Draw(dst *ebiten.Image) {
	if t.Texture == nil {
		return
	}
	p := t.Transform
	if t.Animate != nil {
		p = t.Animate(t.Frame.Time)
	}
	if math.Abs(p.Scale) < .0001 || !finite(p.Scale) {
		return
	}
	b := t.Texture.Bounds()
	s, c := math.Sincos(p.Angle)
	if t.batch == nil {
		t.batch = render.NewBatch(2)
		t.batch.Options.Address = ebiten.AddressRepeat
	}
	t.batch.Begin(dst, t.Texture)
	w, h := float64(dst.Bounds().Dx()), float64(dst.Bounds().Dy())
	var quad [4]ebiten.Vertex
	for i, v := range [4]geometry.Vec2{{}, {X: w}, {X: w, Y: h}, {Y: h}} {
		x, y := (v.X-p.Center.X)/p.Scale, (v.Y-p.Center.Y)/p.Scale
		u, vv := x*c+y*s-p.PhaseX+float64(b.Min.X), -x*s+y*c-p.PhaseY+float64(b.Min.Y)
		quad[i] = render.Vertex(v.X, v.Y, u, vv, color.White)
	}
	t.batch.Quad(quad)
	t.batch.Flush()
}

// Tilemap is a camera-scrolled atlas map. Negative cells are transparent.
type Tilemap struct {
	State
	Atlas         *ebiten.Image
	Tiles         []image.Rectangle
	Cells         []int
	Columns, Rows int
	TileSize      image.Point
	Camera        func(float64) geometry.Vec2
	batch         *render.Batch
}

func NewTilemap(atlas *ebiten.Image, tiles []image.Rectangle, cells []int, columns, rows int, size image.Point) (*Tilemap, error) {
	if atlas == nil || columns <= 0 || rows <= 0 || columns > len(cells)/rows || columns*rows != len(cells) || size.X <= 0 || size.Y <= 0 {
		return nil, fmt.Errorf("effects: invalid tilemap")
	}
	for _, r := range tiles {
		if r.Empty() || !r.In(atlas.Bounds()) {
			return nil, fmt.Errorf("effects: tile outside atlas")
		}
	}
	for _, i := range cells {
		if i >= len(tiles) {
			return nil, fmt.Errorf("effects: invalid tile index")
		}
	}
	return &Tilemap{Atlas: atlas, Tiles: append([]image.Rectangle(nil), tiles...), Cells: append([]int(nil), cells...), Columns: columns, Rows: rows, TileSize: size, batch: render.NewBatch(1024)}, nil
}
func (m *Tilemap) Draw(dst *ebiten.Image) {
	cam := geometry.Vec2{}
	if m.Camera != nil {
		cam = m.Camera(m.Frame.Time)
	}
	w, h := float64(m.TileSize.X), float64(m.TileSize.Y)
	x0, y0 := max(0, int(math.Floor(cam.X/w))), max(0, int(math.Floor(cam.Y/h)))
	x1 := min(m.Columns, int(math.Ceil((cam.X+float64(dst.Bounds().Dx()))/w)))
	y1 := min(m.Rows, int(math.Ceil((cam.Y+float64(dst.Bounds().Dy()))/h)))
	m.batch.Begin(dst, m.Atlas)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			i := m.Cells[y*m.Columns+x]
			if i >= 0 {
				m.batch.Rect(float64(x)*w-cam.X, float64(y)*h-cam.Y, w, h, m.Tiles[i], color.White)
			}
		}
	}
	m.batch.Flush()
}
