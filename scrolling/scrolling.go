// Package scrolling provides one text pipeline for plain, controlled, mixed-font,
// per-glyph animated and post-deformed scrollers, preserving native pixel geometry.
package scrolling

import (
	"errors"
	"fmt"
	"image"
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
	Page             *PageConfig        // Optional multiline vertical layout; Vertical alone is a glyph column.
	Recycled         *RecycledConfig    // Authored recycled-slot transport, advanced once per Update.
	RingLanes        *RingLanesConfig   // Synchronized recycled-slot lanes with paced placement.
	Projected        *ProjectedConfig   // Authored visible-slot plane transport, advanced once per Update.
	Pseudo3D         *Pseudo3DConfig    // Multiple pseudo-3D text banks with shared harmonics and independent transport.
	Bands            *BitmapBandsConfig // Bounded repeated text lanes.
	Slots            *BitmapSlotsConfig // Recycled glyphs with pose mapping and tangent orientation.
	Crawl            *CrawlConfig       // Bounded paragraph transport and configurable perspective rows.
	Sliced           *SlicedConfig      // Streaming DNA with independent transport and rotation clocks.
	Feed             *FeedConfig        // Finite glyph insertion into a persistent scrolling trail.
	Scanline         *ScanlineConfig    // Proportional text with cumulative row sampling and bounce.
	Profiled         *ProfiledConfig    // Bitmap text sampled through a floating-point row profile.
	RowColumn        *RowColumnConfig   // Fixed-advance text, row-source lookup and column displacement.
	RowBands         *RowBandsConfig    // Circular bitmap text with ordered destination-row passes.
	SizeBank         *SizeBankConfig    // Synchronized font-scale layers with one transport and cue clock.
	Ribbon           *RibbonConfig      // Fixed-tick horizontal or vertical atlas ribbon.
	Output           *OutputConfig      // Ordered image operations over the common text renderer.
	// RepeatBounds selects the visible pen coordinates before any mappers run.
	// Empty uses the destination bounds. Enlarge it for paths or projections that
	// bring distant pen positions into view. Only automatic repeat drawing uses it.
	RepeatBounds     image.Rectangle
	MaxGlyphsPerDraw int // Automatic drawing budget; zero defaults to 65536. Manual DrawAt stays explicit.
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
	cycleOrigin                          int  // Internal geometry rebasing for automatic repetition.
	clipGlyphs                           bool // Automatic Draw only; manual transport keeps its exact visit policy.
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
	config               Config
	glyphs               []Glyph
	length, duration     float64
	segments             []segment
	frame                kit.Frame
	draws                []glyphDraw
	repeatMin, repeatMax float64
	backend              kit.Effect
	output               kit.Effect
	drawErr              error
}

var ErrDrawBudget = errors.New("scrolling: automatic drawing exceeds its glyph budget")

// Err reports a resource-budget error from the last automatic Draw.
func (s *Scrolling) Err() error { return s.drawErr }

// Finished reports completion for finite transports such as Feed. Continuous
// transports return false, so a scene may use the same scrolling constructor.
func (s *Scrolling) Finished() bool {
	if b, ok := s.backend.(interface{ Finished() bool }); ok {
		return b.Finished()
	}
	return false
}

// Image exposes a transport's current visible surface when another effect must
// sample it. The surface is borrowed and can change after the next Update.
func (s *Scrolling) Image() *ebiten.Image {
	if b, ok := s.backend.(interface{ Image() *ebiten.Image }); ok {
		return b.Image()
	}
	return nil
}

// SetTransportMultiplier changes a compatible transport's speed without
// resetting its text, wave or column phases. It applies on the next Update.
func (s *Scrolling) SetTransportMultiplier(value float64) error {
	if b, ok := s.backend.(interface{ SetSpeedMultiplier(float64) error }); ok {
		return b.SetSpeedMultiplier(value)
	}
	return fmt.Errorf("scrolling: selected transport has no speed multiplier")
}

