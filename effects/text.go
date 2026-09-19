package effects

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/motion"
	"github.com/olivierh59500/democonstructionkit/render"
)

// Text renders resolved bitmap glyphs. GlyphPose optionally replaces each glyph's
// centered pose, enabling independent 3D letter tweens and per-character waves.
type Text struct {
	State
	Atlas                *ebiten.Image
	Layout               font.Layout
	X, Y, ScaleX, ScaleY float64
	Color                color.Color
	GlyphPose            func(index int, placement font.Placement, seconds float64) Pose
	batch                *render.Batch
}

func NewText(atlas *ebiten.Image, metrics *font.Font, message string, spacing float64) (*Text, error) {
	if atlas == nil || metrics == nil || !metrics.Bounds().In(atlas.Bounds()) {
		return nil, fmt.Errorf("effects: font and atlas bounds do not match")
	}
	l, err := metrics.Layout(message, spacing)
	if err != nil {
		return nil, err
	}
	return &Text{Atlas: atlas, Layout: l, ScaleX: 1, ScaleY: 1, Color: color.White, batch: render.NewBatch(512)}, nil
}
func (t *Text) Draw(dst *ebiten.Image) {
	t.batch.Begin(dst, t.Atlas)
	for i, g := range t.Layout.Glyphs {
		if g.Glyph.Rect.Empty() {
			continue
		}
		w, h := float64(g.Glyph.Rect.Dx()), float64(g.Glyph.Rect.Dy())
		p := IdentityPose(t.X+(g.X+w/2)*t.ScaleX, t.Y+(g.Y+h/2)*t.ScaleY)
		p.ScaleX = t.ScaleX
		p.ScaleY = t.ScaleY
		if t.GlyphPose != nil {
			p = t.GlyphPose(i, g, t.Frame.Time)
		}
		drawGlyph(t.batch, g.Glyph.Rect, p, t.Color)
	}
	t.batch.Flush()
}
func drawGlyph(b *render.Batch, rect image.Rectangle, p Pose, tint color.Color) {
	if p.Alpha <= 0 {
		return
	}
	c := color.NRGBAModel.Convert(tint).(color.NRGBA)
	c.A = uint8(float32(c.A) * max(0, min(1, p.Alpha)))
	w, h := float64(rect.Dx())*p.ScaleX/2, float64(rect.Dy())*p.ScaleY/2
	s, co := math.Sincos(p.Angle)
	var vertices [4]ebiten.Vertex
	xy := [4][2]float64{{-w, -h}, {w, -h}, {w, h}, {-w, h}}
	uv := [4]image.Point{rect.Min, {X: rect.Max.X, Y: rect.Min.Y}, rect.Max, {X: rect.Min.X, Y: rect.Max.Y}}
	for i, v := range xy {
		vertices[i] = render.Vertex(p.X+v[0]*co-v[1]*s, p.Y+v[0]*s+v[1]*co, float64(uv[i].X), float64(uv[i].Y), c)
	}
	b.Quad(vertices)
}

// Scroller repeats a horizontal line or a vertical page at a speed in pixels/s.
// Text geometry is cached. It never creates a texture the size of the message.
type Scroller struct {
	State
	Atlas                   *ebiten.Image
	Layout                  font.Layout
	Speed, Gap, X, Y, Scale float64
	Vertical                bool
	Wave                    motion.Waves
	Color                   color.Color
	width, height           int
	batch                   *render.Batch
	ends                    []float64
}
type ScrollConfig struct {
	Width, Height                    int
	Message                          string
	Spacing, Speed, Gap, X, Y, Scale float64
	Vertical                         bool
	Wave                             motion.Waves
}

func NewScroller(atlas *ebiten.Image, metrics *font.Font, c ScrollConfig) (*Scroller, error) {
	if c.Width <= 0 || c.Height <= 0 || c.Gap < 0 || !finite(c.Speed) || !finite(c.Gap) {
		return nil, fmt.Errorf("effects: invalid scroller config")
	}
	if !c.Vertical && strings.Contains(c.Message, "\n") {
		return nil, fmt.Errorf("effects: horizontal scroller needs one line")
	}
	t, err := NewText(atlas, metrics, c.Message, c.Spacing)
	if err != nil {
		return nil, err
	}
	if c.Scale == 0 {
		c.Scale = 1
	}
	if c.Scale <= 0 || !finite(c.Scale) {
		return nil, fmt.Errorf("effects: invalid text scale")
	}
	s := &Scroller{Atlas: atlas, Layout: t.Layout, Speed: c.Speed, Gap: c.Gap, X: c.X, Y: c.Y, Scale: c.Scale, Vertical: c.Vertical, Wave: c.Wave, Color: color.White, width: c.Width, height: c.Height, batch: t.batch}
	// Prefix maxima keep binary search valid with negative glyph bearings.
	end := 0.0
	for _, g := range s.Layout.Glyphs {
		if c.Vertical {
			end = math.Max(end, g.Y+float64(g.Glyph.Rect.Dy()))
		} else {
			end = math.Max(end, g.X+float64(g.Glyph.Rect.Dx()))
		}
		s.ends = append(s.ends, end)
	}
	return s, nil
}
func (s *Scroller) Draw(dst *ebiten.Image) {
	if len(s.Layout.Glyphs) == 0 {
		return
	}
	length, extent, offset := s.Layout.Width, float64(s.width), s.X
	if s.Vertical {
		length, extent, offset = s.Layout.Height, float64(s.height), s.Y
	}
	period := (length + s.Gap) * s.Scale
	if period <= 0 {
		return
	}
	phase := motion.Wrap(s.Frame.Time*s.Speed-offset, period)
	s.batch.Begin(dst, s.Atlas)
	for base := -phase; base < extent; base += period {
		first := sort.Search(len(s.ends), func(i int) bool { return base+s.ends[i]*s.Scale >= 0 })
		for i := first; i < len(s.Layout.Glyphs); i++ {
			g := s.Layout.Glyphs[i]
			r := g.Glyph.Rect
			if r.Empty() {
				continue
			}
			x, y := s.X+g.X*s.Scale, s.Y+g.Y*s.Scale
			if s.Vertical {
				y = base + g.Y*s.Scale
				if y > extent {
					break
				}
				x += s.Wave.At(y, s.Frame.Time)
			} else {
				x = base + g.X*s.Scale
				if x > extent {
					break
				}
				y += s.Wave.At(x, s.Frame.Time)
			}
			s.batch.Rect(x, y, float64(r.Dx())*s.Scale, float64(r.Dy())*s.Scale, r, s.Color)
		}
	}
	s.batch.Flush()
}
func finite(x float64) bool { return !math.IsNaN(x) && !math.IsInf(x, 0) }
