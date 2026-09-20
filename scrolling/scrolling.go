// Package scrolling provides one text pipeline for plain, controlled, mixed-font,
// per-glyph animated and post-deformed scrollers, preserving native pixel geometry.
package scrolling

import (
	"fmt"
	"math"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
	"github.com/olivierh59500/democonstructionkit/font"
	"github.com/olivierh59500/democonstructionkit/scrolltext"
)

// Face associates independent atlas metrics with optional native scaling.
type Face struct {
	Atlas          *ebiten.Image
	Metrics        *font.Font
	ScaleX, ScaleY float64
}

// Glyph can also be supplied directly when migrating a pre-sliced legacy font.
// Advance and Offset use unscaled pen coordinates; Scale applies to the bitmap.
type Glyph struct {
	Image                                 *ebiten.Image
	Rune                                  rune
	Advance, Offset, X, Y, ScaleX, ScaleY float64
	Font, Effect                          string
}
type Sample struct {
	Index                int
	Glyph                Glyph
	X, Y, Time, Position float64
	Shape                string
}

// Mapper edits the complete draw options and can hide a glyph by returning false.
// It runs in drawing order, so an original sine recurrence can be preserved exactly.
type Mapper func(Sample, *ebiten.DrawImageOptions) bool

type Config struct {
	Text             string
	Fonts            map[string]Face
	Font             string
	Controls         scrolltext.Decoder
	Tokens           []scrolltext.Token
	Glyphs           []Glyph
	Speed, Gap, X, Y float64
	Advance          float64 // Optional pen step independent of each glyph's bitmap width.
	Vertical, Repeat bool
	Map              Mapper
	Effects          map[string]Mapper
	Shapes           map[string]Mapper
	Shape            string
	Modes            map[string]Mode
	Sequence         *ModeSequence // When supplied, overrides text shape controls.
}

// DrawState allows an original production to retain its exact tick counters,
// reset conditions and visible index range while sharing text rendering.
// ScaleX/ScaleY are explicit: use IdentityState to start with unit scale.
type DrawState struct {
	bounded                              bool
	X, Y, ScaleX, ScaleY, Time, Position float64
	First, End                           int
	Reverse                              bool
	Cycle                                bool // Treat First/End as virtual indices through repeated text.
	Shape                                string
	Map                                  Mapper
	Paint                                func(*ebiten.Image, Sample, ebiten.DrawImageOptions)
	Options                              ebiten.DrawImageOptions
}

func IdentityState() DrawState { return DrawState{ScaleX: 1, ScaleY: 1, End: -1} }

type segment struct {
	start, end, position, speed float64
	shape                       string
}

// Scrolling owns cached layout, glyph slices and a deterministic control timeline.
// Assets stay caller-owned. Compose it with composite.Pass for ordered raster,
// scanline, column, masking or perspective operations on the complete text image.
type Scrolling struct {
	config           Config
	glyphs           []Glyph
	length, duration float64
	segments         []segment
	frame            kit.Frame
	draws            []glyphDraw
}

type glyphDraw struct {
	sample  Sample
	options ebiten.DrawImageOptions
	depth   float64
}