// Pseudo3DController exposes per-bank speed, position and time cues when the
// shared scrolling facade uses the pseudo-3D transport.
func (s *Scrolling) Pseudo3DController() *Pseudo3D {
	if p, ok := s.backend.(*Pseudo3D); ok {
		return p
	}
	return nil
}

// SizeBankController exposes the active scale and speed controls of a
// synchronized font bank selected through scrolling.New.
func (s *Scrolling) SizeBankController() *SizeBank {
	if bank, ok := s.backend.(*SizeBank); ok {
		return bank
	}
	return nil
}

// RibbonController exposes the current offset and editable speed of a
// fixed-tick ribbon selected through scrolling.New.
func (s *Scrolling) RibbonController() *Ribbon {
	if ribbon, ok := s.backend.(*Ribbon); ok {
		return ribbon
	}
	return nil
}

// ScanlineController exposes the text cursor and bounce of a proportional
// scanline transport selected through scrolling.New.
func (s *Scrolling) ScanlineController() *ScanlineScroll {
	if scanline, ok := s.backend.(*ScanlineScroll); ok {
		return scanline
	}
	return nil
}

// CursorRune returns the current character for transports that expose a text
// cursor. Other transport kinds return zero.
func (s *Scrolling) CursorRune() rune {
	if b, ok := s.backend.(interface{ CursorRune() rune }); ok {
		return b.CursorRune()
	}
	return 0
}

type glyphDraw struct {
	sample  Sample
	options ebiten.DrawImageOptions
	depth   float64
}

