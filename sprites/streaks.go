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

package sprites

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type StreakConfig struct {
	Width, Height, Speed, Focal float64
	CenterX, CenterY            float64
	Color                       color.Color
	Count                       int
	Random                      func() float64
}
type streak struct {
	x, y, z, px, py, oldX, oldY, width float64
	visible                            bool
}

// Streaks preserves the endpoint history and wrap rules of a projected star
// field. Unlike a static point field, stars leave short depth-dependent lines.
type Streaks struct {
	config StreakConfig
	stars  []streak
	depth  float64
}

func NewStreaks(c StreakConfig) (*Streaks, error) {
	if !dimension(c.Width) || !dimension(c.Height) || !dimension(c.Focal) || c.Count < 0 || c.Count > 100000 || c.Random == nil {
		return nil, fmt.Errorf("sprites: invalid streak field")
	}
	if c.Color == nil {
		c.Color = color.White
	}
	s := &Streaks{config: c, stars: make([]streak, c.Count), depth: (c.Width + c.Height) / 2}
	for i := range s.stars {
		s.stars[i] = streak{x: c.Random()*c.Width*2 - c.Width, y: c.Random()*c.Height*2 - c.Height, z: math.Floor(c.Random()*s.depth + .5)}
	}
	return s, nil
}
func (s *Streaks) Step() {
	c := s.config
	dx, dy := float64(int(c.CenterX-c.Width/2)>>4), float64(int(c.CenterY-c.Height/2)>>4)
	for i := range s.stars {
		p := &s.stars[i]
		p.oldX, p.oldY = p.px, p.py
		p.visible = true
		p.x += dx
		if p.x > c.Width {
			p.x -= 2 * c.Width
			p.visible = false
		}
		if p.x < -c.Width {
			p.x += 2 * c.Width
			p.visible = false
		}
		p.y += dy
		if p.y > c.Height {
			p.y -= 2 * c.Height
			p.visible = false
		}
		if p.y < -c.Height {
			p.y += 2 * c.Height
			p.visible = false
		}
		p.z -= c.Speed
		if p.z > s.depth {
			p.z -= s.depth
			p.visible = false
		}
		if p.z < 0 {
			p.z += s.depth
			p.visible = false
		}
		if p.z == 0 {
			p.px, p.py = math.Inf(1), math.Inf(1)
			p.visible = false
			continue
		}
		p.px = c.Width/2 + p.x/p.z*c.Focal
		p.py = c.Height/2 + p.y/p.z*c.Focal
		p.width = (1 - p.z/s.depth) * 2
		p.visible = p.visible && p.oldX > 0 && p.oldX < c.Width && p.oldY > 0 && p.oldY < c.Height && p.width > 0
	}
}
func (s *Streaks) DrawAt(dst *ebiten.Image, x, y float64) {
	for _, p := range s.stars {
		if p.visible {
			vector.StrokeLine(dst, float32(p.oldX+x), float32(p.oldY+y), float32(p.px+x), float32(p.py+y), float32(p.width), s.config.Color, true)
		}
	}
}
