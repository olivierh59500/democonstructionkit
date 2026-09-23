package scrolling

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"math"
	"strings"
)

// RevealOrder determines when glyphs enter independently of their final layout.
type RevealOrder uint8

const (
	RevealRows RevealOrder = iota
	RevealColumns
	RevealColumnsBottomFirst
)

// RevealConfig describes an independently timed entrance for each text cell.
// FromOffset displaces the final glyph position; UniformStartY replaces its
// initial Y when nonnil. Time, Delay and Duration share any caller-selected unit.
type RevealConfig struct {
	Font                     BitmapGrid
	Lines                    []string
	Columns                  int
	X, Y, AdvanceX, AdvanceY float64
	OutputX, OutputY         float64
	FromX, FromY             float64
	UniformStartY            *float64
	Delay, Duration          float64
	Order                    RevealOrder
	Ease                     func(float64) float64
}
type revealGlyph struct {
	index        int
	x, y, startY float64
}
type Reveal struct {
	text   *BitmapText
	glyphs []revealGlyph
	config RevealConfig
}

func NewReveal(c RevealConfig) (*Reveal, error) {
	if c.Columns <= 0 || len(c.Lines) == 0 || c.Columns > 1<<20 || len(c.Lines) > 1<<20/c.Columns || c.Order > RevealColumnsBottomFirst || !finite(c.Delay) || c.Delay < 0 || !finite(c.Duration) || c.Duration <= 0 {
		return nil, fmt.Errorf("scrolling: invalid text reveal")
	}
	if c.AdvanceX == 0 {
		c.AdvanceX = c.Font.Width
	}
	if c.AdvanceY == 0 {
		c.AdvanceY = c.Font.Height
	}
	for _, v := range []float64{c.X, c.Y, c.AdvanceX, c.AdvanceY, c.FromX, c.FromY, c.OutputX, c.OutputY} {
		if !finite(v) {
			return nil, fmt.Errorf("scrolling: invalid reveal position")
		}
	}
	if c.UniformStartY != nil && !finite(*c.UniformStartY) {
		return nil, fmt.Errorf("scrolling: invalid reveal start")
	}
	var text strings.Builder
	rows := len(c.Lines)
	for _, line := range c.Lines {
		chars := []rune(line)
		for col := 0; col < c.Columns; col++ {
			ch := ' '
			if col < len(chars) {
				ch = chars[col]
			}
			text.WriteRune(ch)
		}
	}
	compiled, err := NewBitmapText(c.Font, text.String(), c.AdvanceX)
	if err != nil {
		return nil, err
	}
	r := &Reveal{text: compiled, config: c}
	for n := 0; n < rows*c.Columns; n++ {
		row, col := n/c.Columns, n%c.Columns
		if c.Order != RevealRows {
			col, row = n/rows, n%rows
			if c.Order == RevealColumnsBottomFirst {
				row = rows - 1 - row
			}
		}
		y := c.Y + float64(row)*c.AdvanceY
		startY := y + c.FromY
		if c.UniformStartY != nil {
			startY = *c.UniformStartY
		}
		r.glyphs = append(r.glyphs, revealGlyph{index: row*c.Columns + col, x: c.X + float64(col)*c.AdvanceX, y: y, startY: startY})
	}
	r.config.Lines = nil
	r.config.UniformStartY = nil
	return r, nil
}
func (r *Reveal) DrawAt(dst *ebiten.Image, time float64) {
	if dst == nil || !finite(time) {
		return
	}
	for n, g := range r.glyphs {
		p := math.Max(0, math.Min(1, (time-float64(n)*r.config.Delay)/r.config.Duration))
		if r.config.Ease != nil {
			p = r.config.Ease(p)
		}
		r.text.DrawGlyph(dst, g.index, g.x+r.config.FromX*(1-p)+r.config.OutputX, g.startY+(g.y-g.startY)*p+r.config.OutputY, 1, 1)
	}
}
