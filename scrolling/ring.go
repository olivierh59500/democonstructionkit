/*
Copyright (c) 2011 Antoine Santo Aka NoNameNo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package scrolling

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/olivierh59500/democonstructionkit/composite"
)

// BitmapGrid supports fractional atlas cells and explicit alphabets. Empty Order
// selects contiguous code points beginning at First. Pixel advances remain exact.
type BitmapGrid struct {
	Image         *ebiten.Image
	Width, Height float64
	Columns       int
	// ColumnSpan optionally preserves legacy imageWidth/cellWidth arithmetic,
	// including partial cells at a sheet's right edge. Zero uses Columns.
	ColumnSpan float64
	First      rune
	Order      string
	Filter     ebiten.Filter
}

func (g BitmapGrid) Region(ch rune) (composite.Region, bool) {
	i := int(ch - g.First)
	if g.Order != "" {
		i = -1
		for n, r := range []rune(g.Order) {
			if r == ch {
				i = n
				break
			}
		}
	}
	if i < 0 || g.Columns < 1 || g.Image == nil {
		return composite.Region{}, false
	}
	b := g.Image.Bounds()
	columns := float64(g.Columns)
	if g.ColumnSpan > 0 {
		columns = g.ColumnSpan
	}
	r := composite.Region{X: float64(b.Min.X) + math.Floor(math.Mod(float64(i), columns))*g.Width, Y: float64(b.Min.Y) + math.Floor(float64(i)/columns)*g.Height, Width: g.Width, Height: g.Height}
	return r, r.Valid() && r.X < float64(b.Max.X) && r.Y < float64(b.Max.Y)
}

// RingWave is a wave sampled in visible-letter order. Phase advances on recycled
// glyphs as well as ticks, retaining the original recycled-slot timing.
type RingWave struct {
	Phase, Amplitude, LetterStep, TickStep float64
	Floor                                  bool
}
type RingConfig struct {
	Text            string
	Font            BitmapGrid
	Viewport, Speed float64
	Waves           []RingWave
	Controls        bool
	Commands        map[string]func()
}
type RingLetter struct {
	X, Y float64
	Rune rune
}

// Ring is a legacy timing adapter for recycled-glyph scrollers. New compositions
// can use Scrolling with multiple Faces and named modes; this adapter preserves
// old slot initialization, fractional widths and command timing exactly.
type Ring struct {
	config          RingConfig
	letters         []RingLetter
	order           []int
	text            []rune
	next            int
	speed, oldSpeed float64
	pause, delay    int
	waves           []RingWave
}

func NewRing(c RingConfig) (*Ring, error) {
	if c.Font.Image == nil || c.Font.Columns < 1 || !finite(c.Font.Width) || !finite(c.Font.Height) || c.Font.Width <= 0 || c.Font.Height <= 0 || !finite(c.Viewport) || c.Viewport <= 0 || !finite(c.Speed) || c.Speed < 0 || c.Text == "" {
		return nil, fmt.Errorf("scrolling: invalid ring configuration")
	}
	wide := int(math.Ceil(c.Viewport/c.Font.Width)) + 1
	if wide > 16383 {
		return nil, fmt.Errorf("scrolling: ring viewport is too large")
	}
	r := &Ring{config: c, text: []rune(c.Text), speed: c.Speed, letters: make([]RingLetter, wide+1), order: make([]int, wide+1), waves: append([]RingWave(nil), c.Waves...)}
	for i := range r.letters {
		r.letters[i] = RingLetter{X: math.Ceil(float64(wide+i) * c.Font.Width), Rune: r.character(r.next)}
		r.next++
		r.order[i] = i
	}
	return r, nil
}

func (r *Ring) character(i int) rune {
	if i < 0 || i >= len(r.text) {
		return -1
	}
	return r.text[i]
}
func (r *Ring) Letters() []RingLetter { return append([]RingLetter(nil), r.letters...) }

// NextRune exposes the cursor for original productions whose embedded controls
// also drive scene choreography, independently of the built-in ^ commands.
func (r *Ring) NextRune() rune { return r.character(r.next) }

// Cursor is the next text index, exposed for original intro/loop transitions.
func (r *Ring) Cursor() int { return r.next }

// Print draws a static bitmap string with the same fractional atlas mapping.
func (g BitmapGrid) Print(dst *ebiten.Image, text string, x, y, scaleX, scaleY float64) {
	for i, ch := range []rune(text) {
		r, ok := g.Region(ch)
		if !ok {
			continue
		}
		op := ebiten.DrawImageOptions{Filter: g.Filter}
		op.GeoM.Scale(scaleX, scaleY)
		op.GeoM.Translate(x+float64(i)*g.Width*scaleX, y)
		composite.DrawRegion(dst, g.Image, r, &op)
	}
}

// Step advances exactly one source tick. Draw never changes speed or phase.
func (r *Ring) Step() {
	if r.speed == 0 {
		r.pause++
		if r.pause == 60*r.delay {
			r.speed = r.oldSpeed
		}
	}
	speed := r.speed
	span := float64(len(r.letters)) * r.config.Font.Width
	for i := range r.letters {
		p := &r.letters[i]
		p.X -= speed
		if p.X <= -r.config.Font.Width {
			if r.config.Controls && r.character(r.next) == '^' {
				switch r.character(r.next + 1) {
				case 'P':
					n, e := strconv.Atoi(string(r.character(r.next + 2)))
					if e == nil {
						r.delay = n
						r.pause = 0
						r.oldSpeed = r.speed
						r.speed = 0
						r.next += 3
					}
				case 'S':
					n, e := strconv.Atoi(string(r.character(r.next + 2)))
					if e == nil {
						r.speed = float64(n)
						r.next += 3
					}
				case 'C':
					end := r.next + 2
					for end < len(r.text) && r.text[end] != ';' {
						end++
					}
					if end < len(r.text) {
						name := string(r.text[r.next+2 : end])
						if fn := r.config.Commands[name]; fn != nil {
							fn()
						}
						r.next = end + 1
					}
				default:
					r.next++
				}
			} else {
				p.X += span
				p.Rune = r.character(r.next)
				r.next++
				if r.next >= len(r.text) {
					r.next = 0
				}
				for j := range r.waves {
					r.waves[j].Phase += r.waves[j].LetterStep
				}
			}
		}
	}
	sort.SliceStable(r.order, func(i, j int) bool { return r.letters[r.order[i]].X < r.letters[r.order[j]].X })
	for rank, index := range r.order {
		y := 0.0
		for _, w := range r.waves {
			v := math.Sin(w.Phase+float64(rank)*w.LetterStep) * w.Amplitude
			if w.Floor {
				v = math.Floor(v)
			}
			y += v
		}
		r.letters[index].Y = y
	}
	for i := range r.waves {
		r.waves[i].Phase += r.waves[i].TickStep
	}
}

func (r *Ring) DrawAt(dst *ebiten.Image, x, y float64) {
	for _, i := range r.order {
		p := r.letters[i]
		src, ok := r.config.Font.Region(p.Rune)
		if !ok {
			continue
		}
		op := ebiten.DrawImageOptions{Filter: r.config.Font.Filter}
		op.GeoM.Translate(x+p.X, y+p.Y)
		composite.DrawRegion(dst, r.config.Font.Image, src, &op)
	}
}

// HasCommands reports which callback names a caller needs to bind explicitly.
func RingCommands(text string) []string {
	var names []string
	for {
		start := strings.Index(text, "^C")
		if start < 0 {
			break
		}
		text = text[start+2:]
		end := strings.IndexByte(text, ';')
		if end < 0 {
			break
		}
		names = append(names, text[:end])
		text = text[end+1:]
	}
	return names
}
