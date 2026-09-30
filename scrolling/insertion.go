package scrolling

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

// InsertionConfig combines an owned horizontal control/insertion program with ordinary
// font-aware glyph rendering. Controller can borrow an existing pure program;
// ExternalClock leaves its advancement to the author while DCK owns the renderer.
type InsertionConfig struct {
	Program       scrolltext.InsertionProgramConfig
	Controller    *scrolltext.InsertionProgram
	ExternalClock bool
	Fonts         map[string]Face
	Font          string
	X, Y          float64
}
type insertionScroll struct {
	program  *scrolltext.InsertionProgram
	renderer *Scrolling
	external bool
	x, y     float64
	frame    kit.Frame
}

func newInsertion(c InsertionConfig) (*insertionScroll, error) {
	if !finite(c.X) || !finite(c.Y) || len(c.Fonts) < 1 || len(c.Fonts) > 256 {
		return nil, fmt.Errorf("scrolling: invalid insertion placement or fonts")
	}
	p := c.Controller
	var err error
	if p == nil {
		p, err = scrolltext.NewInsertionProgram(c.Program)
		if err != nil {
			return nil, err
		}
	} else if len(c.Program.Tokens) > 0 {
		return nil, fmt.Errorf("scrolling: choose insertion program or borrowed controller")
	}
	name := c.Font
	if name == "" {
		name = "default"
	}
	glyphs := make([]Glyph, len(p.Glyphs()))
	for i, t := range p.Glyphs() {
		fontName := t.Font
		if fontName == "" {
			fontName = name
		}
		face, ok := c.Fonts[fontName]
		if !ok || face.Atlas == nil || face.Metrics == nil || !face.Metrics.Bounds().In(face.Atlas.Bounds()) {
			return nil, fmt.Errorf("scrolling: invalid insertion font %q", fontName)
		}
		metric, _ := face.Metrics.Glyph(t.Rune)
		sx, sy := face.ScaleX, face.ScaleY
		if sx == 0 {
			sx = 1
		}
		if sy == 0 {
			sy = 1
		}
		if !finite(sx) || !finite(sy) || sx <= 0 || sy <= 0 {
			return nil, fmt.Errorf("scrolling: invalid insertion face scale")
		}
		var img *ebiten.Image
		if !metric.Rect.Empty() {
			img = face.Atlas.SubImage(metric.Rect).(*ebiten.Image)
		}
		glyphs[i] = Glyph{Rune: t.Rune, Font: fontName, Image: img, Advance: float64(t.Advance), X: metric.OffsetX * sx, Y: metric.OffsetY * sy, ScaleX: sx, ScaleY: sy}
	}
	r, err := New(Config{Glyphs: glyphs})
	if err != nil {
		return nil, err
	}
	return &insertionScroll{program: p, renderer: r, external: c.ExternalClock, x: c.X, y: c.Y}, nil
}
func (i *insertionScroll) Update(f kit.Frame) error {
	if err := i.renderer.Err(); err != nil {
		return err
	}
	i.frame = f
	if !i.external {
		if _, err := i.program.Step(); err != nil {
			return err
		}
	}
	return nil
}
func (i *insertionScroll) Finished() bool { return i.program.State().Finished }
func (i *insertionScroll) Draw(dst *ebiten.Image) {
	if dst == nil {
		return
	}
	s := i.program.State()
	state := IdentityState()
	state.First, state.End = s.First, s.Fetched
	state.Y = i.y
	state.Time = i.frame.Time
	state.Map = func(sample Sample, op *ebiten.DrawImageOptions) bool {
		op.GeoM.Translate(i.x+float64(s.Origins[sample.Index]-s.Distance)-sample.Glyph.Offset, 0)
		return true
	}
	i.renderer.DrawAt(dst, state)
}
func (i *insertionScroll) Close() error { return i.renderer.Close() }
func (s *Scrolling) InsertionController() *scrolltext.InsertionProgram {
	if p, ok := s.backend.(*insertionScroll); ok {
		return p.program
	}
	return nil
}