func New(c Config) (*Scrolling, error) {
	if !finite(c.Speed) || c.Speed < 0 || !finite(c.Gap) || c.Gap < 0 || !finite(c.X) || !finite(c.Y) || !finite(c.Advance) || c.Advance < 0 {
		return nil, fmt.Errorf("scrolling: invalid speed/gap")
	}
	s := &Scrolling{config: c}
	if !s.hasShape(c.Shape) {
		return nil, fmt.Errorf("scrolling: unknown initial mode %q", c.Shape)
	}
	if c.Sequence != nil {
		for _, cue := range c.Sequence.cues {
			if !s.hasShape(cue.Mode) {
				return nil, fmt.Errorf("scrolling: unknown sequence mode %q", cue.Mode)
			}
		}
	}
	if c.Glyphs != nil {
		if c.Text != "" || len(c.Tokens) > 0 {
			return nil, fmt.Errorf("scrolling: choose text or supplied glyphs")
		}
		s.glyphs = append([]Glyph(nil), c.Glyphs...)
		for i := range s.glyphs {
			g := &s.glyphs[i]
			if !finite(g.Advance) || g.Advance < 0 {
				return nil, fmt.Errorf("scrolling: invalid glyph advance")
			}
			g.Offset = s.length
			s.length += g.Advance
			if g.ScaleX == 0 {
				g.ScaleX = 1
			}
			if g.ScaleY == 0 {
				g.ScaleY = 1
			}
		}
		if len(s.glyphs) > 0 && s.length <= 0 {
			return nil, fmt.Errorf("scrolling: a glyph sequence must advance")
		}
		s.addSegment(0, s.length+c.Gap, c.Speed, c.Shape)
		return s, s.prepareModes()
	}
	tokens := c.Tokens
	var err error
	if tokens == nil {
		tokens, err = scrolltext.Parse(c.Text, c.Controls)
		if err != nil {
			return nil, err
		}
	}
	faceName := c.Font
	if faceName == "" {
		faceName = "default"
	}
	sx, sy, spacing := 1.0, 1.0, 0.0
	effect, shape := "", c.Shape
	speed, lastMarker := c.Speed, 0.0
	if _, ok := c.Fonts[faceName]; !ok {
		return nil, fmt.Errorf("scrolling: missing initial font %q", faceName)
	}
	flush := func() { s.addSegment(lastMarker, s.length, speed, shape); lastMarker = s.length }
	cache := map[string]map[rune]*ebiten.Image{}
	for _, t := range tokens {
		switch t.Kind {
		case scrolltext.Text:
			face := c.Fonts[faceName]
			if face.Atlas == nil || face.Metrics == nil || !face.Metrics.Bounds().In(face.Atlas.Bounds()) {
				return nil, fmt.Errorf("scrolling: invalid face %q", faceName)
			}
			fx, fy := face.ScaleX, face.ScaleY
			if fx == 0 {
				fx = 1
			}
			if fy == 0 {
				fy = 1
			}
			for _, r := range t.Text {
				metric, _ := face.Metrics.Glyph(r)
				penAdvance := metric.Advance
				if c.Advance > 0 {
					penAdvance = c.Advance
				}
				advance := (penAdvance + spacing) * sx * fx
				if c.Vertical {
					advance = face.Metrics.LineHeight() * sy * fy
				}
				if advance <= 0 || !finite(advance) {
					return nil, fmt.Errorf("scrolling: nonpositive advance")
				}
				var img *ebiten.Image
				if !metric.Rect.Empty() {
					if cache[faceName] == nil {
						cache[faceName] = map[rune]*ebiten.Image{}
					}
					img = cache[faceName][r]
					if img == nil {
						img = face.Atlas.SubImage(metric.Rect).(*ebiten.Image)
						cache[faceName][r] = img
					}
				}
				s.glyphs = append(s.glyphs, Glyph{Image: img, Rune: r, Advance: advance, Offset: s.length, X: metric.OffsetX * sx * fx, Y: metric.OffsetY * sy * fy, ScaleX: sx * fx, ScaleY: sy * fy, Font: faceName, Effect: effect})
				s.length += advance
			}
		case scrolltext.Font:
			if _, ok := c.Fonts[t.Text]; !ok {
				return nil, fmt.Errorf("scrolling: unknown font %q", t.Text)
			}
			faceName = t.Text
		case scrolltext.Effect:
			if t.Text != "none" {
				if _, ok := c.Effects[t.Text]; !ok {
					return nil, fmt.Errorf("scrolling: unknown glyph effect %q", t.Text)
				}
			}
			effect = t.Text
		case scrolltext.Scale:
			if t.Value <= 0 || t.Second <= 0 || !finite(t.Value) || !finite(t.Second) {
				return nil, fmt.Errorf("scrolling: invalid scale")
			}
			sx, sy = t.Value, t.Second
		case scrolltext.Tracking:
			if !finite(t.Value) {
				return nil, fmt.Errorf("scrolling: invalid tracking")
			}
			spacing = t.Value
		case scrolltext.Speed:
			if t.Value < 0 || !finite(t.Value) {
				return nil, fmt.Errorf("scrolling: invalid speed")
			}
			flush()
			speed = t.Value
		case scrolltext.Shape:
			if t.Text != "none" {
				if !s.hasShape(t.Text) {
					return nil, fmt.Errorf("scrolling: unknown shape %q", t.Text)
				}
			}
			flush()
			shape = t.Text
		case scrolltext.Pause:
			if t.Value < 0 || !finite(t.Value) {
				return nil, fmt.Errorf("scrolling: invalid pause")
			}
			flush()
			s.segments = append(s.segments, segment{start: s.duration, end: s.duration + t.Value, position: s.length, shape: shape})
			s.duration += t.Value
		default:
			return nil, fmt.Errorf("scrolling: unknown token kind")
		}
	}
	s.addSegment(lastMarker, s.length+c.Gap, speed, shape)
	return s, s.prepareModes()
}

