package scrolling

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	kit "github.com/olivierh59500/democonstructionkit"
)

// FeedConfig describes a finite bitmap-text ribbon with a persistent trail.
// Each Update shifts the previous frame, inserts one glyph at the right edge,
// and reports completion after the last glyph has entered. The text is literal:
// unsupported characters retain no glyph advance, as in the supplied atlas.
type FeedConfig struct {
	Font                  *Atlas
	Text                  string
	Width, Height, Margin int
	InsertX               int // Initial glyph entry coordinate; zero defaults to Width.
	Scale                 float64
	Speed                 int
	UppercaseASCII        bool // Resolve ASCII lowercase through uppercase atlas entries.
	ProgressiveEntry      bool // Reveal the active glyph over successive updates at the viewport edge.
}

// Feed owns its two working surfaces. Image returns the active visible viewport
// for a shader or another image effect, and Draw copies it without advancing.
type Feed struct {
	config          FeedConfig
	glyphs          []*ebiten.Image
	advances        []int
	images          [2]*ebiten.Image
	views           [2]*ebiten.Image
	shift           [2]*ebiten.Image
	x, letter, tile int
	done            bool
	op              ebiten.DrawImageOptions
}

func NewFeed(c FeedConfig) (*Feed, error) {
	if c.InsertX == 0 {
		c.InsertX = c.Width
	}
	if c.Font == nil || c.Font.face.Atlas == nil || c.Text == "" || c.Width < 1 || c.Height < 1 || c.Margin < 0 || c.InsertX < 0 || c.InsertX > 8192 || c.Speed < 1 || c.Speed > c.Width || !finite(c.Scale) || c.Scale <= 0 || c.Width > 8192 || c.Margin > 8192-c.Width || c.Height > 8192 {
		return nil, fmt.Errorf("scrolling: invalid feed configuration")
	}
	f := &Feed{config: c, x: -1, letter: -1, tile: -1}
	for _, character := range c.Text {
		r := character
		if c.UppercaseASCII {
			if r >= 'a' && r <= 'z' {
				r = r - 'a' + 'A'
			}
		}
		image, advance := (*ebiten.Image)(nil), 0
		if glyphImage, glyph, ok := c.Font.ExactGlyph(r); ok {
			if !finite(glyph.Advance) || glyph.Advance > 1<<30 {
				return nil, fmt.Errorf("scrolling: feed glyph advance exceeds budget")
			}
			value := float64(int(glyph.Advance)) * c.Scale
			if !finite(value) || value > 1<<30 {
				return nil, fmt.Errorf("scrolling: feed glyph advance overflows")
			}
			image, advance = glyphImage, int(value)
		}
		f.glyphs = append(f.glyphs, image)
		f.advances = append(f.advances, advance)
	}
	for i := range f.images {
		f.images[i] = ebiten.NewImage(c.Width+c.Margin, c.Height)
		f.views[i] = f.images[i].SubImage(image.Rect(0, 0, c.Width, c.Height)).(*ebiten.Image)
		f.shift[i] = f.images[i].SubImage(image.Rect(c.Speed, 0, c.Width+c.Margin, c.Height)).(*ebiten.Image)
	}
	return f, nil
}

func (f *Feed) Update(kit.Frame) error {
	if f.done || f.images[0] == nil {
		return nil
	}
	if f.config.ProgressiveEntry {
		if f.tile < 0 {
			f.x, f.letter, f.tile = 0, 0, 0
		} else {
			for f.x <= -f.advances[f.tile] {
				f.x += f.advances[f.tile]
				f.letter++
				if f.letter >= len(f.glyphs) {
					f.done = true
					return nil
				}
				f.tile = f.letter
			}
		}
	} else if f.x < 0 {
		if f.tile > -1 {
			f.x += f.advances[f.tile]
		}
		f.letter++
		if f.letter >= len(f.glyphs) {
			f.done = true
			return nil
		}
		f.tile = f.letter
	}
	f.x -= f.config.Speed
	f.images[1].Clear()
	f.op.GeoM.Reset()
	f.op.ColorScale.Reset()
	f.images[1].DrawImage(f.shift[0], &f.op)
	if glyph := f.glyphs[f.tile]; glyph != nil {
		f.op.GeoM.Reset()
		f.op.GeoM.Scale(f.config.Scale, f.config.Scale)
		f.op.GeoM.Translate(float64(f.config.InsertX+f.x), 0)
		f.images[1].DrawImage(glyph, &f.op)
	}
	f.images[0], f.images[1] = f.images[1], f.images[0]
	f.views[0], f.views[1] = f.views[1], f.views[0]
	f.shift[0], f.shift[1] = f.shift[1], f.shift[0]
	return nil
}

func (f *Feed) Image() *ebiten.Image {
	if f == nil {
		return nil
	}
	return f.views[0]
}
func (f *Feed) Finished() bool { return f != nil && f.done }
func (f *Feed) Draw(dst *ebiten.Image) {
	if dst != nil && f.Image() != nil {
		dst.DrawImage(f.Image(), nil)
	}
}
func (f *Feed) Close() error {
	for i, img := range f.images {
		if img != nil {
			img.Deallocate()
			f.images[i], f.views[i], f.shift[i] = nil, nil, nil
		}
	}
	return nil
}