func New(c Config) (*Scrolling, error) {
	if c.MaxGlyphsPerDraw == 0 {
		c.MaxGlyphsPerDraw = 65536
	}
	if c.MaxGlyphsPerDraw < 1 || c.MaxGlyphsPerDraw > 1<<24 {
		return nil, fmt.Errorf("scrolling: invalid automatic draw budget")
	}
	if c.Recycled != nil || c.RingLanes != nil || c.Projected != nil || c.Pseudo3D != nil || c.Sliced != nil || c.Crawl != nil || c.Bands != nil || c.Slots != nil || c.Feed != nil || c.Scanline != nil || c.Profiled != nil || c.RowColumn != nil || c.RowBands != nil || c.SizeBank != nil || c.Ribbon != nil {
		return newTransport(c)
	}
	if c.Page != nil {
		if c.Glyphs != nil || !finite(c.Page.Width) || !finite(c.Page.LineHeight) || c.Page.Width < 0 || c.Page.LineHeight < 0 || c.Page.Align > AlignRight {
			return nil, fmt.Errorf("scrolling: invalid page layout")
		}
		c.Vertical = true
	}
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
		return s.finish()
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
	lineX, lineHeight, lineStart := 0.0, 0.0, 0
	finishLine := func() {
		if c.Page == nil {
			return
		}
		shift := 0.0
		if c.Page.Width > 0 {
			if c.Page.Align == AlignCenter {
				shift = (c.Page.Width - lineX) / 2
			}
			if c.Page.Align == AlignRight {
				shift = c.Page.Width - lineX
			}
		}
		for i := lineStart; i < len(s.glyphs); i++ {
			s.glyphs[i].X += shift
		}
		if lineHeight == 0 {
			if metrics := c.Fonts[faceName].Metrics; metrics != nil {
				scale := c.Fonts[faceName].ScaleY
				if scale == 0 {
					scale = 1
				}
				lineHeight = metrics.LineHeight() * sy * scale
			}
		}
		s.length += math.Max(lineHeight, c.Page.LineHeight)
		lineX, lineHeight, lineStart = 0, 0, len(s.glyphs)
	}
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
				if c.Page != nil && r == '\r' {
					continue
				}
				if c.Page != nil && r == '\n' {
					finishLine()
					continue
				}
				metric, _ := face.Metrics.Glyph(r)
				penAdvance := metric.Advance
				if c.Advance > 0 {
					penAdvance = c.Advance
				}
				advance := (penAdvance + spacing) * sx * fx
				if c.Vertical && c.Page == nil {
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
				if c.Page != nil {
					s.glyphs[len(s.glyphs)-1].X += lineX
					lineX += advance
					lineHeight = math.Max(lineHeight, face.Metrics.LineHeight()*sy*fy)
				} else {
					s.length += advance
				}
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
	if c.Page != nil && len(s.glyphs) > lineStart {
		finishLine()
	}
	s.addSegment(lastMarker, s.length+c.Gap, speed, shape)
	return s.finish()
}

func (s *Scrolling) prepareModes() error {
	s.repeatMax = s.length
	for _, g := range s.glyphs {
		position, size := g.X, 0.0
		if s.config.Vertical {
			position = g.Y
		}
		if g.Image != nil {
			size = float64(g.Image.Bounds().Dx()) * g.ScaleX
			if s.config.Vertical {
				size = float64(g.Image.Bounds().Dy()) * g.ScaleY
			}
		}
		s.repeatMin = math.Min(s.repeatMin, g.Offset+math.Min(position, position+size))
		s.repeatMax = math.Max(s.repeatMax, g.Offset+math.Max(position, position+size))
	}
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
func (s *Scrolling) Update(f kit.Frame) error {
	if s.drawErr != nil {
		return s.drawErr
	}
	s.frame = f
	if s.backend != nil {
		if err := s.backend.Update(f); err != nil {
			return err
		}
	}
	if s.output != nil {
		return s.output.Update(f)
	}
	return nil
}

// Draw repeats a continuous ribbon when Repeat is set, keeping both the tail
// and the next copy alive across control timeline boundaries. Manual DrawAt and
// StateAt retain their original single-pass semantics.
func (s *Scrolling) Draw(dst *ebiten.Image) {
	if s.output != nil {
		s.output.Draw(dst)
		return
	}
	s.drawCore(dst)
}

func (s *Scrolling) drawCore(dst *ebiten.Image) {
	s.drawErr = nil
	if s.backend != nil {
		s.backend.Draw(dst)
		return
	}
	state := s.StateAt(s.frame.Time)
	state.clipGlyphs = true
	if s.config.Repeat && len(s.glyphs) > 0 {
		state = s.repeatState(state, dst.Bounds())
	} else if len(s.glyphs) > s.config.MaxGlyphsPerDraw {
		s.drawErr = ErrDrawBudget
		return
	}
	s.DrawAt(dst, state)
}

func (s *Scrolling) repeatState(state DrawState, bounds image.Rectangle) DrawState {
	period := s.length + s.config.Gap
	if !s.config.RepeatBounds.Empty() {
		bounds = s.config.RepeatBounds
	}
	low, high, origin := float64(bounds.Min.X), float64(bounds.Max.X), state.X
	if s.config.Vertical {
		low, high, origin = float64(bounds.Min.Y), float64(bounds.Max.Y), state.Y
	}
	// Keep geometry near the viewport even after many hours, but retain absolute
	// glyph indices/offsets for stable sine, DNA and perspective phases.
	cycle := 0.0
	if s.duration > 0 && finite(s.duration) {
		t := math.Max(0, state.Time)
		// Use the same remainder as StateAt: division can round up at a cycle
		// boundary while Mod still returns a phase just below the duration.
		cycle = math.Round((t - math.Mod(t, s.duration)) / s.duration)
	}
	first := math.Ceil((low - origin - s.repeatMax) / period)
	last := math.Floor((high - origin - s.repeatMin) / period)
	mode := s.config.Modes[state.Shape]
	if s.config.Map != nil || len(s.config.Effects) > 0 || s.config.Shapes[state.Shape] != nil || mode.Map != nil || mode.Paint != nil {
		// Whole neighboring copies give ordinary deformations room to move. For
		// larger/custom projections, callers specify an explicit RepeatBounds.
		first--
		last++
	}
	// Do not introduce the end of a previous copy during the initial entry.
	first = math.Max(first, -cycle)
	state.Cycle, state.bounded = true, true
	state.First, state.End = 0, 0
	work := (last - first + 1) * float64(len(s.glyphs))
	if work > float64(s.config.MaxGlyphsPerDraw) || math.IsInf(work, 1) {
		s.drawErr = ErrDrawBudget
		return state
	}
	// Guard virtual-index conversion for invalid or unrepresentable clocks.
	limit := math.Min(float64(int(^uint(0)>>1)/len(s.glyphs))-2, 1<<52)
	if !finite(cycle) || !finite(first) || !finite(last) || last < first || cycle >= limit || math.Abs(first) >= limit || math.Abs(last) >= limit || cycle+last >= limit {
		return state
	}
	state.cycleOrigin = int(cycle)
	state.First = (state.cycleOrigin + int(first)) * len(s.glyphs)
	state.End = (state.cycleOrigin + int(last) + 1) * len(s.glyphs)
	if state.End-state.First > s.config.MaxGlyphsPerDraw {
		state.First, state.End = 0, 0
		s.drawErr = ErrDrawBudget
		return state
	}
	state.Position += cycle * period
	return state
}

// ValidateRenderBounds checks a conservative maximum automatic-draw workload
// before playback. Editors can reject impractically dense text layouts instead
// of waiting for a draw-time error. Explicit manual windows are unaffected.
func (s *Scrolling) ValidateRenderBounds(bounds image.Rectangle) error {
	if s.backend != nil || len(s.glyphs) == 0 {
		return nil
	}
	count := float64(len(s.glyphs))
	if s.config.Repeat {
		if !s.config.RepeatBounds.Empty() {
			bounds = s.config.RepeatBounds
		}
		extent := float64(bounds.Dx())
		if s.config.Vertical {
			extent = float64(bounds.Dy())
		}
		copies := math.Ceil((extent+s.repeatMax-s.repeatMin)/(s.length+s.config.Gap)) + 1
		if s.config.Map != nil || len(s.config.Effects) > 0 || len(s.config.Shapes) > 0 || len(s.config.Modes) > 0 {
			copies += 2
		}
		count *= copies
	}
	if !finite(count) || count > float64(s.config.MaxGlyphsPerDraw) {
		return ErrDrawBudget
	}
	return nil
}

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
		penOffset := g.Offset + float64(cycle-state.cycleOrigin)*(s.length+s.config.Gap)
		g.Offset += float64(cycle) * (s.length + s.config.Gap)
		x, y := state.X+g.X*state.ScaleX, state.Y+g.Y*state.ScaleY
		if s.config.Vertical {
			y += penOffset * state.ScaleY
		} else {
			x += penOffset * state.ScaleX
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
		// Cull after every mapper, so paths and projections can bring distant
		// pen positions into view. Custom painters may extend beyond the glyph
		// image and retain full control over their own clipping.
		if state.clipGlyphs && paint == nil && g.Image != nil && !glyphIntersects(dst.Bounds(), g.Image.Bounds(), op.GeoM) {
			return
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

func glyphIntersects(view, source image.Rectangle, transform ebiten.GeoM) bool {
	width, height := float64(source.Dx()), float64(source.Dy())
	minX, minY, maxX, maxY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, p := range [4][2]float64{{0, 0}, {width, 0}, {0, height}, {width, height}} {
		x, y := transform.Apply(p[0], p[1])
		if !finite(x) || !finite(y) {
			return false
		}
		minX, minY, maxX, maxY = math.Min(minX, x), math.Min(minY, y), math.Max(maxX, x), math.Max(maxY, y)
	}
	// Retain a one-pixel margin around fractional and filtered edges.
	return maxX > float64(view.Min.X)-1 && maxY > float64(view.Min.Y)-1 && minX < float64(view.Max.X)+1 && minY < float64(view.Max.Y)+1
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
