package motion

import (
	"fmt"
	"math"
)

// GlyphPoint is a glyph center and depth. Renderers may use depth as scale.
type GlyphPoint struct{ X, Y, Z float64 }

// GlyphPose is the current character and position of one page cell.
type GlyphPose struct {
	Rune     rune
	Position GlyphPoint
}

// GlyphPagePhase identifies which side of the completion barrier is active.
type GlyphPagePhase uint8

const (
	GlyphPageEntering GlyphPagePhase = iota
	GlyphPageExiting
)

// GlyphPageConfig controls a reusable text-page choreography. Every coordinate,
// duration, ease and delay pattern can be varied independently of the font.
// All durations and the UpdateAt clock are milliseconds. DelayPatterns contain
// per-cell delay multiples; their values need not be unique. Pages are Unicode
// strings of exactly the same number of runes in a Columns-wide grid.
type GlyphPageConfig struct {
	Pages           []string
	DelayPatterns   [][]int
	Columns         int
	Origin, Center  GlyphPoint
	ColumnStep      float64
	RowStep         float64
	EnterDurationMS float64
	ExitDurationMS  float64
	DelayStepMS     float64
	EnterEase       Ease
	ExitEase        Ease
	InitialPattern  int
}

type glyphPageTween struct {
	from, to                     GlyphPoint
	startMS, durationMS, delayMS float64
	ease                         Ease
	active                       bool
}

// GlyphPageCycle owns the animation clock and stable depth order. It allocates
// during construction only; UpdateAt and DrawOrder allocate nothing.
type GlyphPageCycle struct {
	config        GlyphPageConfig
	pages         [][]rune
	patterns      [][]int
	poses         []GlyphPose
	enter, exit   []glyphPageTween
	order         []int
	page, pattern int
	phase         GlyphPagePhase
	completed     int
}

// NewGlyphPageCycle validates and copies authored pages and delay patterns.
func NewGlyphPageCycle(c GlyphPageConfig) (*GlyphPageCycle, error) {
	if c.Columns <= 0 || len(c.Pages) == 0 || len(c.DelayPatterns) == 0 ||
		c.InitialPattern < 0 || c.InitialPattern >= len(c.DelayPatterns) ||
		c.EnterDurationMS <= 0 || c.ExitDurationMS <= 0 || c.DelayStepMS < 0 {
		return nil, fmt.Errorf("motion: invalid glyph page dimensions or timing")
	}
	for _, v := range [...]float64{c.Origin.X, c.Origin.Y, c.Origin.Z, c.Center.X, c.Center.Y, c.Center.Z,
		c.ColumnStep, c.RowStep, c.EnterDurationMS, c.ExitDurationMS, c.DelayStepMS} {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("motion: nonfinite glyph page parameter")
		}
	}
	count := len([]rune(c.Pages[0]))
	if count == 0 || count > 1<<16 || count%c.Columns != 0 {
		return nil, fmt.Errorf("motion: invalid glyph page cell count")
	}
	g := &GlyphPageCycle{config: c, pages: make([][]rune, len(c.Pages)), patterns: make([][]int, len(c.DelayPatterns)),
		poses: make([]GlyphPose, count), enter: make([]glyphPageTween, count), exit: make([]glyphPageTween, count), order: make([]int, count)}
	for i, page := range c.Pages {
		g.pages[i] = []rune(page)
		if len(g.pages[i]) != count {
			return nil, fmt.Errorf("motion: glyph page %d has %d cells, want %d", i, len(g.pages[i]), count)
		}
	}
	for i, pattern := range c.DelayPatterns {
		if len(pattern) != count {
			return nil, fmt.Errorf("motion: delay pattern %d has %d cells, want %d", i, len(pattern), count)
		}
		g.patterns[i] = append([]int(nil), pattern...)
		for _, delay := range pattern {
			if delay < 0 || !finiteGlyphDelay(float64(delay)*c.DelayStepMS) {
				return nil, fmt.Errorf("motion: invalid delay in pattern %d", i)
			}
		}
	}
	if g.config.EnterEase == nil {
		g.config.EnterEase = Linear
	}
	if g.config.ExitEase == nil {
		g.config.ExitEase = Linear
	}
	for i := range g.order {
		g.order[i] = i
	}
	g.reset(0, c.InitialPattern, 0)
	return g, nil
}