func (s *Scrolling) prepareModes() error {
	for name, mode := range s.config.Modes {
		if mode.Prepare != nil {
			if err := mode.Prepare(s.glyphs); err != nil {
				return fmt.Errorf("scrolling: mode %q: %w", name, err)
			}
		}
	}
	return nil
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func (s *Scrolling) addSegment(from, to, speed float64, shape string) {
	if to <= from {
		return
	}
	duration := math.Inf(1)
	if speed > 0 {
		duration = (to - from) / speed
	}
	s.segments = append(s.segments, segment{start: s.duration, end: s.duration + duration, position: from, speed: speed, shape: shape})
	s.duration += duration
}
func (s *Scrolling) Length() float64 { return s.length }
func (s *Scrolling) GlyphCount() int { return len(s.glyphs) }

func (s *Scrolling) hasShape(name string) bool {
	if name == "none" || name == "" {
		return true
	}
	_, legacy := s.config.Shapes[name]
	_, mode := s.config.Modes[name]
	return legacy || mode
}

// Window starts at an arbitrary character in a circular proportional message and
// includes exactly the glyphs whose pen origins fit extent. This preserves the
// original MegaTwist/Coco letter-boundary texture refill strategy.
func (s *Scrolling) Window(first int, extent float64) DrawState {
	state := IdentityState()
	state.First = first
	state.End = first
	state.Cycle = true
	state.bounded = true
	if len(s.glyphs) == 0 || extent <= 0 || !finite(extent) {
		return state
	}
	offset := s.offsetAt(first)
	state.X = -offset
	if s.config.Vertical {
		state.X = 0
		state.Y = -offset
	}
	for s.offsetAt(state.End)-offset < extent {
		state.End++
	}
	return state
}
func (s *Scrolling) offsetAt(index int) float64 {
	n := len(s.glyphs)
	i := ((index % n) + n) % n
	cycle := (index - i) / n
	return s.glyphs[i].Offset + float64(cycle)*(s.length+s.config.Gap)
}

// StateAt integrates speed/pause controls exactly, including updates that skip
// multiple commands. A loop restarts its initial control state deterministically.
func (s *Scrolling) StateAt(seconds float64) DrawState {
	state := IdentityState()
	state.X = s.config.X
	state.Y = s.config.Y
	state.Time = seconds
	state.Shape = s.config.Shape
	t := math.Max(0, seconds)
	if s.config.Repeat && s.duration > 0 && !math.IsInf(s.duration, 0) {
		t = math.Mod(t, s.duration)
	}
	state.Position = s.length + s.config.Gap
	for _, part := range s.segments {
		if t < part.end {
			state.Position = part.position + (t-part.start)*part.speed
			state.Shape = part.shape
			break
		}
		state.Shape = part.shape
	}
	if s.config.Vertical {
		state.Y -= state.Position
	} else {
		state.X -= state.Position
	}
	if s.config.Sequence != nil {
		state.Shape = s.config.Sequence.At(seconds)
	}
	return state
}
func (s *Scrolling) Update(f kit.Frame) error { s.frame = f; return nil }
func (s *Scrolling) Draw(dst *ebiten.Image)   { state := s.StateAt(s.frame.Time); s.DrawAt(dst, state) }

// DrawAt does not advance time or position. Original render order and rounding
// belong to DrawState/Mapper, so no generic sine or gap is silently substituted.
func (s *Scrolling) DrawAt(dst *ebiten.Image, state DrawState) {
	if len(s.glyphs) == 0 {
		return
	}
	first, end := state.First, state.End
	if !state.Cycle {
		first = max(0, first)
	}
	if (!state.bounded && end < 0) || (!state.Cycle && end > len(s.glyphs)) {
		end = len(s.glyphs)
	}
	if first >= end {
		return
	}
	mode := s.config.Modes[state.Shape]
	s.draws = s.draws[:0]
	paint := state.Paint
	if paint == nil {
		paint = mode.Paint
	}
	flush := func(sample Sample, op ebiten.DrawImageOptions) {
		if paint != nil {
			paint(dst, sample, op)
		} else if sample.Glyph.Image != nil {
			dst.DrawImage(sample.Glyph.Image, &op)
		}
	}
	draw := func(i int) {
		index := i
		cycle := 0
		if state.Cycle {
			index = ((i % len(s.glyphs)) + len(s.glyphs)) % len(s.glyphs)
			cycle = (i - index) / len(s.glyphs)
		}
		g := s.glyphs[index]
		g.Offset += float64(cycle) * (s.length + s.config.Gap)
		x, y := state.X+g.X*state.ScaleX, state.Y+g.Y*state.ScaleY
		if s.config.Vertical {
			y += g.Offset * state.ScaleY
		} else {
			x += g.Offset * state.ScaleX
		}
		op := state.Options
		op.GeoM.Scale(g.ScaleX*state.ScaleX, g.ScaleY*state.ScaleY)
		op.GeoM.Translate(x, y)
		sample := Sample{Index: i, Glyph: g, X: x, Y: y, Time: state.Time, Position: state.Position, Shape: state.Shape}
		for _, mapper := range []Mapper{s.config.Map, s.config.Shapes[state.Shape], mode.Map, s.config.Effects[g.Effect], state.Map} {
			if mapper != nil && !mapper(sample, &op) {
				return
			}
		}
		if mode.Depth != nil {
			s.draws = append(s.draws, glyphDraw{sample, op, mode.Depth(sample)})
		} else {
			flush(sample, op)
		}
	}
	if state.Reverse {
		for i := end - 1; i >= first; i-- {
			draw(i)
		}
	} else {
		for i := first; i < end; i++ {
			draw(i)
		}
	}
	if mode.Depth != nil {
		sort.SliceStable(s.draws, func(i, j int) bool { return s.draws[i].depth > s.draws[j].depth })
		for _, d := range s.draws {
			flush(d.sample, d.options)
		}
	}
}

// FromImages is a migration adapter for already sliced alphabets. Missing images
// still advance the pen. New programs should prefer Text plus configured Faces.
func FromImages(images []*ebiten.Image, advance float64) (*Scrolling, error) {
	glyphs := make([]Glyph, len(images))
	for i, img := range images {
		glyphs[i] = Glyph{Image: img, Advance: advance}
	}
	return New(Config{Glyphs: glyphs})
}