func finiteGlyphDelay(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }

func (g *GlyphPageCycle) reset(page, pattern int, nowMS float64) {
	g.page, g.pattern, g.phase, g.completed = page, pattern, GlyphPageEntering, 0
	for i := range g.poses {
		g.poses[i] = GlyphPose{Rune: g.pages[page][i], Position: g.config.Center}
		target := GlyphPoint{X: g.config.Origin.X + float64(i%g.config.Columns)*g.config.ColumnStep,
			Y: g.config.Origin.Y + float64(i/g.config.Columns)*g.config.RowStep, Z: g.config.Origin.Z}
		g.enter[i] = glyphPageTween{from: g.config.Center, to: target, startMS: nowMS,
			durationMS: g.config.EnterDurationMS, delayMS: float64(g.patterns[pattern][i]) * g.config.DelayStepMS,
			ease: g.config.EnterEase, active: true}
		g.exit[i].active = false
	}
}

// SetPageAt starts a selected page and pattern at an explicit simulation time.
// This is suitable for a timeline cue or an interactive scene editor.
func (g *GlyphPageCycle) SetPageAt(page, pattern int, nowMS float64) error {
	if g == nil || page < 0 || page >= len(g.pages) || pattern < 0 || pattern >= len(g.patterns) || !finiteGlyphDelay(nowMS) {
		return fmt.Errorf("motion: invalid glyph page cue")
	}
	g.reset(page, pattern, nowMS)
	return nil
}

func (t *glyphPageTween) update(position *GlyphPoint, nowMS float64) bool {
	if !t.active {
		return false
	}
	elapsed := nowMS - t.startMS - t.delayMS
	if elapsed < 0 {
		return false
	}
	if elapsed >= t.durationMS {
		*position = t.to
		t.active = false
		return true
	}
	v := t.ease(elapsed / t.durationMS)
	position.X = t.from.X + (t.to.X-t.from.X)*v
	position.Y = t.from.Y + (t.to.Y-t.from.Y)*v
	position.Z = t.from.Z + (t.to.Z-t.from.Z)*v
	return false
}

// UpdateAt advances the choreography to nowMS. A whole-page barrier begins
// each outgoing or incoming wave at the exact update where its last glyph
// finishes, preserving the stagger even when a frame skips forward.
func (g *GlyphPageCycle) UpdateAt(nowMS float64) error {
	if g == nil || !finiteGlyphDelay(nowMS) {
		return fmt.Errorf("motion: invalid glyph page time")
	}
	for i := range g.poses {
		if g.enter[i].update(&g.poses[i].Position, nowMS) {
			g.completed++
			if g.completed == len(g.poses) {
				g.pattern = (g.pattern + 1) % len(g.patterns)
				g.phase, g.completed = GlyphPageExiting, 0
				for j := range g.poses {
					g.exit[j] = glyphPageTween{from: g.poses[j].Position, to: g.config.Center,
						startMS: nowMS, durationMS: g.config.ExitDurationMS,
						delayMS: float64(g.patterns[g.pattern][j]) * g.config.DelayStepMS,
						ease:    g.config.ExitEase, active: true}
				}
			}
		}
		if g.exit[i].update(&g.poses[i].Position, nowMS) {
			g.completed++
			if g.completed == len(g.poses) {
				g.reset((g.page+1)%len(g.pages), (g.pattern+1)%len(g.patterns), nowMS)
			}
		}
	}
	return nil
}

func (g *GlyphPageCycle) Count() int            { return len(g.poses) }
func (g *GlyphPageCycle) PageIndex() int        { return g.page }
func (g *GlyphPageCycle) PatternIndex() int     { return g.pattern }
func (g *GlyphPageCycle) Phase() GlyphPagePhase { return g.phase }
func (g *GlyphPageCycle) Glyph(i int) GlyphPose { return g.poses[i] }

// DrawOrder returns a borrowed stable depth sort. Call it during rendering;
// clients must not mutate the returned slice.
func (g *GlyphPageCycle) DrawOrder() []int {
	for i := 1; i < len(g.order); i++ {
		index := g.order[i]
		depth := g.poses[index].Position.Z
		j := i
		for j > 0 && g.poses[g.order[j-1]].Position.Z > depth {
			g.order[j] = g.order[j-1]
			j--
		}
		g.order[j] = index
	}
	return g.order
}
